// Package otel provides OpenTelemetry export functionality.
package otel

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Exporter manages OpenTelemetry export configurations and sending.
type Exporter struct {
	logger  zerolog.Logger
	configs map[uuid.UUID]*domain.TelemetryConfig
	mu      sync.RWMutex
	client  *http.Client

	// Stats
	tracesExported  int64
	metricsExported int64
	logsExported    int64
	exportErrors    int64
	bytesSent       int64
	totalLatencyMs  int64
	exportCount     int64
	lastExportAt    time.Time

	// Batch queue
	spanQueue   []domain.TelemetrySpan
	metricQueue []domain.TelemetryMetric
	queueMu     sync.Mutex
}

// NewExporter creates a new OpenTelemetry exporter.
func NewExporter(logger zerolog.Logger) *Exporter {
	e := &Exporter{
		logger:      logger,
		configs:     make(map[uuid.UUID]*domain.TelemetryConfig),
		client:      &http.Client{Timeout: 30 * time.Second},
		spanQueue:   make([]domain.TelemetrySpan, 0),
		metricQueue: make([]domain.TelemetryMetric, 0),
	}

	// Create demo config
	e.createDemoConfig()

	// Start background export loop
	go e.exportLoop()

	logger.Info().Msg("OpenTelemetry exporter initialized")
	return e
}

func (e *Exporter) createDemoConfig() {
	config := &domain.TelemetryConfig{
		ID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		OrgID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:          "Demo OTLP Exporter",
		ExporterType:  domain.TelemetryExporterOTLP,
		Endpoint:      "https://otel-collector.example.com:4318",
		Protocol:      domain.TelemetryProtocolHTTP,
		Headers:       map[string]string{"Authorization": "Bearer demo-token"},
		Insecure:      false,
		Enabled:       false, // Disabled by default for demo
		ExportTraces:  true,
		ExportMetrics: true,
		ExportLogs:    false,
		SampleRate:    1.0,
		BatchSize:     100,
		BatchTimeout:  5000,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	e.configs[config.ID] = config
}

// CreateConfig creates a new telemetry configuration.
func (e *Exporter) CreateConfig(input domain.TelemetryConfigInput, orgID uuid.UUID) *domain.TelemetryConfig {
	e.mu.Lock()
	defer e.mu.Unlock()

	config := &domain.TelemetryConfig{
		ID:            uuid.New(),
		OrgID:         orgID,
		Name:          input.Name,
		ExporterType:  input.ExporterType,
		Endpoint:      input.Endpoint,
		Protocol:      input.Protocol,
		Headers:       input.Headers,
		Insecure:      input.Insecure,
		Enabled:       input.Enabled,
		ExportTraces:  input.ExportTraces,
		ExportMetrics: input.ExportMetrics,
		ExportLogs:    input.ExportLogs,
		SampleRate:    input.SampleRate,
		BatchSize:     input.BatchSize,
		BatchTimeout:  input.BatchTimeout,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if config.SampleRate <= 0 || config.SampleRate > 1 {
		config.SampleRate = 1.0
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.BatchTimeout <= 0 {
		config.BatchTimeout = 5000
	}

	e.configs[config.ID] = config

	e.logger.Info().
		Str("config_id", config.ID.String()).
		Str("name", config.Name).
		Str("endpoint", config.Endpoint).
		Msg("Telemetry config created")

	return config
}

// GetConfig returns a config by ID.
func (e *Exporter) GetConfig(id uuid.UUID) *domain.TelemetryConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.configs[id]
}

// ListConfigs returns all configs.
func (e *Exporter) ListConfigs() []domain.TelemetryConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	configs := make([]domain.TelemetryConfig, 0, len(e.configs))
	for _, c := range e.configs {
		configs = append(configs, *c)
	}
	return configs
}

// UpdateConfig updates an existing config.
func (e *Exporter) UpdateConfig(id uuid.UUID, input domain.TelemetryConfigInput) *domain.TelemetryConfig {
	e.mu.Lock()
	defer e.mu.Unlock()

	config, exists := e.configs[id]
	if !exists {
		return nil
	}

	config.Name = input.Name
	config.ExporterType = input.ExporterType
	config.Endpoint = input.Endpoint
	config.Protocol = input.Protocol
	config.Headers = input.Headers
	config.Insecure = input.Insecure
	config.Enabled = input.Enabled
	config.ExportTraces = input.ExportTraces
	config.ExportMetrics = input.ExportMetrics
	config.ExportLogs = input.ExportLogs
	config.SampleRate = input.SampleRate
	config.BatchSize = input.BatchSize
	config.BatchTimeout = input.BatchTimeout
	config.UpdatedAt = time.Now()

	return config
}

// DeleteConfig deletes a config.
func (e *Exporter) DeleteConfig(id uuid.UUID) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.configs[id]; exists {
		delete(e.configs, id)
		return true
	}
	return false
}

