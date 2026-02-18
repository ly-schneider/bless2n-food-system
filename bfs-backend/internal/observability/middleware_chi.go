package observability

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// ChiMiddleware creates a server span for every incoming request, capturing
// route patterns and status codes so the collector can tail-sample accurately.
func ChiMiddleware(tracer trace.Tracer) func(http.Handler) http.Handler {
	if tracer == nil {
		tracer = otel.Tracer("http-server")
	}
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			ctx = context.WithValue(ctx, rootSpanContextKey{}, span)
			rec := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r.WithContext(ctx))

			routePattern := routePatternFromContext(ctx)
			if routePattern != "" {
				span.SetName(fmt.Sprintf("%s %s", r.Method, routePattern))
			} else {
				routePattern = r.URL.Path
			}

			statusCode := rec.status
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			attrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.URLScheme(requestScheme(r)),
				semconv.HTTPResponseStatusCode(statusCode),
				semconv.URLPath(r.URL.Path),
			}
			if routePattern != "" {
				attrs = append(attrs, semconv.HTTPRouteKey.String(routePattern))
			}
			if host := r.Host; host != "" {
				attrs = append(attrs, semconv.ServerAddressKey.String(host))
			}
			if rid := chiMw.GetReqID(ctx); rid != "" {
				attrs = append(attrs, attribute.String("http.request_id", rid))
			}

			span.SetAttributes(attrs...)

			if statusCode >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(statusCode))
			} else {
				span.SetStatus(codes.Unset, "")
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func routePatternFromContext(ctx context.Context) string {
	if rc := chi.RouteContext(ctx); rc != nil {
		return rc.RoutePattern()
	}
	return ""
}

func requestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		return forwarded
	}
	return "http"
}
