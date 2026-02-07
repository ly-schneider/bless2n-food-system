package observability

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

// Init wires the OTLP exporter, tracer provider, and global propagators.
// It returns a shutdown function that must be called during application exit.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(
		ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(semconv.ServiceNameKey.String(cfg.ServiceName)),
		resource.WithAttributes(cfg.ResourceAttributes...),
	)
	if err != nil {
		return nil, fmt.Errorf("build resource: %w", err)
	}

	clientOpts := buildHTTPClientOptions(cfg.ExporterEndpoint)
	exp, err := otlptrace.New(ctx, otlptracehttp.NewClient(clientOpts...))
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(samplerFromConfig(cfg.TracesSampler, cfg.TracesSamplerArg)),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exp)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	shutdown := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}

	return shutdown, nil
}

func samplerFromConfig(name, arg string) sdktrace.Sampler {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "parentbased_always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	case "traceidratio", "parentbased_traceidratio":
		ratio := parseRatio(arg)
		base := sdktrace.TraceIDRatioBased(ratio)
		if strings.HasPrefix(name, "parentbased") {
			return sdktrace.ParentBased(base)
		}
		return base
	default:
		return sdktrace.AlwaysSample()
	}
}

func parseRatio(val string) float64 {
	if val == "" {
		return 1.0
	}
	ratio, err := strconv.ParseFloat(val, 64)
	if err != nil || ratio < 0 || ratio > 1 {
		return 1.0
	}
	return ratio
}

func buildHTTPClientOptions(rawEndpoint string) []otlptracehttp.Option {
	endpoint, insecure, path := normalizeEndpoint(rawEndpoint)
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}
	if path != "" && path != "/" {
		opts = append(opts, otlptracehttp.WithURLPath(path))
	}
	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	return opts
}

func normalizeEndpoint(raw string) (endpoint string, insecure bool, path string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = defaultExporterEndpoint
	}

	if !strings.Contains(raw, "://") {
		// Assume host:port without scheme; default to http for local collector.
		return raw, true, ""
	}

	u, err := url.Parse(raw)
	if err != nil {
		return raw, strings.HasPrefix(raw, "http://"), ""
	}

	insecure = u.Scheme == "http"
	endpoint = u.Host
	path = strings.TrimSuffix(u.Path, "/")

	return endpoint, insecure, path
}
