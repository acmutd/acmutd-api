package handlers

import (
	"net/http"
	"strings"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
)

// GetAllCourses returns a helpful error since callers must provide a term.
func (h *Handler) GetAllCourses(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Term parameter is required. Use /api/v1/courses/{term}",
	})
}

// GetCoursesByTerm fetches courses within a term.
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

	var (
		courses []types.Course
		hasNext bool
		err     error
	)

	courses, hasNext, err = h.db.GetAllCoursesByTerm(c.Request.Context(), term, params.Limit, params.Offset)

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

// GetCourseBySection fetches a specific course by term, prefix, number, and section.
func (h *Handler) GetCourseBySection(c *gin.Context) {
	term := normalizeTerm(c.Param("term"))
	prefix := normalizePrefix(c.Param("prefix"))
	number := normalizeCourseNumber(c.Param("number"))
	section := strings.TrimSpace(c.Param("section"))

	if term == "" || prefix == "" || number == "" || section == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term, prefix, number, and section parameters are required"})
		return
	}

	course, err := h.db.GetCourseBySection(c.Request.Context(), term, prefix, number, section)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"course": course,
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
