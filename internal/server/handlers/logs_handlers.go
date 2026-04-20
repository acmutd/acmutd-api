package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListScraperLogs returns all scraper pipeline logs ordered by started_at desc. Admin only.
func (h *Handler) ListScraperLogs(c *gin.Context) {
	logs, err := h.db.ListScraperLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch scraper logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// ListCronLogs returns all cron execution logs ordered by run_at desc. Admin only.
func (h *Handler) ListCronLogs(c *gin.Context) {
	logs, err := h.db.ListCronLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch cron logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// ListServerStateLogs returns all server state change logs ordered by timestamp desc. Admin only.
func (h *Handler) ListServerStateLogs(c *gin.Context) {
	logs, err := h.db.ListServerStateLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch server state logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// ListPromotionLogs returns all data promotion logs ordered by promoted_at desc. Admin only.
func (h *Handler) ListPromotionLogs(c *gin.Context) {
	logs, err := h.db.ListPromotionLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch promotion logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
