// Package handlers — OpenAPI 3.1 spec + Swagger UI.
//
// Covers GreenMetrics-GAPS B-02. We ship a hand-maintained OpenAPI 3.1
// document as the source of truth (the alternative, full `swaggo/swag`
// annotations on every handler, is a much larger refactor). The document
// lives in docs/openapi.json at the repo root and is embedded here.
package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// OpenAPIHandler serves the bundled spec.
func OpenAPIHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.SendString(openAPISpec)
}

// SwaggerUIHandler serves a minimal Swagger UI shell that loads /api/docs/openapi.json.
func SwaggerUIHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(swaggerUIHTML)
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>GreenMetrics API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/api/docs/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
      });
    };
  </script>
</body>
</html>`

// openAPISpec is the full OpenAPI 3.1 document. Kept as a string for single-
// binary deploys (no filesystem dependency). Update as routes change.
const openAPISpec = `{
  "openapi": "3.1.0",
  "info": {
    "title": "GreenMetrics API",
    "version": "1.0.0",
    "description": "Energy monitoring + carbon accounting API. CSRD/ESRS E1 + Piano 5.0.",
    "contact": {"name": "GreenMetrics Support", "email": "support@greenmetrics.it"},
    "license": {"name": "Proprietary"}
  },
  "servers": [
    {"url": "http://localhost:8082", "description": "Local dev"},
    {"url": "https://app.greenmetrics.it", "description": "Production"}
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
    },
    "schemas": {
      "Problem": {
        "type": "object",
        "required": ["type", "title", "status"],
        "properties": {
          "type": {"type": "string", "format": "uri"},
          "title": {"type": "string"},
          "status": {"type": "integer"},
          "detail": {"type": "string"},
          "instance": {"type": "string"},
          "code": {"type": "string"}
        }
      },
      "LoginRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": {"type": "string", "format": "email"},
          "password": {"type": "string", "minLength": 12}
        }
      },
      "LoginResponse": {
        "type": "object",
        "properties": {
          "access_token": {"type": "string"},
          "refresh_token": {"type": "string"},
          "expires_in": {"type": "integer"},
          "token_type": {"type": "string"}
        }
      },
      "Reading": {
        "type": "object",
        "required": ["ts", "meter_id", "channel_id", "value", "unit"],
        "properties": {
          "ts": {"type": "string", "format": "date-time"},
          "meter_id": {"type": "string"},
          "channel_id": {"type": "string"},
          "value": {"type": "number"},
          "unit": {"type": "string"},
          "quality_code": {"type": "integer"}
        }
      },
      "Meter": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "label": {"type": "string"},
          "meter_type": {"type": "string"},
          "protocol": {"type": "string"},
          "unit": {"type": "string"},
          "pod_code": {"type": "string"},
          "pdr_code": {"type": "string"}
        }
      },
      "HealthResponse": {
        "type": "object",
        "required": ["status", "service", "version"],
        "properties": {
          "status": {"type": "string", "enum": ["ok", "degraded", "error"]},
          "service": {"type": "string"},
          "version": {"type": "string"},
          "uptime_seconds": {"type": "integer"},
          "time": {"type": "string", "format": "date-time"},
          "dependencies": {"type": "object"}
        }
      }
    }
  },
  "paths": {
    "/api/health": {"get": {"summary": "Liveness + dependency status", "responses": {"200": {"description": "ok", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/HealthResponse"}}}}}}},
    "/api/ready": {"get": {"summary": "Readiness probe", "responses": {"200": {"description": "ready"}}}},
    "/api/live": {"get": {"summary": "Liveness probe", "responses": {"200": {"description": "alive"}}}},
    "/metrics": {"get": {"summary": "Prometheus scrape endpoint", "responses": {"200": {"description": "text/plain Prometheus exposition"}}}},
    "/api/v1/auth/login": {"post": {"summary": "Issue access/refresh JWT", "requestBody": {"required": true, "content": {"application/json": {"schema": {"$ref": "#/components/schemas/LoginRequest"}}}}, "responses": {"200": {"description": "ok", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/LoginResponse"}}}}, "401": {"description": "invalid credentials"}, "429": {"description": "rate limited or account locked"}}}},
    "/api/v1/auth/refresh": {"post": {"summary": "Refresh access token", "responses": {"200": {"description": "new access token"}}}},
    "/api/v1/auth/logout": {"post": {"summary": "Logout (client-side)", "responses": {"204": {"description": "ok"}}}},
    "/api/v1/meters": {
      "get": {"summary": "List meters", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}},
      "post": {"summary": "Create meter", "security": [{"BearerAuth": []}], "responses": {"201": {"description": "created"}}}
    },
    "/api/v1/meters/{id}": {
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "get": {"summary": "Get meter", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}},
      "put": {"summary": "Update meter", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}},
      "delete": {"summary": "Delete meter", "security": [{"BearerAuth": []}], "responses": {"204": {"description": "deleted"}}}
    },
    "/api/v1/readings": {"get": {"summary": "Raw readings", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}}},
    "/api/v1/readings/ingest": {"post": {"summary": "Bulk ingest readings", "security": [{"BearerAuth": []}], "responses": {"202": {"description": "accepted"}}}},
    "/api/v1/readings/aggregated": {"get": {"summary": "Continuous-aggregate slice (15m/1h/1d)", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}}},
    "/api/v1/readings/export": {"get": {"summary": "Streaming CSV export", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "text/csv"}}}},
    "/api/v1/reports": {
      "get": {"summary": "List reports", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}},
      "post": {"summary": "Generate ESRS E1 / Piano 5.0 report", "security": [{"BearerAuth": []}], "responses": {"201": {"description": "created"}}}
    },
    "/api/v1/reports/{id}": {
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "get": {"summary": "Get report metadata", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}}
    },
    "/api/v1/alerts": {"get": {"summary": "Open alerts", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}}},
    "/api/v1/emission-factors": {
      "get": {"summary": "List emission factors (ISPRA-seeded)", "security": [{"BearerAuth": []}], "responses": {"200": {"description": "ok"}}},
      "post": {"summary": "Create emission factor", "security": [{"BearerAuth": []}], "responses": {"201": {"description": "created"}}}
    },
    "/api/v1/pulse/ingest": {"post": {"summary": "Pulse-counter webhook (HMAC-signed)", "security": [{"BearerAuth": []}], "responses": {"202": {"description": "accepted"}, "401": {"description": "invalid signature"}}}}
  }
}`
