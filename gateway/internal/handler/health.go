package handler

import (
	"net/http"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/server"
)

// HealthChecker defines interface for service health checks.
type HealthChecker interface {
	Health() bool
	Ready() bool
}

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	checkers []HealthChecker
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(checkers ...HealthChecker) *HealthHandler {
	return &HealthHandler{checkers: checkers}
}

// HealthResponse represents health check response.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

// ReadyResponse represents readiness check response.
type ReadyResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// Health handles GET /health - liveness check.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Liveness: is the service running?
	healthy := true
	for _, checker := range h.checkers {
		if !checker.Health() {
			healthy = false
			break
		}
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !healthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	WriteJSON(w, httpStatus, HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    server.Uptime().String(),
	})
}

// Ready handles GET /ready - readiness check.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// Readiness: is the service ready to accept traffic?
	checks := make(map[string]string)
	allReady := true

	for i, checker := range h.checkers {
		if checker.Ready() {
			checks[string(rune('0'+i))] = "ready"
		} else {
			checks[string(rune('0'+i))] = "not_ready"
			allReady = false
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allReady {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	WriteJSON(w, httpStatus, ReadyResponse{
		Status: status,
		Checks: checks,
	})
}