// TestConfig tests connectivity to the OTLP endpoint.
func (e *Exporter) TestConfig(id uuid.UUID) domain.TelemetryTestResult {
	e.mu.RLock()
	config, exists := e.configs[id]
	e.mu.RUnlock()

	if !exists {
		return domain.TelemetryTestResult{
			Success: false,
			Message: "Configuration not found",
		}
	}

	// Create test span
	testSpan := domain.TelemetrySpan{
		TraceID:   generateTraceID(),
		SpanID:    generateSpanID(),
		Name:      "gatewayops.test",
		Kind:      domain.SpanKindInternal,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Millisecond),
		Status:    domain.SpanStatusOK,
		Attributes: map[string]string{
			"test": "true",
		},
	}

	start := time.Now()
	err := e.exportSpans(*config, []domain.TelemetrySpan{testSpan})
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return domain.TelemetryTestResult{
			Success:   false,
			Message:   fmt.Sprintf("Export failed: %v", err),
			LatencyMs: latency,
		}
	}

	return domain.TelemetryTestResult{
		Success:   true,
		Message:   "Successfully connected and exported test span",
		LatencyMs: latency,
	}
}

// RecordSpan adds a span to the export queue.
func (e *Exporter) RecordSpan(span domain.TelemetrySpan) {
	e.queueMu.Lock()
	defer e.queueMu.Unlock()
	e.spanQueue = append(e.spanQueue, span)
}

// RecordMetric adds a metric to the export queue.
func (e *Exporter) RecordMetric(metric domain.TelemetryMetric) {
	e.queueMu.Lock()
	defer e.queueMu.Unlock()
	e.metricQueue = append(e.metricQueue, metric)
}

// GetStats returns export statistics.
func (e *Exporter) GetStats() domain.TelemetryStats {
	count := atomic.LoadInt64(&e.exportCount)
	avgLatency := float64(0)
	if count > 0 {
		avgLatency = float64(atomic.LoadInt64(&e.totalLatencyMs)) / float64(count)
	}

	return domain.TelemetryStats{
		TracesExported:  atomic.LoadInt64(&e.tracesExported),
		MetricsExported: atomic.LoadInt64(&e.metricsExported),
		LogsExported:    atomic.LoadInt64(&e.logsExported),
		ExportErrors:    atomic.LoadInt64(&e.exportErrors),
		LastExportAt:    e.lastExportAt,
		BytesSent:       atomic.LoadInt64(&e.bytesSent),
		AvgLatencyMs:    avgLatency,
	}
}

// exportLoop runs periodically to export queued spans and metrics.
func (e *Exporter) exportLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		e.flush()
	}
}

// flush exports all queued spans and metrics.
func (e *Exporter) flush() {
	e.queueMu.Lock()
	spans := e.spanQueue
	metrics := e.metricQueue
	e.spanQueue = make([]domain.TelemetrySpan, 0)
	e.metricQueue = make([]domain.TelemetryMetric, 0)
	e.queueMu.Unlock()

	if len(spans) == 0 && len(metrics) == 0 {
		return
	}

	e.mu.RLock()
	configs := make([]*domain.TelemetryConfig, 0)
	for _, c := range e.configs {
		if c.Enabled {
			configs = append(configs, c)
		}
	}
	e.mu.RUnlock()

	for _, config := range configs {
		if config.ExportTraces && len(spans) > 0 {
			// Apply sampling
			sampled := e.sampleSpans(spans, config.SampleRate)
			if len(sampled) > 0 {
				if err := e.exportSpans(*config, sampled); err != nil {
					e.logger.Error().
						Err(err).
						Str("config_id", config.ID.String()).
						Msg("Failed to export spans")
					atomic.AddInt64(&e.exportErrors, 1)
				} else {
					atomic.AddInt64(&e.tracesExported, int64(len(sampled)))
				}
			}
		}

		if config.ExportMetrics && len(metrics) > 0 {
			if err := e.exportMetrics(*config, metrics); err != nil {
				e.logger.Error().
					Err(err).
					Str("config_id", config.ID.String()).
					Msg("Failed to export metrics")
				atomic.AddInt64(&e.exportErrors, 1)
			} else {
				atomic.AddInt64(&e.metricsExported, int64(len(metrics)))
			}
		}
	}

	e.lastExportAt = time.Now()
}

func (e *Exporter) sampleSpans(spans []domain.TelemetrySpan, rate float64) []domain.TelemetrySpan {
	if rate >= 1.0 {
		return spans
	}

	sampled := make([]domain.TelemetrySpan, 0)
	for i, span := range spans {
		// Simple deterministic sampling based on index
		if float64(i%100)/100.0 < rate {
			sampled = append(sampled, span)
		}
	}
	return sampled
}

