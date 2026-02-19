package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *firebase.Firestore
}

const (
	defaultLimit = 100
	maxLimit     = 100
)

type paginationParams struct {
	Limit  int
	Page   int
	Offset int
}

func New(db *firebase.Firestore) *Handler {
	return &Handler{db: db}
}

// Health responds with a simple service heartbeat.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "ACM API is running",
	})
}

// CreateAPIKey provisions a new API key.
func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req struct {
		RateLimit     int    `json:"rate_limit" binding:"required"`
		WindowSeconds int    `json:"window_seconds" binding:"required"`
		IsAdmin       bool   `json:"is_admin"`
		ExpiresAt     string `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RateLimit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rate limit must be greater than 0"})
		return
	}

	if req.WindowSeconds <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "window seconds must be greater than 0"})
		return
	}

	var expiresAt time.Time
	if req.ExpiresAt != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at format"})
			return
		}

		if expiresAt.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expiration date must be in the future"})
			return
		}
	}

	key, err := h.db.GenerateAPIKey(
		c.Request.Context(),
		req.RateLimit,
		req.WindowSeconds,
		req.IsAdmin,
		expiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"key": key})
}

// GetAPIKey retrieves metadata for a stored API key.
func (h *Handler) GetAPIKey(c *gin.Context) {
	key := c.Param("key")

	apiKey, err := h.db.GetAPIKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get API key"})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

func normalizeTerm(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizePrefix(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeCourseNumber(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeSchool(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func parsePaginationParams(c *gin.Context) (paginationParams, error) {
	limitValue := strings.TrimSpace(c.Query("limit"))
	if limitValue == "" {
		limitValue = strconv.Itoa(defaultLimit)
	}

	limit, err := strconv.Atoi(limitValue)
	if err != nil || limit <= 0 {
		return paginationParams{}, fmt.Errorf("limit parameter must be a positive integer")
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	pageValue := strings.TrimSpace(c.Query("page"))
	if pageValue == "" {
		pageValue = "1"
	}

	page, err := strconv.Atoi(pageValue)
	if err != nil || page <= 0 {
		return paginationParams{}, fmt.Errorf("page parameter must be a positive integer")
	}

	offset := (page - 1) * limit

	return paginationParams{
		Limit:  limit,
		Page:   page,
		Offset: offset,
	}, nil
}

func parsePaginationOrRespond(c *gin.Context) (paginationParams, bool) {
	params, err := parsePaginationParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return paginationParams{}, false
	}
	return params, true
}

func buildPaginationMeta(params paginationParams, itemsReturned int, hasNext bool) gin.H {
	meta := gin.H{
		"page":     params.Page,
		"limit":    params.Limit,
		"has_next": hasNext,
	}

	if hasNext {
		meta["next_page"] = params.Page + 1
	} else {
		meta["total"] = params.Offset + itemsReturned
	}

	return meta
}
