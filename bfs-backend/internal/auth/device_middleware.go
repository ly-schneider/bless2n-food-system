package auth

import (
	"context"
	"net/http"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/repository"
	"backend/internal/trace"

	"go.uber.org/zap"
)

// DeviceAuthMiddleware validates device bearer tokens by looking up the
// SHA-256 hash in the device_binding table, then validating the underlying
// Better Auth session.
type DeviceAuthMiddleware struct {
	bindingRepo repository.DeviceBindingRepository
	sessionRepo repository.SessionRepository
	logger      *zap.Logger
}

func NewDeviceAuthMiddleware(
	bindingRepo repository.DeviceBindingRepository,
	sessionRepo repository.SessionRepository,
	logger *zap.Logger,
) *DeviceAuthMiddleware {
	return &DeviceAuthMiddleware{
		bindingRepo: bindingRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// RequireDevice returns middleware that requires a valid device bearer token
// of the specified type. It:
// 1. Extracts the bearer token
// 2. Hashes it and looks up the device binding
// 3. Validates the underlying Better Auth session is not expired
// 4. Performs sliding session refresh
// 5. Sets device context (type, ID) and user context from the session
func (m *DeviceAuthMiddleware) RequireDevice(deviceType DeviceType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spanCtx, finish := trace.StartSpan(r.Context(), "auth", "device.require")
			trace.Data(spanCtx, "device.expected_type", string(deviceType))

			tokenString, err := ExtractBearerToken(r)
			if err != nil {
				trace.Err(spanCtx, err)
				finish()
				w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
				http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
				return
			}

			tokenHash := repository.HashToken(tokenString)
			binding, err := m.bindingRepo.GetByTokenHash(spanCtx, tokenHash)
			if err != nil {
				trace.Err(spanCtx, err)
				finish()
				m.logger.Debug("device binding not found", zap.Error(err))
				w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
				http.Error(w, "Unauthorized: invalid device token", http.StatusUnauthorized)
				return
			}

			expectedType := deviceTypeToDBValue(deviceType)
			if string(binding.DeviceType) != expectedType {
				trace.Tag(spanCtx, "error", "true")
				trace.Data(spanCtx, "device.actual_type", string(binding.DeviceType))
				finish()
				m.logger.Debug("device type mismatch",
					zap.String("expected", expectedType),
					zap.String("actual", string(binding.DeviceType)),
				)
				http.Error(w, "Forbidden: wrong device type", http.StatusForbidden)
				return
			}

			session, err := m.sessionRepo.GetByToken(spanCtx, tokenString)
			if err != nil {
				trace.Err(spanCtx, err)
				finish()
				m.logger.Debug("device session invalid or expired",
					zap.String("deviceID", binding.ID.String()),
					zap.Error(err),
				)
				w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
				http.Error(w, "Unauthorized: device session expired", http.StatusUnauthorized)
				return
			}

			if session.UpdatedAt.Add(sessionUpdateAge).Before(time.Now().UTC()) {
				if refreshErr := m.sessionRepo.RefreshSession(spanCtx, tokenString, sessionExpiresIn); refreshErr != nil {
					m.logger.Warn("failed to refresh device session", zap.Error(refreshErr))
				}
			}

			trace.Data(spanCtx, "device.id", binding.ID.String())
			trace.Data(spanCtx, "user.id", session.UserID)
			finish()

			go func(bindingID ent.DeviceBinding) {
				bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := m.bindingRepo.UpdateLastSeen(bgCtx, bindingID.ID); err != nil {
					m.logger.Warn("failed to update device last_seen", zap.Error(err))
				}
			}(*binding)

			ctx := r.Context()
			ctx = WithAuthType(ctx, AuthTypeDevice)
			if binding.DeviceID != nil {
				ctx = WithDeviceID(ctx, *binding.DeviceID)
			} else {
				ctx = WithDeviceID(ctx, binding.ID)
			}
			ctx = WithDeviceType(ctx, deviceType)
			ctx = WithUserID(ctx, session.UserID)
			if session.Role != "" {
				ctx = WithUserRole(ctx, string(session.Role))
			}

			trace.IdentifyUser(ctx, session.UserID, "", "")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// deviceTypeToDBValue converts the context DeviceType (lowercase) to the DB enum value (uppercase).
func deviceTypeToDBValue(dt DeviceType) string {
	switch dt {
	case DeviceTypePOS:
		return "POS"
	case DeviceTypeStation:
		return "STATION"
	default:
		return string(dt)
	}
}
