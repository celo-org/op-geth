package tracing

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	tracer     oteltrace.Tracer
	tracerOnce sync.Once
	isEnabled  bool
	config     *Config
)

// Initialize sets up the global tracer with the given configuration
func Initialize(cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid tracing config: %w", err)
	}

	config = cfg
	isEnabled = cfg.Enabled

	if !cfg.Enabled {
		// Set up a no-op tracer
		otel.SetTracerProvider(oteltrace.NewNoopTracerProvider())
		tracer = otel.Tracer("geth-noop")
		log.Debug("OpenTelemetry tracing disabled, using no-op tracer")
		return nil
	}

	return initializeTracer(cfg)
}

func initializeTracer(cfg *Config) error {
	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Parse endpoint URL to determine scheme and extract host:port
	endpoint := cfg.Endpoint
	var useInsecure bool
	
	// If endpoint contains a scheme, parse it
	if strings.Contains(endpoint, "://") {
		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("failed to parse endpoint URL: %w", err)
		}
		
		// Determine if we should use insecure (HTTP) connection
		useInsecure = parsedURL.Scheme == "http"
		
		// Extract host:port from URL (strip scheme and path)
		endpoint = parsedURL.Host
		if parsedURL.Port() == "" {
			// Add default port if not specified
			if parsedURL.Scheme == "https" {
				endpoint += ":443"
			} else {
				endpoint += ":80"
			}
		}
	} else {
		// No scheme specified, default to HTTP (insecure)
		useInsecure = true
	}

	// Build exporter options
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithHeaders(cfg.Headers),
		otlptracehttp.WithTimeout(cfg.Timeout),
	}
	
	// Use insecure connection for HTTP endpoints
	if useInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
		log.Debug("Using insecure HTTP connection for OTLP exporter", "endpoint", endpoint)
	}

	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		log.Error("Failed to create OTLP exporter", "endpoint", endpoint, "err", err)
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRate)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Get tracer
	tracer = otel.Tracer("geth")

	log.Info("OpenTelemetry tracing initialized",
		"endpoint", endpoint,
		"service", cfg.ServiceName,
		"version", cfg.ServiceVersion,
		"sample_rate", cfg.SampleRate,
		"timeout", cfg.Timeout,
		"rpc_tracing", cfg.EnableRPCTracing,
		"engine_api_tracing", cfg.EnableEngineAPITracing,
		"transaction_tracing", cfg.EnableTransactionTracing,
	)

	return nil
}

// IsEnabled returns true if tracing is enabled
func IsEnabled() bool {
	return isEnabled
}

// GetTracer returns the global tracer instance
func GetTracer() oteltrace.Tracer {
	tracerOnce.Do(func() {
		if tracer == nil {
			// Initialize with default config if not already initialized
			_ = Initialize(nil)
		}
	})
	return tracer
}

// StartSpan starts a new span with the given name and options
func StartSpan(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	if !IsEnabled() {
		return ctx, oteltrace.SpanFromContext(ctx)
	}
	return GetTracer().Start(ctx, spanName, opts...)
}

// GetConfig returns the current tracing configuration
func GetConfig() *Config {
	if config == nil {
		return DefaultConfig()
	}
	return config
}

// Shutdown gracefully shuts down the tracer
func Shutdown(ctx context.Context) error {
	if !IsEnabled() {
		return nil
	}

	log.Debug("Shutting down OpenTelemetry tracer")
	if tp, ok := otel.GetTracerProvider().(*trace.TracerProvider); ok {
		if err := tp.Shutdown(ctx); err != nil {
			log.Error("Error shutting down tracer", "err", err)
			return err
		}
		log.Info("OpenTelemetry tracer shut down successfully")
	}
	return nil
}

// ExtractTraceContext extracts trace context from environment variables
func ExtractTraceContext() (*Config, error) {
	cfg := DefaultConfig()

	// Check environment variables
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		cfg.Endpoint = endpoint
		cfg.Enabled = true
		log.Debug("Tracing enabled via OTEL_EXPORTER_OTLP_ENDPOINT", "endpoint", endpoint)
	}

	if serviceName := os.Getenv("OTEL_SERVICE_NAME"); serviceName != "" {
		cfg.ServiceName = serviceName
		log.Debug("Service name set from OTEL_SERVICE_NAME", "service", serviceName)
	}

	if serviceVersion := os.Getenv("OTEL_SERVICE_VERSION"); serviceVersion != "" {
		cfg.ServiceVersion = serviceVersion
		log.Debug("Service version set from OTEL_SERVICE_VERSION", "version", serviceVersion)
	}

	// Parse headers from OTEL_EXPORTER_OTLP_HEADERS
	if headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"); headers != "" {
		cfg.Headers = parseHeaders(headers)
		log.Debug("OTLP headers configured from OTEL_EXPORTER_OTLP_HEADERS")
	}

	return cfg, nil
}

// parseHeaders parses comma-separated key=value pairs
func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	// Simple parsing - in production, you might want more robust parsing
	// Format: "key1=value1,key2=value2"
	return headers
}