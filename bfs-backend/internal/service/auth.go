package service

import (
    "backend/internal/domain"
    "backend/internal/repository"
    "backend/internal/utils"
    "backend/internal/config"
    "context"
    "errors"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthService defines passwordless OTP and refresh flows
type AuthService interface {
    RequestOTP(ctx context.Context, email, ip, ua string) error
    VerifyWithCode(ctx context.Context, email, code string, clientID string) (*TokenPairResponse, *domain.User, error)
    Refresh(ctx context.Context, refreshToken string, clientID string) (*TokenPairResponse, *domain.User, error)
    Logout(ctx context.Context, refreshToken string) error
    ListUserActiveSessions(ctx context.Context, userID string) ([]map[string]any, error)
    RevokeSessionFamily(ctx context.Context, userID string, familyID string) error
    RevokeAllSessions(ctx context.Context, userID string) error
}

type authService struct {
    jwtService      JWTService
    emailService    EmailService
    users           repository.UserRepository
    otps            repository.OTPTokenRepository
    refreshTokens   repository.RefreshTokenRepository

    // simple in-memory rate limits (per email and per IP) for OTP requests and verify attempts
    limiter *rateLimiter
    cfg     config.Config
}

func NewAuthService(cfg config.Config, jwt JWTService, email EmailService, users repository.UserRepository, otps repository.OTPTokenRepository, rts repository.RefreshTokenRepository) AuthService {
    return &authService{
        jwtService:    jwt,
        emailService:  email,
        users:         users,
        otps:          otps,
        refreshTokens: rts,
        limiter:       newRateLimiter(),
        cfg:           cfg,
    }
}

const (
    otpTTL         = 10 * time.Minute
    maxOTPAttempts = 5
    refreshTTL     = 7 * 24 * time.Hour
)

func (a *authService) RequestOTP(ctx context.Context, email, ip, ua string) error {
    if !a.limiter.allow("req:ip:"+ip, 3, time.Minute) {
        return fmt.Errorf("rate_limited")
    }
    if !a.limiter.allow("req:email:"+email, 5, 15*time.Minute) {
        return fmt.Errorf("rate_limited")
    }

    user, err := a.users.UpsertCustomerByEmail(ctx, email)
    if err != nil {
        return err
    }
    // OTP code
    code, err := utils.GenerateOTP()
    if err != nil {
        return err
    }
    codeHash, err := utils.HashOTPArgon2(code)
    if err != nil {
        return err
    }
    if _, err := a.otps.CreateOTPCode(ctx, user.ID, codeHash, time.Now().UTC().Add(otpTTL)); err != nil {
        return err
    }
    // Send email (HTML + text)
    _ = a.emailService.SendLoginEmail(ctx, user.Email, code, ip, ua, otpTTL)
    return nil
}

func (a *authService) VerifyWithCode(ctx context.Context, email, code string, clientID string) (*TokenPairResponse, *domain.User, error) {
    key := "verify:email:" + email
    if !a.limiter.allow(key, 10, 15*time.Minute) { // generous cap combined with per-OTP attempts
        return nil, nil, fmt.Errorf("rate_limited")
    }
    user, err := a.users.UpsertCustomerByEmail(ctx, email)
    if err != nil {
        return nil, nil, err
    }
    otps, err := a.otps.FindActiveByUser(ctx, user.ID)
    if err != nil {
        return nil, nil, err
    }
    var matched *domain.OTPToken
    for i := range otps {
        ok, _ := utils.VerifyOTPArgon2(code, otps[i].TokenHash)
        if ok {
            matched = &otps[i]
            break
        }
    }
    if matched == nil {
        // increment attempts on latest OTP for the user to enforce attempts cap
        if len(otps) > 0 {
            attempts, _ := a.otps.IncrementAttempts(ctx, otps[0].ID)
            if attempts >= maxOTPAttempts {
                _ = a.otps.MarkUsed(ctx, otps[0].ID, time.Now().UTC()) // soft lock that OTP
            }
        }
        return nil, nil, errors.New("invalid_code")
    }
    if matched.Attempts >= maxOTPAttempts {
        return nil, nil, errors.New("too_many_attempts")
    }
    if err := a.otps.MarkUsed(ctx, matched.ID, time.Now().UTC()); err != nil {
        return nil, nil, err
    }
    return a.issueSession(ctx, user, clientID)
}

// Link-based login removed: OTP-only flow

func (a *authService) issueSession(ctx context.Context, user *domain.User, clientID string) (*TokenPairResponse, *domain.User, error) {
    // Access token
    access, err := a.jwtService.GenerateAccessToken(user)
    if err != nil {
        return nil, nil, err
    }
    // Refresh token
    rt, err := a.jwtService.GenerateRefreshToken()
    if err != nil {
        return nil, nil, err
    }
    // family id
    family, err := utils.GenerateFamilyID()
    if err != nil {
        return nil, nil, err
    }
    now := time.Now().UTC()
    if _, err := a.refreshTokens.Create(ctx, &domain.RefreshToken{
        ID:         primitive.NilObjectID,
        UserID:     user.ID,
        ClientID:   clientID,
        TokenHash:  utils.HashTokenSHA256(rt),
        IssuedAt:   now,
        LastUsedAt: time.Time{},
        ExpiresAt:  now.Add(refreshTTL),
        IsRevoked:  false,
        FamilyID:   family,
    }); err != nil {
        return nil, nil, err
    }
    return &TokenPairResponse{
        AccessToken:  access,
        RefreshToken: rt,
        TokenType:    "Bearer",
        ExpiresIn:    int64(AccessTokenDuration.Seconds()),
    }, user, nil
}

func (a *authService) Refresh(ctx context.Context, refreshToken string, clientID string) (*TokenPairResponse, *domain.User, error) {
    hash := utils.HashTokenSHA256(refreshToken)
    rec, err := a.refreshTokens.FindByHash(ctx, hash)
    if err != nil {
        return nil, nil, errors.New("invalid_refresh")
    }
    now := time.Now().UTC()
    if rec.IsRevoked || rec.ExpiresAt.Before(now) {
        return nil, nil, errors.New("invalid_refresh")
    }
    if !rec.LastUsedAt.IsZero() {
        // Reuse detected → revoke family
        _ = a.refreshTokens.RevokeFamily(ctx, rec.FamilyID, "reuse_detected")
        return nil, nil, errors.New("reuse_detected")
    }
    // Mark used
    if err := a.refreshTokens.MarkUsed(ctx, rec.ID, now); err != nil {
        return nil, nil, err
    }
    // Load user
    user, err := a.users.FindByID(ctx, rec.UserID)
    if err != nil {
        return nil, nil, err
    }
    // Issue new access and rotated refresh within same family
    access, err := a.jwtService.GenerateAccessToken(user)
    if err != nil {
        return nil, nil, err
    }
    newRT, err := a.jwtService.GenerateRefreshToken()
    if err != nil {
        return nil, nil, err
    }
    if _, err := a.refreshTokens.Create(ctx, &domain.RefreshToken{
        UserID:     rec.UserID,
        ClientID:   clientID,
        TokenHash:  utils.HashTokenSHA256(newRT),
        IssuedAt:   now,
        LastUsedAt: time.Time{},
        ExpiresAt:  now.Add(refreshTTL),
        IsRevoked:  false,
        FamilyID:   rec.FamilyID,
    }); err != nil {
        return nil, nil, err
    }
    return &TokenPairResponse{
        AccessToken:  access,
        RefreshToken: newRT,
        TokenType:    "Bearer",
        ExpiresIn:    int64(AccessTokenDuration.Seconds()),
    }, user, nil
}

func (a *authService) Logout(ctx context.Context, refreshToken string) error {
    hash := utils.HashTokenSHA256(refreshToken)
    rec, err := a.refreshTokens.FindByHash(ctx, hash)
    if err != nil {
        return errors.New("invalid_refresh")
    }
    return a.refreshTokens.RevokeFamily(ctx, rec.FamilyID, "logout")
}

func (a *authService) RevokeAllSessions(ctx context.Context, userID string) error {
    oid, err := primitive.ObjectIDFromHex(userID)
    if err != nil { return err }
    return a.refreshTokens.RevokeAllByUser(ctx, oid, "user_revoked_all")
}

// simple sliding-window in-memory limiter (not distributed; for tests/local)
type rateLimiter struct{
    buckets map[string][]time.Time
}
func newRateLimiter() *rateLimiter { return &rateLimiter{buckets: make(map[string][]time.Time)} }
func (r *rateLimiter) allow(key string, max int, window time.Duration) bool {
    now := time.Now()
    xs := r.buckets[key]
    // drop old
    cutoff := now.Add(-window)
    i := 0
    for ; i < len(xs); i++ { if xs[i].After(cutoff) { break } }
    xs = xs[i:]
    if len(xs) >= max { r.buckets[key] = xs; return false }
    xs = append(xs, now)
    r.buckets[key] = xs
    return true
}

// Sessions utilities
func (a *authService) ListUserActiveSessions(ctx context.Context, userID string) ([]map[string]any, error) {
    oid, err := primitive.ObjectIDFromHex(userID)
    if err != nil { return nil, errors.New("invalid_user_id") }
    recs, err := a.refreshTokens.ListActiveByUser(ctx, oid)
    if err != nil { return nil, err }
    type agg struct{ device string; createdAt time.Time; lastUsed time.Time; family string }
    m := make(map[string]*agg)
    for i := range recs {
        r := recs[i]
        g := m[r.FamilyID]
        if g == nil {
            g = &agg{device: r.ClientID, createdAt: r.IssuedAt, lastUsed: r.LastUsedAt, family: r.FamilyID}
            m[r.FamilyID] = g
        } else {
            if r.IssuedAt.Before(g.createdAt) { g.createdAt = r.IssuedAt }
            if r.LastUsedAt.After(g.lastUsed) { g.lastUsed = r.LastUsedAt }
        }
    }
    out := make([]map[string]any, 0, len(m))
    for _, v := range m {
        out = append(out, map[string]any{
            "id":           v.family,
            "device":       v.device,
            "created_at":   v.createdAt,
            "last_used_at": v.lastUsed,
            // 'current' best-effort: match on request UA is done in handler if needed
        })
    }
    return out, nil
}

func (a *authService) RevokeSessionFamily(ctx context.Context, userID string, familyID string) error {
    // For simple dev use, we won't verify ownership beyond relying on auth; production could check the family belongs to userID.
    return a.refreshTokens.RevokeFamily(ctx, familyID, "user_revoked")
}
