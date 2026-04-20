package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMyDailyStats returns daily usage statistics for the authenticated user.
func (h *Handler) GetMyDailyStats(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	stats, err := h.db.GetDailyStats(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetMyRecentRequests returns recent request events for the authenticated user.
func (h *Handler) GetMyRecentRequests(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	events, err := h.db.GetRecentRequests(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": events})
}

// ListAllDailyStats returns all daily usage stats across all users. Admin only.
func (h *Handler) ListAllDailyStats(c *gin.Context) {
	stats, err := h.db.ListAllDailyStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// ListAllRecentRequests returns all recent request events across all users. Admin only.
func (h *Handler) ListAllRecentRequests(c *gin.Context) {
	events, err := h.db.ListAllRecentRequests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": events})
}
