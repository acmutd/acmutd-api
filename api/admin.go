package api

import (
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/types"
	"github.com/gin-gonic/gin"
)

func (api *API) SetupAdminRoutes() {
	admin := api.router.Group("/admin")
	admin.Use(api.AdminAuthMiddleware())
	{
		keys := admin.Group("/keys")
		{
			keys.POST("/", api.createAPIKey)
			keys.GET("/", api.getAllAPIKeys)
			keys.GET("/:id", api.getAPIKey)
			keys.PUT("/:id", api.updateAPIKey)
			keys.DELETE("/:id", api.deleteAPIKey)
		}
	}
}

func (api *API) createAPIKey(c *gin.Context) {
	var req types.APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := GenerateAPIKeyID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key ID"})
		return
	}

	key, err := GenerateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key"})
		return
	}

	now := time.Now()
	rateInterval := req.RateInterval.ToDuration()
	apiKey := &types.APIKey{
		ID:               id,
		Key:              key,
		ExpiresAt:        time.Time{}, // Zero time means no expiration
		LastUsedAt:       time.Time{},
		UsageCount:       0,
		RateLimit:        req.RateLimit,
		RateInterval:     rateInterval,
		CurrentWindow:    now.Add(rateInterval),
		RequestsInWindow: 0,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if req.ExpiresAt != nil {
		apiKey.ExpiresAt = *req.ExpiresAt
	}

	if err := api.db.CreateAPIKey(c.Request.Context(), apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	response := types.APIKeyResponse{
		ID:           apiKey.ID,
		Key:          apiKey.Key,
		ExpiresAt:    apiKey.ExpiresAt,
		RateLimit:    apiKey.RateLimit,
		RateInterval: apiKey.RateInterval.String(),
		IsActive:     apiKey.IsActive,
		CreatedAt:    apiKey.CreatedAt,
		UsageCount:   apiKey.UsageCount,
		LastUsedAt:   apiKey.LastUsedAt,
	}

	c.JSON(http.StatusCreated, response)
}

func (api *API) getAllAPIKeys(c *gin.Context) {
	apiKeys, err := api.db.GetAllAPIKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve API keys"})
		return
	}

	var responses []types.APIKeyResponse
	for _, key := range apiKeys {
		response := types.APIKeyResponse{
			ID:           key.ID,
			Key:          "<hidden>",
			ExpiresAt:    key.ExpiresAt,
			RateLimit:    key.RateLimit,
			RateInterval: key.RateInterval.String(),
			IsActive:     key.IsActive,
			CreatedAt:    key.CreatedAt,
			UsageCount:   key.UsageCount,
			LastUsedAt:   key.LastUsedAt,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(responses),
		"keys":  responses,
	})
}

func (api *API) getAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	apiKey, err := api.db.GetAPIKeyByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	response := types.APIKeyResponse{
		ID:           apiKey.ID,
		Key:          "<hidden>",
		ExpiresAt:    apiKey.ExpiresAt,
		RateLimit:    apiKey.RateLimit,
		RateInterval: apiKey.RateInterval.String(),
		IsActive:     apiKey.IsActive,
		CreatedAt:    apiKey.CreatedAt,
		UsageCount:   apiKey.UsageCount,
		LastUsedAt:   apiKey.LastUsedAt,
	}

	c.JSON(http.StatusOK, response)
}

func (api *API) updateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	var req types.APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing API key
	apiKey, err := api.db.GetAPIKeyByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	// Update fields
	apiKey.RateLimit = req.RateLimit
	apiKey.RateInterval = req.RateInterval.ToDuration()
	apiKey.UpdatedAt = time.Now()

	if req.ExpiresAt != nil {
		apiKey.ExpiresAt = *req.ExpiresAt
	}

	// Save to database
	if err := api.db.UpdateAPIKey(c.Request.Context(), apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key"})
		return
	}

	response := types.APIKeyResponse{
		ID:           apiKey.ID,
		Key:          "<hidden>",
		ExpiresAt:    apiKey.ExpiresAt,
		RateLimit:    apiKey.RateLimit,
		RateInterval: apiKey.RateInterval.String(),
		IsActive:     apiKey.IsActive,
		CreatedAt:    apiKey.CreatedAt,
		UsageCount:   apiKey.UsageCount,
		LastUsedAt:   apiKey.LastUsedAt,
	}

	c.JSON(http.StatusOK, response)
}

func (api *API) deleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	_, err := api.db.GetAPIKeyByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	if err := api.db.DeleteAPIKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}
