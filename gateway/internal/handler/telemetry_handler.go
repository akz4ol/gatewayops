package handler

import (
	"encoding/json"
	"net/http"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/otel"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// TelemetryHandler handles telemetry-related HTTP requests.
type TelemetryHandler struct {
	logger   zerolog.Logger
	exporter *otel.Exporter
}

// NewTelemetryHandler creates a new telemetry handler.
func NewTelemetryHandler(logger zerolog.Logger, exporter *otel.Exporter) *TelemetryHandler {
	return &TelemetryHandler{
		logger:   logger,
		exporter: exporter,
	}
}

// ListConfigs returns all telemetry configurations.
func (h *TelemetryHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	configs := h.exporter.ListConfigs()
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"configs": configs,
		"total":   len(configs),
	})
}

// GetConfig returns a single configuration by ID.
func (h *TelemetryHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "configID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid config ID")
		return
	}

	config := h.exporter.GetConfig(id)
	if config == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Configuration not found")
		return
	}

	WriteJSON(w, http.StatusOK, config)
}

// CreateConfig creates a new telemetry configuration.
func (h *TelemetryHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var input domain.TelemetryConfigInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}
	if input.Endpoint == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Endpoint is required")
		return
	}
	if input.ExporterType == "" {
		input.ExporterType = domain.TelemetryExporterOTLP
	}
	if input.Protocol == "" {
		input.Protocol = domain.TelemetryProtocolHTTP
	}

	// Demo org
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	config := h.exporter.CreateConfig(input, orgID)
	WriteJSON(w, http.StatusCreated, config)
}

// UpdateConfig updates an existing configuration.
func (h *TelemetryHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "configID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid config ID")
		return
	}

	var input domain.TelemetryConfigInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	config := h.exporter.UpdateConfig(id, input)
	if config == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Configuration not found")
		return
	}

	WriteJSON(w, http.StatusOK, config)
}

// DeleteConfig deletes a configuration.
func (h *TelemetryHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "configID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid config ID")
		return
	}

	if !h.exporter.DeleteConfig(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Configuration not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// TestConfig tests connectivity to the OTLP endpoint.
func (h *TelemetryHandler) TestConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "configID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid config ID")
		return
	}

	result := h.exporter.TestConfig(id)

	status := http.StatusOK
	if !result.Success {
		status = http.StatusBadRequest
	}

	WriteJSON(w, status, result)
}

// GetStats returns telemetry export statistics.
func (h *TelemetryHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.exporter.GetStats()
	WriteJSON(w, http.StatusOK, stats)
}

// ExportSpan manually exports a span (for testing).
func (h *TelemetryHandler) ExportSpan(w http.ResponseWriter, r *http.Request) {
	var span domain.TelemetrySpan
	if err := json.NewDecoder(r.Body).Decode(&span); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	h.exporter.RecordSpan(span)

	WriteJSON(w, http.StatusAccepted, map[string]string{
		"status":  "queued",
		"message": "Span queued for export",
	})
}

// ExportMetric manually exports a metric (for testing).
func (h *TelemetryHandler) ExportMetric(w http.ResponseWriter, r *http.Request) {
	var metric domain.TelemetryMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	h.exporter.RecordMetric(metric)

	WriteJSON(w, http.StatusAccepted, map[string]string{
		"status":  "queued",
		"message": "Metric queued for export",
	})
}

// GetSupportedExporters returns the list of supported exporter types.
func (h *TelemetryHandler) GetSupportedExporters(w http.ResponseWriter, r *http.Request) {
	exporters := []map[string]interface{}{
		{
			"type":        domain.TelemetryExporterOTLP,
			"name":        "OpenTelemetry Protocol (OTLP)",
			"description": "Standard OpenTelemetry Protocol exporter for traces, metrics, and logs",
			"protocols":   []string{"grpc", "http"},
		},
		{
			"type":        domain.TelemetryExporterJaeger,
			"name":        "Jaeger",
			"description": "Jaeger distributed tracing backend",
			"protocols":   []string{"grpc", "http"},
		},
		{
			"type":        domain.TelemetryExporterZipkin,
			"name":        "Zipkin",
			"description": "Zipkin distributed tracing system",
			"protocols":   []string{"http"},
		},
		{
			"type":        domain.TelemetryExporterDatadog,
			"name":        "Datadog",
			"description": "Datadog APM and monitoring",
			"protocols":   []string{"http"},
		},
		{
			"type":        domain.TelemetryExporterNewRelic,
			"name":        "New Relic",
			"description": "New Relic observability platform",
			"protocols":   []string{"http"},
		},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"exporters": exporters,
	})
}
