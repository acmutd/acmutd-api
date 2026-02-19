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

// GetCourses fetches courses with optional filtering by query parameters.
// Path parameter: term (required)
// Query parameters: prefix, number, section, school (all optional)
// Filtering priority:
//   - section: requires prefix and number, returns single course
//   - number: requires prefix, filters by course number
//   - prefix: filters by course prefix
//   - school: filters by school
//   - none: returns all courses for term
func (h *Handler) GetCourses(c *gin.Context) {
	term := normalizeTerm(c.Param("term"))
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
		return
	}

	// Parse optional query parameters
	prefix := normalizePrefix(c.Query("prefix"))
	number := normalizeCourseNumber(c.Query("number"))
	section := normalizeSection(c.Query("section"))
	school := normalizeSchool(c.Query("school"))

	// Validate parameter dependencies
	if section != "" && (prefix == "" || number == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Section parameter requires both prefix and number parameters"})
		return
	}
	if number != "" && prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number parameter requires prefix parameter"})
		return
	}

	// Handle section lookup (returns single course, no pagination)
	if section != "" {
		course, err := h.db.GetCourseBySection(c.Request.Context(), term, prefix, number, section)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"term":    term,
			"prefix":  prefix,
			"number":  number,
			"section": section,
			"course":  course,
		})
		return
	}

	// For all other queries, parse pagination
	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	var (
		courses []types.Course
		hasNext bool
		err     error
	)

	// Build response metadata based on filters used
	responseMeta := gin.H{"term": term}

	switch {
	case number != "":
		// Filter by prefix and number
		courses, hasNext, err = h.db.QueryByCourseNumber(c.Request.Context(), term, prefix, number, params.Limit, params.Offset)
		responseMeta["prefix"] = prefix
		responseMeta["number"] = number
	case prefix != "":
		// Filter by prefix only
		courses, hasNext, err = h.db.QueryByCoursePrefix(c.Request.Context(), term, prefix, params.Limit, params.Offset)
		responseMeta["prefix"] = prefix
	case school != "":
		// Filter by school
		courses, hasNext, err = h.db.QueryBySchool(c.Request.Context(), term, school, params.Limit, params.Offset)
		responseMeta["school"] = school
	default:
		// No filters, return all courses for term
		courses, hasNext, err = h.db.GetAllCoursesByTerm(c.Request.Context(), term, params.Limit, params.Offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	responseMeta["count"] = len(courses)
	responseMeta["courses"] = courses
	responseMeta["pagination"] = pagination

	c.JSON(http.StatusOK, responseMeta)
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
