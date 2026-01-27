package auth

import (
	"context"
)

type contextKey string

const (
	// UserIDKey is the context key for the user ID (sub claim from JWT).
	UserIDKey contextKey = "userID"
	// UserRoleKey is the context key for the user role.
	UserRoleKey contextKey = "userRole"
)

// GetUserID retrieves the user ID from the context.
// Returns the user ID and true if found, empty string and false otherwise.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// RequireUserID retrieves the user ID from the context.
// Panics if the user ID is not found (use only when auth middleware has already validated).
func RequireUserID(ctx context.Context) string {
	userID, ok := GetUserID(ctx)
	if !ok {
		panic("auth: user ID not found in context - ensure auth middleware is applied")
	}
	return userID
}

// GetUserRole retrieves the user role from the context.
// Returns the role and true if found, empty string and false otherwise.
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

// RequireUserRole retrieves the user role from the context.
// Panics if the role is not found (use only when auth middleware has already validated).
func RequireUserRole(ctx context.Context) string {
	role, ok := GetUserRole(ctx)
	if !ok {
		panic("auth: user role not found in context - ensure auth middleware is applied")
	}
	return role
}

// WithUserID returns a new context with the user ID set.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithUserRole returns a new context with the user role set.
func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRoleKey, role)
}
