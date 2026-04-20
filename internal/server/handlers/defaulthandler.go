package handlers

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed swagger.html
var swaggerUIPage []byte

//go:embed openapi.yaml
var openAPISpec []byte

// Default serves the embedded Swagger UI page.
func (h *Handler) Default(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUIPage)
}

// OpenAPISpec serves the OpenAPI document consumed by Swagger UI.
func (h *Handler) OpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
}
