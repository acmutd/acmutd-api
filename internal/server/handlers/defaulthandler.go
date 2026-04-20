package handlers

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed swagger.html
var swaggerUIPage []byte

// Default serves the embedded Swagger UI page.
func (h *Handler) Default(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUIPage)
}

// OpenAPISpec serves the OpenAPI document consumed by Swagger UI.
func (h *Handler) OpenAPISpec(c *gin.Context) {
	specPath := openAPISpecPath()
	content, err := os.ReadFile(specPath)
	if err != nil {
		c.Data(http.StatusServiceUnavailable, "text/plain; charset=utf-8", []byte(fmt.Sprintf("OpenAPI spec is missing. Expected file: %s", specPath)))
		return
	}

	c.Data(http.StatusOK, "application/yaml; charset=utf-8", content)
}

func openAPISpecPath() string {
	if custom := strings.TrimSpace(os.Getenv("OPENAPI_SPEC_PATH")); custom != "" {
		return custom
	}

	return filepath.Join("internal", "server", "docs", "openapi.yaml")
}
