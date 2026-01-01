// Package router sets up the HTTP router and middleware chain.
package router

import (
	"net/http"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
)

// Dependencies holds all dependencies needed by the router.
type Dependencies struct {
	Config            *config.Config
	Logger            zerolog.Logger
	AuthStore         middleware.AuthStore
	RateLimiter       middleware.RateLimiter
	InjectionDetector middleware.InjectionDetector
	AuditLogger       middleware.AuditLogger
	MCPHandler        *handler.MCPHandler
	HealthHandler     *handler.HealthHandler
	TraceHandler      *handler.TraceHandler
	CostHandler       *handler.CostHandler
	APIKeyHandler     *handler.APIKeyHandler
	MetricsHandler    *handler.MetricsHandler
	DocsHandler       *handler.DocsHandler
	SafetyHandler     *handler.SafetyHandler
	AuditHandler      *handler.AuditHandler
	AlertHandler      *handler.AlertHandler
	TelemetryHandler  *handler.TelemetryHandler
	ApprovalHandler   *handler.ApprovalHandler
	RBACHandler       *handler.RBACHandler
	SSOHandler        *handler.SSOHandler
}

// New creates a new router with all middleware and routes configured.
func New(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// CORS middleware - must be first
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://gatewayops-dashboard.fly.dev", "http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Trace-ID"},
		ExposedHeaders:   []string{"X-MCP-Server", "X-MCP-Duration-Ms", "X-MCP-Cost"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Global middleware (order matters!)
	r.Use(chimiddleware.RequestID)                                // 1. Add request ID
	r.Use(chimiddleware.RealIP)                                   // 2. Get real IP from headers
	r.Use(middleware.Recoverer(deps.Logger))                      // 3. Recover from panics
	r.Use(middleware.Logger(deps.Logger))                         // 4. Log requests
	r.Use(middleware.Trace())                                     // 5. Add trace context
	r.Use(chimiddleware.Timeout(deps.Config.Server.WriteTimeout)) // 6. Request timeout

	// Health endpoints (no auth required)
	r.Get("/health", deps.HealthHandler.Health)
	r.Get("/ready", deps.HealthHandler.Ready)

	// API Documentation (no auth required)
	if deps.DocsHandler != nil {
		r.Get("/docs", deps.DocsHandler.SwaggerUI)
		r.Get("/openapi.yaml", deps.DocsHandler.OpenAPISpec)
	}

	// SSO OAuth callbacks (no auth required - part of login flow)
	if deps.SSOHandler != nil {
		r.Get("/v1/sso/authorize/{providerID}", deps.SSOHandler.Authorize)
		r.Get("/v1/sso/callback/{providerID}", deps.SSOHandler.Callback)
		r.Post("/v1/sso/logout", deps.SSOHandler.Logout)
	}

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// MCP routes (require authentication)
		r.Route("/mcp/{server}", func(r chi.Router) {
			r.Use(middleware.Auth(deps.AuthStore, deps.Logger))        // Authentication
			r.Use(middleware.RateLimit(deps.RateLimiter, deps.Logger)) // Rate limiting
			if deps.InjectionDetector != nil {
				r.Use(middleware.Injection(deps.InjectionDetector, deps.Logger)) // Prompt injection detection
			}
			if deps.AuditLogger != nil {
				r.Use(middleware.Audit(deps.AuditLogger, deps.Logger)) // Audit logging
			}

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

		// Dashboard metrics (public for demo - in production, add auth)
		r.Route("/metrics", func(r chi.Router) {
			// NOTE: Auth disabled for demo. Enable for production:
			// r.Use(middleware.Auth(deps.AuthStore, deps.Logger))

			r.Get("/overview", deps.MetricsHandler.Overview)
			r.Get("/requests-chart", deps.MetricsHandler.RequestsChart)
			r.Get("/top-servers", deps.MetricsHandler.TopServers)
			r.Get("/recent-traces", deps.MetricsHandler.RecentTraces)
		})

		// Traces - public for demo
		r.Route("/traces", func(r chi.Router) {
			// NOTE: Auth disabled for demo
			r.Get("/", deps.TraceHandler.List)
			r.Get("/stats", deps.TraceHandler.Stats)
			r.Get("/{traceID}", deps.TraceHandler.Get)
		})

		// Costs - public for demo
		r.Route("/costs", func(r chi.Router) {
			// NOTE: Auth disabled for demo
			r.Get("/summary", deps.CostHandler.Summary)
			r.Get("/by-team", deps.CostHandler.ByTeam)
			r.Get("/by-server", deps.CostHandler.ByServer)
			r.Get("/daily", deps.CostHandler.Daily)
		})

		// API Keys - public for demo
		r.Route("/api-keys", func(r chi.Router) {
			// NOTE: Auth disabled for demo
			r.Get("/", deps.APIKeyHandler.List)
			r.Post("/", deps.APIKeyHandler.Create)
			r.Get("/{keyID}", deps.APIKeyHandler.Get)
			r.Delete("/{keyID}", deps.APIKeyHandler.Delete)
			r.Post("/{keyID}/rotate", deps.APIKeyHandler.Rotate)
		})

		// Safety policies and detection - public for demo
		if deps.SafetyHandler != nil {
			r.Route("/safety", func(r chi.Router) {
				// Policies
				r.Get("/policies", deps.SafetyHandler.ListPolicies)
				r.Post("/policies", deps.SafetyHandler.CreatePolicy)
				r.Get("/policies/{policyID}", deps.SafetyHandler.GetPolicy)
				r.Put("/policies/{policyID}", deps.SafetyHandler.UpdatePolicy)
				r.Delete("/policies/{policyID}", deps.SafetyHandler.DeletePolicy)

				// Detection testing
				r.Post("/test", deps.SafetyHandler.TestInput)

				// Detections
				r.Get("/detections", deps.SafetyHandler.ListDetections)
				r.Get("/summary", deps.SafetyHandler.GetSummary)
			})
		}

		// Audit logs - public for demo
		if deps.AuditHandler != nil {
			r.Route("/audit-logs", func(r chi.Router) {
				r.Get("/", deps.AuditHandler.List)
				r.Get("/search", deps.AuditHandler.Search)
				r.Get("/export", deps.AuditHandler.Export)
				r.Get("/stats", deps.AuditHandler.Stats)
				r.Get("/{logID}", deps.AuditHandler.Get)
			})
		}

		// Alerts - public for demo
		if deps.AlertHandler != nil {
			r.Route("/alerts", func(r chi.Router) {
				// Alerts
				r.Get("/", deps.AlertHandler.ListAlerts)
				r.Get("/active", deps.AlertHandler.GetActiveAlerts)
				r.Post("/test", deps.AlertHandler.TriggerTestAlert)
				r.Post("/{alertID}/acknowledge", deps.AlertHandler.AcknowledgeAlert)
				r.Post("/{alertID}/resolve", deps.AlertHandler.ResolveAlert)

				// Rules
				r.Route("/rules", func(r chi.Router) {
					r.Get("/", deps.AlertHandler.ListRules)
					r.Post("/", deps.AlertHandler.CreateRule)
					r.Get("/{ruleID}", deps.AlertHandler.GetRule)
					r.Put("/{ruleID}", deps.AlertHandler.UpdateRule)
					r.Delete("/{ruleID}", deps.AlertHandler.DeleteRule)
				})

				// Channels
				r.Route("/channels", func(r chi.Router) {
					r.Get("/", deps.AlertHandler.ListChannels)
					r.Post("/", deps.AlertHandler.CreateChannel)
					r.Get("/{channelID}", deps.AlertHandler.GetChannel)
					r.Put("/{channelID}", deps.AlertHandler.UpdateChannel)
					r.Delete("/{channelID}", deps.AlertHandler.DeleteChannel)
					r.Post("/{channelID}/test", deps.AlertHandler.TestChannel)
				})
			})
		}

		// Telemetry / OpenTelemetry Export - public for demo
		if deps.TelemetryHandler != nil {
			r.Route("/telemetry", func(r chi.Router) {
				// Exporters info
				r.Get("/exporters", deps.TelemetryHandler.GetSupportedExporters)
				r.Get("/stats", deps.TelemetryHandler.GetStats)

				// Configurations
				r.Route("/configs", func(r chi.Router) {
					r.Get("/", deps.TelemetryHandler.ListConfigs)
					r.Post("/", deps.TelemetryHandler.CreateConfig)
					r.Get("/{configID}", deps.TelemetryHandler.GetConfig)
					r.Put("/{configID}", deps.TelemetryHandler.UpdateConfig)
					r.Delete("/{configID}", deps.TelemetryHandler.DeleteConfig)
					r.Post("/{configID}/test", deps.TelemetryHandler.TestConfig)
				})

				// Manual export (for testing)
				r.Post("/spans", deps.TelemetryHandler.ExportSpan)
				r.Post("/metrics", deps.TelemetryHandler.ExportMetric)
			})
		}

		// Tool Approvals - public for demo
		if deps.ApprovalHandler != nil {
			r.Route("/approvals", func(r chi.Router) {
				// Approval requests
				r.Get("/", deps.ApprovalHandler.ListApprovals)
				r.Post("/", deps.ApprovalHandler.RequestApproval)
				r.Get("/pending-count", deps.ApprovalHandler.GetPendingCount)
				r.Get("/{approvalID}", deps.ApprovalHandler.GetApproval)
				r.Post("/{approvalID}/approve", deps.ApprovalHandler.ApproveRequest)
				r.Post("/{approvalID}/deny", deps.ApprovalHandler.DenyRequest)

				// Access check
				r.Get("/check-access", deps.ApprovalHandler.CheckAccess)
			})

			r.Route("/tool-classifications", func(r chi.Router) {
				r.Get("/", deps.ApprovalHandler.ListClassifications)
				r.Post("/", deps.ApprovalHandler.SetClassification)
				r.Get("/{server}/{tool}", deps.ApprovalHandler.GetClassification)
				r.Delete("/{server}/{tool}", deps.ApprovalHandler.DeleteClassification)
			})

			r.Route("/tool-permissions", func(r chi.Router) {
				r.Get("/", deps.ApprovalHandler.ListPermissions)
				r.Post("/", deps.ApprovalHandler.GrantPermission)
				r.Delete("/{permissionID}", deps.ApprovalHandler.RevokePermission)
			})
		}

		// RBAC - Role-Based Access Control - public for demo
		if deps.RBACHandler != nil {
			r.Route("/rbac", func(r chi.Router) {
				// Permissions info
				r.Get("/permissions", deps.RBACHandler.ListPermissions)
				r.Get("/permissions/check", deps.RBACHandler.CheckPermission)
				r.Get("/me", deps.RBACHandler.GetMyPermissions)

				// Roles
				r.Route("/roles", func(r chi.Router) {
					r.Get("/", deps.RBACHandler.ListRoles)
					r.Post("/", deps.RBACHandler.CreateRole)
					r.Get("/{roleID}", deps.RBACHandler.GetRole)
					r.Put("/{roleID}", deps.RBACHandler.UpdateRole)
					r.Delete("/{roleID}", deps.RBACHandler.DeleteRole)
					r.Get("/{roleID}/users", deps.RBACHandler.GetRoleUsers)
				})

				// User role assignments
				r.Route("/users/{userID}/roles", func(r chi.Router) {
					r.Get("/", deps.RBACHandler.GetUserRoles)
					r.Post("/", deps.RBACHandler.AssignRole)
					r.Delete("/{assignmentID}", deps.RBACHandler.RevokeRole)
				})
			})
		}

		// SSO Provider Management - public for demo
		if deps.SSOHandler != nil {
			r.Route("/sso", func(r chi.Router) {
				// Provider info
				r.Get("/providers/supported", deps.SSOHandler.GetSupportedProviders)
				r.Get("/stats", deps.SSOHandler.GetStats)

				// Provider management
				r.Route("/providers", func(r chi.Router) {
					r.Get("/", deps.SSOHandler.ListProviders)
					r.Post("/", deps.SSOHandler.CreateProvider)
					r.Get("/{providerID}", deps.SSOHandler.GetProvider)
					r.Put("/{providerID}", deps.SSOHandler.UpdateProvider)
					r.Delete("/{providerID}", deps.SSOHandler.DeleteProvider)
					r.Post("/{providerID}/test", deps.SSOHandler.TestConnection)
				})

				// Session management
				r.Route("/sessions", func(r chi.Router) {
					r.Get("/", deps.SSOHandler.ListSessions)
					r.Delete("/all", deps.SSOHandler.RevokeAllSessions)
					r.Delete("/{sessionID}", deps.SSOHandler.RevokeSession)
				})
			})
		}
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
