package handlers

import (
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
)

// buildAPIKeyInput constructs a types.APIKey from raw request fields.
func buildAPIKeyInput(userID, ownerEmail, label, description string, isAdmin bool, rateLimit, windowSeconds int, expiresAt string) types.APIKey {
	key := types.APIKey{
		UserID:        userID,
		OwnerEmail:    ownerEmail,
		Label:         label,
		Description:   description,
		IsAdmin:       isAdmin,
		RateLimit:     rateLimit,
		WindowSeconds: windowSeconds,
		Status:        "active",
	}
	if expiresAt != "" {
		if t, err := time.Parse(time.RFC3339, expiresAt); err == nil {
			key.ExpiresAt = t
		}
	}
	return key
}

// GetMyAPIKey returns the API key owned by the authenticated user, or null if none.
func (h *Handler) GetMyAPIKey(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	key, err := h.db.GetUserAPIKey(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch API key"})
		return
	}

	c.JSON(http.StatusOK, key)
}

// RequestAPIKey creates a pending API key request for the authenticated user.
func (h *Handler) RequestAPIKey(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}

	var req struct {
		Label       string `json:"label" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.GetUser(c.Request.Context(), uid)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user profile"})
		return
	}

	if err := h.db.RequestAPIKey(c.Request.Context(), uid, user.Email, req.Label, req.Description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "API key request submitted"})
}

// RegenerateKey generates new key bytes for the caller's own key.
func (h *Handler) RegenerateKey(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}
	keyID := c.Param("keyId")

	existing, err := h.db.GetUserAPIKey(c.Request.Context(), uid)
	if err != nil || existing == nil || existing.KeyID != keyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "key not found or not owned by you"})
		return
	}

	updated, err := h.db.RegenerateKey(c.Request.Context(), keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to regenerate key"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// RevokeKey sets the caller's own key to "inactive".
func (h *Handler) RevokeKey(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}
	keyID := c.Param("keyId")

	existing, err := h.db.GetUserAPIKey(c.Request.Context(), uid)
	if err != nil || existing == nil || existing.KeyID != keyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "key not found or not owned by you"})
		return
	}

	if err := h.db.RevokeKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key revoked"})
}

// DeleteKey permanently deletes the caller's own key document from Firestore.
func (h *Handler) DeleteKey(c *gin.Context) {
	uid, ok := getDashboardUID(c)
	if !ok {
		return
	}
	keyID := c.Param("keyId")

	existing, err := h.db.GetUserAPIKey(c.Request.Context(), uid)
	if err != nil || existing == nil || existing.KeyID != keyID {
		c.JSON(http.StatusForbidden, gin.H{"error": "key not found or not owned by you"})
		return
	}

	if err := h.db.DeleteKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

// ListAllKeys returns all dashboard-managed API keys. Admin only.
func (h *Handler) ListAllKeys(c *gin.Context) {
	keys, err := h.db.ListAllKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

// ListPendingKeys returns all keys with status "pending". Admin only.
func (h *Handler) ListPendingKeys(c *gin.Context) {
	keys, err := h.db.ListPendingKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list pending keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

// ApproveKey sets a key's status to "active". Admin only.
func (h *Handler) ApproveKey(c *gin.Context) {
	keyID := c.Param("keyId")
	if err := h.db.ApproveKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "key approved"})
}

// RejectKey sets a key's status to "rejected". Admin only.
func (h *Handler) RejectKey(c *gin.Context) {
	keyID := c.Param("keyId")
	if err := h.db.RejectKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "key rejected"})
}

// AdminRegenerateKey regenerates any key by keyId. Admin only.
func (h *Handler) AdminRegenerateKey(c *gin.Context) {
	keyID := c.Param("keyId")
	updated, err := h.db.RegenerateKey(c.Request.Context(), keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to regenerate key"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// AdminRevokeKey sets any key's status to "inactive". Admin only.
func (h *Handler) AdminRevokeKey(c *gin.Context) {
	keyID := c.Param("keyId")
	if err := h.db.RevokeKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "key revoked"})
}

// AdminDeleteKey permanently deletes a key document from Firestore. Admin only.
func (h *Handler) AdminDeleteKey(c *gin.Context) {
	keyID := c.Param("keyId")
	if err := h.db.DeleteKey(c.Request.Context(), keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

// AddKey creates a new API key with all fields specified. Admin only.
func (h *Handler) AddKey(c *gin.Context) {
	var req struct {
		OwnerEmail    string `json:"ownerEmail" binding:"required"`
		Label         string `json:"label" binding:"required"`
		Description   string `json:"description"`
		IsAdmin       bool   `json:"isAdmin"`
		RateLimit     int    `json:"rateLimit" binding:"required"`
		WindowSeconds int    `json:"windowSeconds" binding:"required"`
		ExpiresAt     string `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.db.GetUserByEmail(c.Request.Context(), req.OwnerEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up owner"})
		return
	}

	userID := ""
	if user != nil {
		userID = user.UID
	}

	input := buildAPIKeyInput(userID, req.OwnerEmail, req.Label, req.Description, req.IsAdmin, req.RateLimit, req.WindowSeconds, req.ExpiresAt)

	created, err := h.db.AddKey(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// EditKey updates mutable fields on an existing key. Admin only.
func (h *Handler) EditKey(c *gin.Context) {
	keyID := c.Param("keyId")

	var req struct {
		Label         *string `json:"label"`
		Description   *string `json:"description"`
		IsAdmin       *bool   `json:"isAdmin"`
		RateLimit     *int    `json:"rateLimit"`
		WindowSeconds *int    `json:"windowSeconds"`
		ExpiresAt     *string `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]any)
	if req.Label != nil {
		updates["label"] = *req.Label
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}
	if req.RateLimit != nil {
		updates["rate_limit"] = *req.RateLimit
	}
	if req.WindowSeconds != nil {
		updates["window_seconds"] = *req.WindowSeconds
	}
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expiresAt format, expected RFC3339"})
			return
		}
		updates["expires_at"] = t
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	updated, err := h.db.UpdateKey(c.Request.Context(), keyID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update key"})
		return
	}

	c.JSON(http.StatusOK, updated)
}
