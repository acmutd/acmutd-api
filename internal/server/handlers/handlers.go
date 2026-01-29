package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/acmutd/acmutd-api/internal/firebase"
	"github.com/acmutd/acmutd-api/internal/types"
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

// GetAllCourses returns a helpful error since callers must provide a term.
func (h *Handler) GetAllCourses(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Term parameter is required. Use /api/v1/courses/{term}",
	})
}

// GetCoursesByTerm fetches courses and applies optional prefix/number filters.
func (h *Handler) GetCoursesByTerm(c *gin.Context) {
	term := normalizeTerm(c.Param("term"))
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	prefix := normalizePrefix(c.Query("prefix"))
	number := normalizeCourseNumber(c.Query("number"))

	var (
		courses []types.Course
		hasNext bool
		err     error
	)

	switch {
	case prefix != "" && number != "":
		courses, hasNext, err = h.db.QueryByCourseNumber(c.Request.Context(), term, prefix, number, params.Limit, params.Offset)
	case prefix != "":
		courses, hasNext, err = h.db.QueryByCoursePrefix(c.Request.Context(), term, prefix, params.Limit, params.Offset)
	default:
		courses, hasNext, err = h.db.GetAllCoursesByTerm(c.Request.Context(), term, params.Limit, params.Offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"term":       term,
		"count":      len(courses),
		"courses":    courses,
		"pagination": pagination,
	})
}

// GetCoursesByPrefix fetches courses by prefix within a term.
func (h *Handler) GetCoursesByPrefix(c *gin.Context) {
	term := normalizeTerm(c.Param("term"))
	prefix := normalizePrefix(c.Param("prefix"))

	if term == "" || prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term and prefix parameters are required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	courses, hasNext, err := h.db.QueryByCoursePrefix(c.Request.Context(), term, prefix, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"term":       term,
		"prefix":     prefix,
		"count":      len(courses),
		"courses":    courses,
		"pagination": pagination,
	})
}

// GetCoursesByNumber fetches courses by prefix and number in a term.
func (h *Handler) GetCoursesByNumber(c *gin.Context) {
	term := normalizeTerm(c.Param("term"))
	prefix := normalizePrefix(c.Param("prefix"))
	number := normalizeCourseNumber(c.Param("number"))

	if term == "" || prefix == "" || number == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term, prefix, and number parameters are required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	courses, hasNext, err := h.db.QueryByCourseNumber(c.Request.Context(), term, prefix, number, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"term":       term,
		"prefix":     prefix,
		"number":     number,
		"count":      len(courses),
		"courses":    courses,
		"pagination": pagination,
	})
}

// SearchCourses runs a text search against courses for a term.
func (h *Handler) SearchCourses(c *gin.Context) {
	term := normalizeTerm(c.Param("term"))
	query := strings.TrimSpace(c.Query("q"))

	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
		return
	}

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query parameter 'q' is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	courses, hasNext, err := h.db.SearchCourses(c.Request.Context(), term, query, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"term":       term,
		"query":      query,
		"count":      len(courses),
		"courses":    courses,
		"pagination": pagination,
	})
}

// GetTerms returns all known terms.
func (h *Handler) GetTerms(c *gin.Context) {
	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	terms, hasNext, err := h.db.QueryAllTerms(c.Request.Context(), params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(terms), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(terms),
		"terms":      terms,
		"pagination": pagination,
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

// GetProfessorByID loads a professor by ID.
func (h *Handler) GetProfessorByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor ID is required"})
		return
	}

	professor, err := h.db.GetProfessorById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get professor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"professor": professor,
	})
}

// GetProfessorsByName loads professors by name.
func (h *Handler) GetProfessorsByName(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor name is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	professors, hasNext, err := h.db.GetProfessorsByName(c.Request.Context(), name, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get professors"})
		return
	}

	pagination := buildPaginationMeta(params, len(professors), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(professors),
		"professors": professors,
		"pagination": pagination,
	})
}

// GetGradesByProfID loads grade distributions by professor ID.
func (h *Handler) GetGradesByProfID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor ID is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByProfId(c.Request.Context(), id, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByProfName loads grade distributions by professor name.
func (h *Handler) GetGradesByProfName(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Professor name is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByProfName(c.Request.Context(), name, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByPrefix loads grade distributions by prefix.
func (h *Handler) GetGradesByPrefix(c *gin.Context) {
	prefix := c.Param("prefix")

	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix is required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByPrefix(c.Request.Context(), prefix, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByPrefixAndNumber loads grade distributions by prefix and course number.
func (h *Handler) GetGradesByPrefixAndNumber(c *gin.Context) {
	prefix := c.Param("prefix")
	number := c.Param("number")

	if prefix == "" || number == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix and number are required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByPrefixAndNumber(c.Request.Context(), prefix, number, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByPrefixAndTerm loads grade distributions by prefix and term.
func (h *Handler) GetGradesByPrefixAndTerm(c *gin.Context) {
	prefix := c.Param("prefix")
	term := c.Param("term")

	if prefix == "" || term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prefix and term are required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByPrefixAndTerm(c.Request.Context(), prefix, term, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesByNumberAndTerm loads grade distributions by course number in a specific term
func (h *Handler) GetGradesByNumberAndTerm(c *gin.Context) {
	term := c.Param("term")
	prefix := c.Param("prefix")
	number := c.Param("number")

	if term == "" || prefix == "" || number == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term, prefix, and number are required"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	grades, hasNext, err := h.db.GetGradesByNumberAndTerm(c.Request.Context(), term, prefix, number, params.Limit, params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	pagination := buildPaginationMeta(params, len(grades), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(grades),
		"grades":     grades,
		"pagination": pagination,
	})
}

// GetGradesBySection loads grade distributions for specific section
func (h *Handler) GetGradesBySection(c *gin.Context) {
	term := c.Param("term")
	prefix := c.Param("prefix")
	number := c.Param("number")
	section := c.Param("section")

	if term == "" || prefix == "" || number == "" || section == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term, prefix, number, and section are required"})
		return
	}

	grades, err := h.db.GetGradesBySection(c.Request.Context(), term, prefix, number, section)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get grades"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grades": grades,
	})
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
