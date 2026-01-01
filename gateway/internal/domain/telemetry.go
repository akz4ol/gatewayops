package domain

import (
	"time"

	"github.com/google/uuid"
)

// TelemetryExporterType represents the type of telemetry exporter.
type TelemetryExporterType string

const (
	TelemetryExporterOTLP       TelemetryExporterType = "otlp"
	TelemetryExporterJaeger     TelemetryExporterType = "jaeger"
	TelemetryExporterZipkin     TelemetryExporterType = "zipkin"
	TelemetryExporterPrometheus TelemetryExporterType = "prometheus"
	TelemetryExporterDatadog    TelemetryExporterType = "datadog"
	TelemetryExporterNewRelic   TelemetryExporterType = "newrelic"
)

// TelemetryProtocol represents the protocol used for exporting.
type TelemetryProtocol string

const (
	TelemetryProtocolGRPC TelemetryProtocol = "grpc"
	TelemetryProtocolHTTP TelemetryProtocol = "http"
)

// TelemetryConfig represents the OpenTelemetry export configuration.
type TelemetryConfig struct {
	ID            uuid.UUID             `json:"id"`
	OrgID         uuid.UUID             `json:"org_id"`
	Name          string                `json:"name"`
	ExporterType  TelemetryExporterType `json:"exporter_type"`
	Endpoint      string                `json:"endpoint"`
	Protocol      TelemetryProtocol     `json:"protocol"`
	Headers       map[string]string     `json:"headers,omitempty"`
	Insecure      bool                  `json:"insecure"`
	Enabled       bool                  `json:"enabled"`
	ExportTraces  bool                  `json:"export_traces"`
	ExportMetrics bool                  `json:"export_metrics"`
	ExportLogs    bool                  `json:"export_logs"`
	SampleRate    float64               `json:"sample_rate"` // 0.0 to 1.0
	BatchSize     int                   `json:"batch_size"`
	BatchTimeout  int                   `json:"batch_timeout_ms"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	LastExportAt  *time.Time            `json:"last_export_at,omitempty"`
	LastError     string                `json:"last_error,omitempty"`
}

// TelemetryConfigInput represents input for creating/updating telemetry config.
type TelemetryConfigInput struct {
	Name          string                `json:"name"`
	ExporterType  TelemetryExporterType `json:"exporter_type"`
	Endpoint      string                `json:"endpoint"`
	Protocol      TelemetryProtocol     `json:"protocol"`
	Headers       map[string]string     `json:"headers,omitempty"`
	Insecure      bool                  `json:"insecure"`
	Enabled       bool                  `json:"enabled"`
	ExportTraces  bool                  `json:"export_traces"`
	ExportMetrics bool                  `json:"export_metrics"`
	ExportLogs    bool                  `json:"export_logs"`
	SampleRate    float64               `json:"sample_rate"`
	BatchSize     int                   `json:"batch_size"`
	BatchTimeout  int                   `json:"batch_timeout_ms"`
}

// TelemetrySpan represents a span to be exported.
type TelemetrySpan struct {
	TraceID      string            `json:"trace_id"`
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id,omitempty"`
	Name         string            `json:"name"`
	Kind         SpanKind          `json:"kind"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Status       SpanStatus        `json:"status"`
	StatusMsg    string            `json:"status_message,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	Events       []SpanEvent       `json:"events,omitempty"`
}

// SpanKind represents the type of span.
type SpanKind string

const (
	SpanKindInternal SpanKind = "internal"
	SpanKindServer   SpanKind = "server"
	SpanKindClient   SpanKind = "client"
	SpanKindProducer SpanKind = "producer"
	SpanKindConsumer SpanKind = "consumer"
)

// SpanStatus represents the status of a span.
type SpanStatus string

const (
	SpanStatusUnset SpanStatus = "unset"
	SpanStatusOK    SpanStatus = "ok"
	SpanStatusError SpanStatus = "error"
)

// SpanEvent represents an event within a span.
type SpanEvent struct {
	Name       string            `json:"name"`
	Timestamp  time.Time         `json:"timestamp"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// TelemetryMetric represents a metric to be exported.
type TelemetryMetric struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Unit        string            `json:"unit,omitempty"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// MetricType represents the type of metric.
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// TelemetryStats represents telemetry export statistics.
type TelemetryStats struct {
	TracesExported  int64     `json:"traces_exported"`
	MetricsExported int64     `json:"metrics_exported"`
	LogsExported    int64     `json:"logs_exported"`
	ExportErrors    int64     `json:"export_errors"`
	LastExportAt    time.Time `json:"last_export_at"`
	BytesSent       int64     `json:"bytes_sent"`
	AvgLatencyMs    float64   `json:"avg_latency_ms"`
}

// TelemetryTestResult represents the result of testing a telemetry config.
type TelemetryTestResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	LatencyMs  int64  `json:"latency_ms"`
	StatusCode int    `json:"status_code,omitempty"`
}

// OTLPExportRequest represents an OTLP export request.
type OTLPExportRequest struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans,omitempty"`
}

// ResourceSpan represents spans with resource attributes.
type ResourceSpan struct {
	Resource   Resource    `json:"resource"`
	ScopeSpans []ScopeSpan `json:"scopeSpans"`
}

// Resource represents resource attributes.
type Resource struct {
	Attributes []KeyValue `json:"attributes"`
}

// ScopeSpan represents spans within an instrumentation scope.
type ScopeSpan struct {
	Scope InstrumentationScope `json:"scope"`
	Spans []OTLPSpan           `json:"spans"`
}

// InstrumentationScope represents the instrumentation scope.
type InstrumentationScope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// OTLPSpan represents a span in OTLP format.
type OTLPSpan struct {
	TraceID           string     `json:"traceId"`
	SpanID            string     `json:"spanId"`
	ParentSpanID      string     `json:"parentSpanId,omitempty"`
	Name              string     `json:"name"`
	Kind              int        `json:"kind"`
	StartTimeUnixNano int64      `json:"startTimeUnixNano"`
	EndTimeUnixNano   int64      `json:"endTimeUnixNano"`
	Attributes        []KeyValue `json:"attributes,omitempty"`
	Status            OTLPStatus `json:"status"`
}

// OTLPStatus represents span status in OTLP format.
type OTLPStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// KeyValue represents a key-value pair for attributes.
type KeyValue struct {
	Key   string     `json:"key"`
	Value AttrValue  `json:"value"`
}

// AttrValue represents an attribute value.
type AttrValue struct {
	StringValue string `json:"stringValue,omitempty"`
	IntValue    int64  `json:"intValue,omitempty"`
	DoubleValue float64 `json:"doubleValue,omitempty"`
	BoolValue   bool   `json:"boolValue,omitempty"`
}
