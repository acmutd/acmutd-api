package handlers

import (
	"net/http"
	"strings"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
)

// GetSections queries section documents for a given term with optional filters.
//
// Required query parameters:
//   - term: the term to query (e.g., "25s", "24f")
//
// Optional query parameters:
//   - prefix: filter by course prefix (e.g., "cs")
//   - number: filter by course number (e.g., "2305")
//   - instructor: exact match on an element of the instructors array (e.g., "John Smith")
//   - days: exact match on an element of the days array (e.g., "Monday")
//   - building: exact match on building (e.g., "ECSS")
//   - title: substring match on title (e.g., "data structures")
//   - limit, page: pagination
func (h *Handler) GetSections(c *gin.Context) {
	term := strings.ToLower(strings.TrimSpace(c.Query("term")))
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "term query parameter is required (e.g., ?term=25s)"})
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	query := types.SectionQuery{
		Term:       term,
		Prefix:     strings.ToLower(strings.TrimSpace(c.Query("prefix"))),
		Number:     strings.ToLower(strings.TrimSpace(c.Query("number"))),
		Instructor: strings.ToLower(strings.TrimSpace(c.Query("instructor"))),
		Days:       strings.TrimSpace(c.Query("days")),
		Building:   strings.ToUpper(strings.TrimSpace(c.Query("building"))),
		Title:      strings.ToLower(strings.TrimSpace(c.Query("title"))),
		Limit:      params.Limit,
		Offset:     params.Offset,
	}

	sections, hasNext, err := h.db.QuerySections(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(sections), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(sections),
		"sections":   sections,
		"pagination": pagination,
	})
}

// GetSectionByParams retrieves a single section
func (h *Handler) GetSectionByParams(c *gin.Context) {
	prefix := strings.ToLower(strings.TrimSpace(c.Param("prefix")))
	number := strings.ToLower(strings.TrimSpace(c.Param("number")))
	section := strings.ToLower(strings.TrimSpace(c.Param("section")))
	term := strings.ToLower(strings.TrimSpace(c.Param("term")))

	result, err := h.db.GetSectionByParams(c.Request.Context(), prefix, number, term, section)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "section not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGeneralCourses queries CourseGeneralInfo documents.
// If both prefix and number are provided, returns a single course directly.
//
// Required query parameters:
//   - prefix: course prefix (e.g., "cs")
//
// Optional query parameters:
//   - number: course number (e.g., "2305"); if provided alongside prefix, returns a single course
//   - q: substring search on title or description
//   - limit, page: pagination
func (h *Handler) GetGeneralCourses(c *gin.Context) {
	prefix := strings.ToLower(strings.TrimSpace(c.Query("prefix")))
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prefix query parameter is required"})
		return
	}

	number := strings.ToLower(strings.TrimSpace(c.Query("number")))
	if number != "" {
		course, err := h.db.GetGeneralCourse(c.Request.Context(), prefix, number)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if course == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		c.JSON(http.StatusOK, course)
		return
	}

	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	query := types.GeneralCourseQuery{
		Prefix: prefix,
		Search: strings.TrimSpace(c.Query("q")),
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	courses, hasNext, err := h.db.QueryGeneralCourses(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(courses),
		"courses":    courses,
		"pagination": pagination,
	})
}

// GetGeneralCourse retrieves a single CourseGeneralInfo by prefix and number.
func (h *Handler) GetGeneralCourse(c *gin.Context) {
	prefix := strings.ToLower(strings.TrimSpace(c.Param("prefix")))
	number := strings.ToLower(strings.TrimSpace(c.Param("number")))

	course, err := h.db.GetGeneralCourse(c.Request.Context(), prefix, number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if course == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}

	c.JSON(http.StatusOK, course)
}
