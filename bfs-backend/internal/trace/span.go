package trace

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
)

func StartSpan(ctx context.Context, op, description string) (context.Context, func()) {
	parent := sentry.SpanFromContext(ctx)
	if parent == nil {
		return ctx, func() {}
	}
	span := parent.StartChild(op)
	span.Description = description
	return span.Context(), span.Finish
}

func IdentifyUser(ctx context.Context, id, email, name string) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		return
	}
	hub.Scope().SetUser(sentry.User{
		ID:       id,
		Email:    email,
		Username: name,
	})
}

func Tag(ctx context.Context, key, value string) {
	span := sentry.SpanFromContext(ctx)
	if span == nil {
		return
	}
	span.SetTag(key, value)
}

func Data(ctx context.Context, key string, value any) {
	span := sentry.SpanFromContext(ctx)
	if span == nil {
		return
	}
	span.SetData(key, value)
}

func Breadcrumb(ctx context.Context, category, message string, data map[string]any) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		return
	}
	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category:  category,
		Message:   message,
		Level:     sentry.LevelInfo,
		Data:      data,
		Timestamp: time.Now(),
	}, nil)
}

func Err(ctx context.Context, err error) {
	if err == nil {
		return
	}
	Tag(ctx, "error", "true")
	Data(ctx, "error.message", err.Error())
}
