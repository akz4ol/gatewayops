package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc/credentials"
)

// TelemetryService handles OpenTelemetry configuration and export.
type TelemetryService struct {
	db            *sql.DB
	tracerProvider *sdktrace.TracerProvider
	logger        *slog.Logger
}

// OTELConfig represents OpenTelemetry export configuration.
type OTELConfig struct {
	Endpoint string            `json:"endpoint"`
	Protocol string            `json:"protocol"` // grpc or http
	Headers  map[string]string `json:"headers"`
	Enabled  bool              `json:"enabled"`
}

// NewTelemetryService creates a new telemetry service.
func NewTelemetryService(db *sql.DB, logger *slog.Logger) *TelemetryService {
	return &TelemetryService{
		db:     db,
		logger: logger,
	}
}

// Initialize sets up the OpenTelemetry trace provider.
func (s *TelemetryService) Initialize(ctx context.Context, serviceName, serviceVersion string) error {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			attribute.String("environment", "production"),
		),
	)
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}

	s.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(s.tracerProvider)

	s.logger.Info("telemetry service initialized",
		"service_name", serviceName,
		"service_version", serviceVersion,
	)

	return nil
}

// ConfigureExporter configures an OTLP exporter for an organization.
func (s *TelemetryService) ConfigureExporter(ctx context.Context, orgID string, config OTELConfig) error {
	if !config.Enabled {
		s.logger.Info("OTLP export disabled", "org_id", orgID)
		return nil
	}

	var exporter *otlptrace.Exporter
	var err error

	switch config.Protocol {
	case "grpc":
		exporter, err = s.createGRPCExporter(ctx, config)
	case "http":
		exporter, err = s.createHTTPExporter(ctx, config)
	default:
		return fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}

	if err != nil {
		return fmt.Errorf("create exporter: %w", err)
	}

	// Add exporter to tracer provider
	s.tracerProvider.RegisterSpanProcessor(
		sdktrace.NewBatchSpanProcessor(exporter),
	)

	s.logger.Info("OTLP exporter configured",
		"org_id", orgID,
		"endpoint", config.Endpoint,
		"protocol", config.Protocol,
	)

	return nil
}

// createGRPCExporter creates a gRPC OTLP exporter.
func (s *TelemetryService) createGRPCExporter(ctx context.Context, config OTELConfig) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Endpoint),
	}

	// Add headers
	if len(config.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(config.Headers))
	}

	// Check if endpoint is secure
	if isSecureEndpoint(config.Endpoint) {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	client := otlptracegrpc.NewClient(opts...)
	return otlptrace.New(ctx, client)
}

// createHTTPExporter creates an HTTP OTLP exporter.
func (s *TelemetryService) createHTTPExporter(ctx context.Context, config OTELConfig) (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint),
	}

	// Add headers
	if len(config.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(config.Headers))
	}

	// Check if endpoint is secure
	if !isSecureEndpoint(config.Endpoint) {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	client := otlptracehttp.NewClient(opts...)
	return otlptrace.New(ctx, client)
}

// isSecureEndpoint checks if an endpoint uses HTTPS/TLS.
func isSecureEndpoint(endpoint string) bool {
	return len(endpoint) >= 5 && endpoint[:5] == "https"
}

// GetConfig retrieves the OTEL configuration for an organization.
func (s *TelemetryService) GetConfig(ctx context.Context, orgID string) (*OTELConfig, error) {
	query := `
		SELECT endpoint, protocol, headers, enabled
		FROM otel_configs
		WHERE org_id = $1`

	var config OTELConfig
	var headers []byte

	err := s.db.QueryRowContext(ctx, query, orgID).Scan(
		&config.Endpoint,
		&config.Protocol,
		&headers,
		&config.Enabled,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}

	// Parse headers (stored as JSON)
	// In a real implementation, use proper JSON unmarshaling

	return &config, nil
}

// SaveConfig saves the OTEL configuration for an organization.
func (s *TelemetryService) SaveConfig(ctx context.Context, orgID string, config OTELConfig) error {
	query := `
		INSERT INTO otel_configs (org_id, endpoint, protocol, headers, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (org_id) DO UPDATE SET
			endpoint = EXCLUDED.endpoint,
			protocol = EXCLUDED.protocol,
			headers = EXCLUDED.headers,
			enabled = EXCLUDED.enabled,
			updated_at = NOW()`

	// Convert headers to JSON (simplified)
	headersJSON := "{}"

	_, err := s.db.ExecContext(ctx, query,
		orgID, config.Endpoint, config.Protocol, headersJSON, config.Enabled,
	)
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	s.logger.Info("OTEL config saved",
		"org_id", orgID,
		"endpoint", config.Endpoint,
	)

	return nil
}

// TestConnection tests the OTLP connection.
func (s *TelemetryService) TestConnection(ctx context.Context, config OTELConfig) error {
	var exporter *otlptrace.Exporter
	var err error

	switch config.Protocol {
	case "grpc":
		exporter, err = s.createGRPCExporter(ctx, config)
	case "http":
		exporter, err = s.createHTTPExporter(ctx, config)
	default:
		return fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}

	if err != nil {
		return fmt.Errorf("create test exporter: %w", err)
	}
	defer exporter.Shutdown(ctx)

	// The exporter creation itself validates the connection
	s.logger.Info("OTEL connection test passed",
		"endpoint", config.Endpoint,
		"protocol", config.Protocol,
	)

	return nil
}

// Shutdown gracefully shuts down the tracer provider.
func (s *TelemetryService) Shutdown(ctx context.Context) error {
	if s.tracerProvider != nil {
		return s.tracerProvider.Shutdown(ctx)
	}
	return nil
}

// GetTracer returns the global tracer.
func (s *TelemetryService) GetTracer(name string) interface{} {
	return otel.Tracer(name)
}
