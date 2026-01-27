package observability

import (
	"context"
	"strings"
	"unicode"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const maxAttributeValueLen = 256

type rootSpanContextKey struct{}

// SpanContext bundles the current span and the request root span (if any).
type SpanContext struct {
	Span trace.Span
	Root trace.Span
}

// From returns the current and root spans (if present) for convenience.
func From(ctx context.Context) SpanContext {
	return SpanContext{
		Span: trace.SpanFromContext(ctx),
		Root: rootSpanFromContext(ctx),
	}
}

// SetAttrs attaches attributes to the current span.
func SetAttrs(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !isUsable(span) {
		return
	}
	span.SetAttributes(sanitizeAttributes(attrs)...)
}

// SetRootAttrs ensures attributes land on the root span (useful for tail sampling).
func SetRootAttrs(ctx context.Context, attrs ...attribute.KeyValue) {
	root := rootSpanFromContext(ctx)
	if !isUsable(root) {
		root = trace.SpanFromContext(ctx)
	}
	if !isUsable(root) {
		return
	}
	root.SetAttributes(sanitizeAttributes(attrs)...)
}

// Event records a structured event on the current span.
func Event(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if name == "" {
		name = "event"
	}
	span := trace.SpanFromContext(ctx)
	if !isUsable(span) {
		return
	}
	span.AddEvent(name, trace.WithAttributes(sanitizeAttributes(attrs)...))
}

// Decision records a decision event and mirrors key outcomes onto the root span.
func Decision(ctx context.Context, name, outcome, reason string, attrs ...attribute.KeyValue) {
	if name == "" {
		name = "decision"
	}
	eventAttrs := []attribute.KeyValue{
		attribute.String("event.kind", "decision"),
		attribute.String("decision.name", truncateString(name)),
		attribute.String("decision.outcome", truncateString(outcome)),
	}
	if reason != "" {
		eventAttrs = append(eventAttrs, attribute.String("decision.reason", truncateString(reason)))
	}
	eventAttrs = append(eventAttrs, sanitizeAttributes(attrs)...)

	span := trace.SpanFromContext(ctx)
	if isUsable(span) {
		span.AddEvent("decision:"+truncateString(name), trace.WithAttributes(eventAttrs...))
	}

	root := rootSpanFromContext(ctx)
	if !isUsable(root) {
		root = span
	}
	if !isUsable(root) {
		return
	}

	keyBase := sanitizeKeyForAttr(name)
	summary := []attribute.KeyValue{
		attribute.String("decision."+keyBase+".outcome", truncateString(outcome)),
	}
	if reason != "" {
		summary = append(summary, attribute.String("decision."+keyBase+".reason", truncateString(reason)))
	}
	root.SetAttributes(summary...)
}

func rootSpanFromContext(ctx context.Context) trace.Span {
	if v := ctx.Value(rootSpanContextKey{}); v != nil {
		if span, ok := v.(trace.Span); ok {
			return span
		}
	}
	return nil
}

func isUsable(span trace.Span) bool {
	if span == nil {
		return false
	}
	sc := span.SpanContext()
	return sc.IsValid() && span.IsRecording()
}

func sanitizeAttributes(attrs []attribute.KeyValue) []attribute.KeyValue {
	out := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if cleaned, ok := sanitizeAttribute(attr); ok {
			out = append(out, cleaned)
		}
	}
	return out
}

func sanitizeAttribute(attr attribute.KeyValue) (attribute.KeyValue, bool) {
	if attr.Key == "" || looksSensitive(string(attr.Key)) {
		return attribute.KeyValue{}, false
	}

	switch attr.Value.Type() {
	case attribute.STRING:
		return attr.Key.String(truncateString(attr.Value.AsString())), true
	case attribute.BOOL:
		return attr.Key.Bool(attr.Value.AsBool()), true
	case attribute.INT64:
		return attr.Key.Int64(attr.Value.AsInt64()), true
	case attribute.FLOAT64:
		return attr.Key.Float64(attr.Value.AsFloat64()), true
	case attribute.STRINGSLICE:
		vals := attr.Value.AsStringSlice()
		for i := range vals {
			vals[i] = truncateString(vals[i])
		}
		return attr.Key.StringSlice(vals), true
	case attribute.INT64SLICE:
		return attr, true
	case attribute.BOOLSLICE:
		return attr, true
	case attribute.FLOAT64SLICE:
		return attr, true
	default:
		return attr, true
	}
}

func looksSensitive(key string) bool {
	key = strings.ToLower(key)
	for _, marker := range []string{"token", "secret", "password", "authorization", "cookie", "apikey"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func truncateString(s string) string {
	if len(s) <= maxAttributeValueLen {
		return s
	}
	return s[:maxAttributeValueLen]
}

func sanitizeKeyForAttr(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "decision"
	}

	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		if r == '_' || r == '-' || r == '.' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('_')
	}

	out := strings.Trim(b.String(), "._-")
	if out == "" {
		return "decision"
	}
	if len(out) > 64 {
		return out[:64]
	}
	return out
}
