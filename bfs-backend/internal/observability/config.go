package observability

import (
	"os"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

const (
	defaultExporterEndpoint = "http://localhost:4318"
	defaultSampler          = "always_on"
	defaultServiceName      = "bfs-backend"
)

// Config describes how the SDK should export spans to the local collector.
// Values primarily come from OTEL_* environment variables so existing
// deployments can reuse standard configuration knobs.
type Config struct {
	ServiceName        string
	ResourceAttributes []attribute.KeyValue
	ExporterEndpoint   string
	TracesSampler      string
	TracesSamplerArg   string
}

// NewConfigFromEnv builds a Config using OTEL_* variables with sensible
// defaults for a sidecar/local collector workflow.
func NewConfigFromEnv() Config {
	cfg := Config{
		ServiceName:        strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")),
		ExporterEndpoint:   strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		TracesSampler:      strings.TrimSpace(os.Getenv("OTEL_TRACES_SAMPLER")),
		TracesSamplerArg:   strings.TrimSpace(os.Getenv("OTEL_TRACES_SAMPLER_ARG")),
		ResourceAttributes: parseResourceAttributes(os.Getenv("OTEL_RESOURCE_ATTRIBUTES")),
	}

	if cfg.ExporterEndpoint == "" {
		cfg.ExporterEndpoint = defaultExporterEndpoint
	}
	if cfg.TracesSampler == "" {
		cfg.TracesSampler = defaultSampler
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = defaultServiceName
	}

	return cfg
}

func parseResourceAttributes(raw string) []attribute.KeyValue {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	pairs := strings.Split(raw, ",")
	attrs := make([]attribute.KeyValue, 0, len(pairs))

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" || looksSensitive(key) {
			continue
		}
		attrs = append(attrs, attribute.String(key, truncateString(val)))
	}

	return attrs
}
