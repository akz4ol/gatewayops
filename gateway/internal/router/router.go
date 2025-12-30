// Package router sets up the HTTP router and middleware chain.
package router

import (
	"net/http"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// Dependencies holds all dependencies needed by the router.
type Dependencies struct {
	Config        *config.Config
	Logger        zerolog.Logger
	AuthStore     middleware.AuthStore
	RateLimiter   middleware.RateLimiter
	MCPHandler    *handler.MCPHandler
	HealthHandler *handler.HealthHandler
}

// New creates a new router with all middleware and routes configured.
func New(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// Global middleware (order matters!)
	r.Use(chimiddleware.RequestID)                                    // 1. Add request ID
	r.Use(chimiddleware.RealIP)                                       // 2. Get real IP from headers
	r.Use(middleware.Recoverer(deps.Logger))                          // 3. Recover from panics
	r.Use(middleware.Logger(deps.Logger))                             // 4. Log requests
	r.Use(middleware.Trace())                                         // 5. Add trace context
	r.Use(chimiddleware.Timeout(deps.Config.Server.WriteTimeout))     // 6. Request timeout

	// Health endpoints (no auth required)
	r.Get("/health", deps.HealthHandler.Health)
	r.Get("/ready", deps.HealthHandler.Ready)

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// MCP routes (require authentication)
		r.Route("/mcp/{server}", func(r chi.Router) {
			r.Use(middleware.Auth(deps.AuthStore, deps.Logger))           // Authentication
			r.Use(middleware.RateLimit(deps.RateLimiter, deps.Logger))    // Rate limiting

			// Tools
			r.Post("/tools/call", deps.MCPHandler.ToolsCall)
			r.Post("/tools/list", deps.MCPHandler.ToolsList)

			// Resources
			r.Post("/resources/read", deps.MCPHandler.ResourcesRead)
			r.Post("/resources/list", deps.MCPHandler.ResourcesList)

			// Prompts
			r.Post("/prompts/get", deps.MCPHandler.PromptsGet)
			r.Post("/prompts/list", deps.MCPHandler.PromptsList)
		})

		// API Keys management (require authentication)
		r.Route("/api-keys", func(r chi.Router) {
			r.Use(middleware.Auth(deps.AuthStore, deps.Logger))

			r.Get("/", notImplemented)
			r.Post("/", notImplemented)
			r.Get("/{keyID}", notImplemented)
			r.Delete("/{keyID}", notImplemented)
			r.Post("/{keyID}/rotate", notImplemented)
		})

		// Traces (require authentication)
		r.Route("/traces", func(r chi.Router) {
			r.Use(middleware.Auth(deps.AuthStore, deps.Logger))

			r.Get("/", notImplemented)
			r.Get("/{traceID}", notImplemented)
			r.Post("/search", notImplemented)
		})

		// Costs (require authentication)
		r.Route("/costs", func(r chi.Router) {
			r.Use(middleware.Auth(deps.AuthStore, deps.Logger))

			r.Get("/summary", notImplemented)
			r.Get("/by-team", notImplemented)
			r.Get("/by-server", notImplemented)
			r.Get("/trace/{traceID}", notImplemented)
		})
	})

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteError(w, http.StatusNotFound, "not_found", "The requested resource was not found")
	})

	// 405 handler
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		handler.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "The requested method is not allowed")
	})

	return r
}

// notImplemented is a placeholder handler for endpoints not yet implemented.
func notImplemented(w http.ResponseWriter, r *http.Request) {
	handler.WriteError(w, http.StatusNotImplemented, "not_implemented", "This endpoint is not yet implemented")
}
