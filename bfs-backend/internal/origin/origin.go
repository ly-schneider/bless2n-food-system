package origin

import "context"

type contextKey struct{}

// Info carries resolved frontend/backend origins for a request.
type Info struct {
	Frontend string
	Backend  string
}

var key contextKey

// WithContext stores origin info in the context for downstream consumers.
func WithContext(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, key, info)
}

// FromContext retrieves origin info from the context.
func FromContext(ctx context.Context) (Info, bool) {
	if ctx == nil {
		return Info{}, false
	}
	if v, ok := ctx.Value(key).(Info); ok {
		return v, true
	}
	return Info{}, false
}
