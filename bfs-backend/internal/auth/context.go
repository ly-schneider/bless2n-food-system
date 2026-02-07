package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey      contextKey = "userID"
	UserRoleKey    contextKey = "userRole"
	UserEmailKey   contextKey = "userEmail"
	UserNameKey    contextKey = "userName"
	IsAnonymousKey contextKey = "isAnonymous"
	AuthTypeKey    contextKey = "authType"
	DeviceIDKey    contextKey = "deviceID"
	DeviceTypeKey  contextKey = "deviceType"
)

type AuthType string

const (
	AuthTypeUser   AuthType = "user"
	AuthTypeDevice AuthType = "device"
)

type DeviceType string

const (
	DeviceTypePOS     DeviceType = "pos"
	DeviceTypeStation DeviceType = "station"
)

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// RequireUserID panics if the user ID is not in context.
// Only safe after auth middleware has validated the request.
func RequireUserID(ctx context.Context) string {
	userID, ok := GetUserID(ctx)
	if !ok {
		panic("auth: user ID not found in context - ensure auth middleware is applied")
	}
	return userID
}

func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

// RequireUserRole panics if the user role is not in context.
// Only safe after auth middleware has validated the request.
func RequireUserRole(ctx context.Context) string {
	role, ok := GetUserRole(ctx)
	if !ok {
		panic("auth: user role not found in context - ensure auth middleware is applied")
	}
	return role
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRoleKey, role)
}

func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, UserEmailKey, email)
}

func GetUserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(UserNameKey).(string)
	return name, ok
}

func WithUserName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, UserNameKey, name)
}

func IsAnonymousUser(ctx context.Context) bool {
	isAnon, ok := ctx.Value(IsAnonymousKey).(bool)
	return ok && isAnon
}

func WithIsAnonymous(ctx context.Context, isAnonymous bool) context.Context {
	return context.WithValue(ctx, IsAnonymousKey, isAnonymous)
}

func GetAuthType(ctx context.Context) (AuthType, bool) {
	authType, ok := ctx.Value(AuthTypeKey).(AuthType)
	return authType, ok
}

func WithAuthType(ctx context.Context, authType AuthType) context.Context {
	return context.WithValue(ctx, AuthTypeKey, authType)
}

func GetDeviceID(ctx context.Context) (uuid.UUID, bool) {
	deviceID, ok := ctx.Value(DeviceIDKey).(uuid.UUID)
	return deviceID, ok
}

// RequireDeviceID panics if the device ID is not in context.
// Only safe after device auth middleware has validated the request.
func RequireDeviceID(ctx context.Context) uuid.UUID {
	deviceID, ok := GetDeviceID(ctx)
	if !ok {
		panic("auth: device ID not found in context - ensure device auth middleware is applied")
	}
	return deviceID
}

func WithDeviceID(ctx context.Context, deviceID uuid.UUID) context.Context {
	return context.WithValue(ctx, DeviceIDKey, deviceID)
}

func GetDeviceType(ctx context.Context) (DeviceType, bool) {
	deviceType, ok := ctx.Value(DeviceTypeKey).(DeviceType)
	return deviceType, ok
}

// RequireDeviceType panics if the device type is not in context.
// Only safe after device auth middleware has validated the request.
func RequireDeviceType(ctx context.Context) DeviceType {
	deviceType, ok := GetDeviceType(ctx)
	if !ok {
		panic("auth: device type not found in context - ensure device auth middleware is applied")
	}
	return deviceType
}

func WithDeviceType(ctx context.Context, deviceType DeviceType) context.Context {
	return context.WithValue(ctx, DeviceTypeKey, deviceType)
}

func IsDeviceAuth(ctx context.Context) bool {
	authType, ok := GetAuthType(ctx)
	return ok && authType == AuthTypeDevice
}

func IsUserAuth(ctx context.Context) bool {
	authType, ok := GetAuthType(ctx)
	return ok && authType == AuthTypeUser
}