func (e *Exporter) exportSpans(config domain.TelemetryConfig, spans []domain.TelemetrySpan) error {
	// In demo mode, just log
	if config.Endpoint == "https://otel-collector.example.com:4318" {
		e.logger.Debug().
			Int("span_count", len(spans)).
			Msg("Demo mode: Would export spans to OTLP endpoint")
		return nil
	}

	// Convert to OTLP format
	payload := e.buildOTLPTracePayload(spans)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Determine endpoint URL
	endpoint := config.Endpoint
	if config.Protocol == domain.TelemetryProtocolHTTP {
		endpoint = fmt.Sprintf("%s/v1/traces", config.Endpoint)
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(context.Background(), "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()
	atomic.AddInt64(&e.totalLatencyMs, latency)
	atomic.AddInt64(&e.exportCount, 1)
	atomic.AddInt64(&e.bytesSent, int64(len(body)))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

func (e *Exporter) exportMetrics(config domain.TelemetryConfig, metrics []domain.TelemetryMetric) error {
	// In demo mode, just log
	if config.Endpoint == "https://otel-collector.example.com:4318" {
		e.logger.Debug().
			Int("metric_count", len(metrics)).
			Msg("Demo mode: Would export metrics to OTLP endpoint")
		return nil
	}

	// For now, just log - full metrics export would require more complex OTLP metrics format
	e.logger.Debug().
		Int("metric_count", len(metrics)).
		Str("endpoint", config.Endpoint).
		Msg("Exporting metrics")

	return nil
}

func (e *Exporter) buildOTLPTracePayload(spans []domain.TelemetrySpan) domain.OTLPExportRequest {
	otlpSpans := make([]domain.OTLPSpan, 0, len(spans))
	for _, span := range spans {
		attrs := make([]domain.KeyValue, 0)
		for k, v := range span.Attributes {
			attrs = append(attrs, domain.KeyValue{
				Key:   k,
				Value: domain.AttrValue{StringValue: v},
			})
		}

		statusCode := 0 // Unset
		switch span.Status {
		case domain.SpanStatusOK:
			statusCode = 1
		case domain.SpanStatusError:
			statusCode = 2
		}

		kind := 0 // Internal
		switch span.Kind {
		case domain.SpanKindServer:
			kind = 2
		case domain.SpanKindClient:
			kind = 3
		case domain.SpanKindProducer:
			kind = 4
		case domain.SpanKindConsumer:
			kind = 5
		}

		otlpSpans = append(otlpSpans, domain.OTLPSpan{
			TraceID:           span.TraceID,
			SpanID:            span.SpanID,
			ParentSpanID:      span.ParentSpanID,
			Name:              span.Name,
			Kind:              kind,
			StartTimeUnixNano: span.StartTime.UnixNano(),
			EndTimeUnixNano:   span.EndTime.UnixNano(),
			Attributes:        attrs,
			Status: domain.OTLPStatus{
				Code:    statusCode,
				Message: span.StatusMsg,
			},
		})
	}

	return domain.OTLPExportRequest{
		ResourceSpans: []domain.ResourceSpan{
			{
				Resource: domain.Resource{
					Attributes: []domain.KeyValue{
						{Key: "service.name", Value: domain.AttrValue{StringValue: "gatewayops"}},
						{Key: "service.version", Value: domain.AttrValue{StringValue: "1.0.0"}},
					},
				},
				ScopeSpans: []domain.ScopeSpan{
					{
						Scope: domain.InstrumentationScope{
							Name:    "gatewayops",
							Version: "1.0.0",
						},
						Spans: otlpSpans,
					},
				},
			},
		},
	}
}

func generateTraceID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> (i * 8))
	}
	return hex.EncodeToString(b)
}

func generateSpanID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> (i * 8))
	}
	return hex.EncodeToString(b)
}

// RecordMCPCall records an MCP call as a span for export.
func (e *Exporter) RecordMCPCall(traceID, serverName, toolName string, duration time.Duration, status string, costUSD float64) {
	span := domain.TelemetrySpan{
		TraceID:   traceID,
		SpanID:    generateSpanID(),
		Name:      fmt.Sprintf("mcp.%s.%s", serverName, toolName),
		Kind:      domain.SpanKindClient,
		StartTime: time.Now().Add(-duration),
		EndTime:   time.Now(),
		Status:    domain.SpanStatusOK,
		Attributes: map[string]string{
			"mcp.server":    serverName,
			"mcp.tool":      toolName,
			"mcp.cost_usd":  fmt.Sprintf("%.6f", costUSD),
			"mcp.status":    status,
			"mcp.duration":  duration.String(),
		},
	}

	if status == "error" {
		span.Status = domain.SpanStatusError
	}

	e.RecordSpan(span)
}
