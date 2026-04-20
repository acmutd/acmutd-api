package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	db         *firebase.Firestore
	trackStats *atomic.Bool
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

func New(db *firebase.Firestore, trackStats *atomic.Bool) *Handler {
	return &Handler{db: db, trackStats: trackStats}
}

// Health responds with a simple service heartbeat.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "ACM API is running",
	})
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

func getQueryParam(c *gin.Context, name string) string {
	value := strings.ToLower(strings.TrimSpace(c.Query(name)))
	return value
}

// getDashboardUID extracts the authenticated user's UID from the gin context.
// Set by the FirebaseAuth middleware. Responds 401 and returns false if missing.
func getDashboardUID(c *gin.Context) (string, bool) {
	val, exists := c.Get("dash_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return "", false
	}
	uid, ok := val.(string)
	if !ok || uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return "", false
	}
	return uid, true
}
