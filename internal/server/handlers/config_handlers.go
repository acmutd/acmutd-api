package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAppConfig returns the current application configuration.
func (h *Handler) GetAppConfig(c *gin.Context) {
	cfg, err := h.db.GetAppConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch config"})
		return
	}
	if cfg == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// UpdateAppConfig applies a partial update to the application configuration. Admin only.
// Only the five user-editable fields are accepted; all other fields (e.g. updatedAt,
// updatedBy sent back from the frontend) are ignored so they can't corrupt Firestore types.
func (h *Handler) UpdateAppConfig(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	var req struct {
		AutoApproveMode      *string `json:"autoApproveMode"`
		HackathonModeEnabled *bool   `json:"hackathonModeEnabled"`
		HackathonEndDate     *string `json:"hackathonEndDate"`
		CurrentSemester      *string `json:"currentSemester"`
		InstanceType         *string `json:"instanceType"`
		KeysExpiresAtDate    *string `json:"keysExpiresAtDate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]any)
	if req.AutoApproveMode != nil {
		updates["auto_approve_mode"] = *req.AutoApproveMode
	}
	if req.HackathonModeEnabled != nil {
		updates["hackathon_mode_enabled"] = *req.HackathonModeEnabled
	}
	if req.HackathonEndDate != nil && *req.HackathonEndDate != "" {
		if _, err := time.Parse(time.RFC3339, *req.HackathonEndDate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hackathonEndDate format, expected RFC3339"})
			return
		}
		updates["hackathon_end_date"] = *req.HackathonEndDate
	}
	if req.CurrentSemester != nil {
		updates["current_semester"] = *req.CurrentSemester
	}
	if req.InstanceType != nil {
		updates["instance_type"] = *req.InstanceType
	}
	if req.KeysExpiresAtDate != nil && *req.KeysExpiresAtDate != "" {
		t, err := time.Parse(time.RFC3339, *req.KeysExpiresAtDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid keysExpiresAtDate format, expected RFC3339"})
			return
		}
		updates["keys_expires_at_date"] = t
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields to update"})
		return
	}

	updates["updated_by"] = uid

	cfg, err := h.db.UpdateAppConfig(c.Request.Context(), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update config"})
		return
	}

	c.JSON(http.StatusOK, cfg)
}
