package handler

import (
	"net/http"

	"github.com/rs/zerolog"
)

// DocsHandler handles API documentation requests.
type DocsHandler struct {
	logger  zerolog.Logger
	openAPI []byte
}

// NewDocsHandler creates a new docs handler.
func NewDocsHandler(logger zerolog.Logger, openAPISpec []byte) *DocsHandler {
	return &DocsHandler{
		logger:  logger,
		openAPI: openAPISpec,
	}
}

// SwaggerUI serves the Swagger UI page.
func (h *DocsHandler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swaggerUIHTML))
}

// OpenAPISpec serves the OpenAPI specification.
func (h *DocsHandler) OpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(h.openAPI)
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GatewayOps API Documentation</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
        .swagger-ui .topbar {
            display: none;
        }
        .swagger-ui .info {
            margin: 20px 0;
        }
        .swagger-ui .info .title {
            font-size: 36px;
            color: #3b4151;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px 40px;
            display: flex;
            align-items: center;
            gap: 16px;
        }
        .header h1 {
            margin: 0;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            font-size: 24px;
            font-weight: 600;
        }
        .header .badge {
            background: rgba(255,255,255,0.2);
            padding: 4px 12px;
            border-radius: 16px;
            font-size: 12px;
            font-weight: 500;
        }
        #swagger-ui {
            max-width: 1460px;
            margin: 0 auto;
            padding: 0 20px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>GatewayOps API</h1>
        <span class="badge">v1.0.0</span>
    </div>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/openapi.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                defaultModelsExpandDepth: 0,
                docExpansion: "list",
                filter: true,
                tagsSorter: "alpha",
                operationsSorter: "alpha"
            });
        };
    </script>
</body>
</html>`
