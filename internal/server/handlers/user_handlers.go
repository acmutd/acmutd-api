package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMe returns the authenticated user's profile.
func (h *Handler) GetMe(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	user, err := h.db.GetUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers returns all registered users. Admin only.
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.db.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// BanUser sets a user's approval_status to "banned" and deactivates their keys. Admin only.
func (h *Handler) BanUser(c *gin.Context) {
	targetUID := c.Param("uid")
	if targetUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uid is required"})
		return
	}

	if err := h.db.SetUserBanStatus(c.Request.Context(), targetUID, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ban user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user banned"})
}

// UnbanUser sets a user's approval_status back to "approved". Admin only.
func (h *Handler) UnbanUser(c *gin.Context) {
	targetUID := c.Param("uid")
	if targetUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uid is required"})
		return
	}

	if err := h.db.SetUserBanStatus(c.Request.Context(), targetUID, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unban user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unbanned"})
}
