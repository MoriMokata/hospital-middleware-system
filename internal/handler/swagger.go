package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerUIPage = `<!doctype html>
<html>
<head>
  <title>Hospital Middleware API — Docs</title>
  <meta charset="utf-8"/>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>`

// SwaggerUI serves a Swagger UI page (loaded from a CDN) pointed at the
// OpenAPI spec served by OpenAPISpec. Not part of the assignment's
// required scope — added on request; see docs/design.md.
func SwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIPage))
}

// OpenAPISpec serves the given OpenAPI 3.0 YAML spec bytes (embedded from
// api/openapi.yaml — see api/openapi.go).
func OpenAPISpec(spec []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", spec)
	}
}
