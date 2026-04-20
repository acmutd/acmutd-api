package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UserDashboardPage(c *gin.Context) {
	h.serveDashboardAppPage(c)
}

func (h *Handler) AdminDashboardPage(c *gin.Context) {
	h.serveDashboardAppPage(c)
}

func (h *Handler) serveDashboardAppPage(c *gin.Context) {
	htmlPath := filepath.Join(frontendDistDir(), "index.html")
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		c.Data(http.StatusServiceUnavailable, "text/plain; charset=utf-8", []byte(fmt.Sprintf("Dashboard assets are not built yet. Run: cd frontend && npm install && npm run build\nMissing file: %s", htmlPath)))
		return
	}

	firebaseCfg := fmt.Sprintf(
		`{"apiKey":%q,"authDomain":%q,"projectId":%q}`,
		os.Getenv("FIREBASE_WEB_API_KEY"),
		os.Getenv("FIREBASE_AUTH_DOMAIN"),
		os.Getenv("FIREBASE_PROJECT_ID"),
	)
	// Replace the quoted placeholder so the JS receives a raw object, not a string
	html := strings.ReplaceAll(string(content), `"__FIREBASE_CONFIG__"`, firebaseCfg)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func frontendDistDir() string {
	if custom := strings.TrimSpace(os.Getenv("FRONTEND_DIST_DIR")); custom != "" {
		return custom
	}
	return filepath.Join("frontend", "dist")
}
