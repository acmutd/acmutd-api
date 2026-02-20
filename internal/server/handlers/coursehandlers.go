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
// Query parameters: prefix, number, section, school, q (all optional, can be combined)
//
// All filters are independent and can be used in any combination:
//   - prefix: filters by course prefix (e.g., "cs")
//   - number: filters by course number (e.g., "1337") - works with or without prefix
//   - section: filters by section (e.g., "001") - works with or without prefix/number
//   - school: filters by school code (e.g., "ecs", "nsm")
//   - q: search query for title, topic, or instructor name
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
	search := strings.TrimSpace(c.Query("q"))

	// For all other queries, parse pagination
	params, ok := parsePaginationOrRespond(c)
	if !ok {
		return
	}

	// Build query with all parameters
	query := types.CourseQuery{
		Term:         term,
		CoursePrefix: prefix,
		CourseNumber: number,
		Section:      section,
		School:       school,
		Search:       search,
		Limit:        params.Limit,
		Offset:       params.Offset,
	}

	// Build response metadata based on filters used
	responseMeta := gin.H{"term": term}

	// Handle specific section lookup (returns single course only when prefix+number+section all provided)
	if section != "" && prefix != "" && number != "" {
		courses, _, err := h.db.QueryCourses(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var course *types.Course
		if len(courses) > 0 {
			course = &courses[0]
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

	courses, hasNext, err := h.db.QueryCourses(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add filter info to response
	if prefix != "" {
		responseMeta["prefix"] = prefix
	}
	if number != "" {
		responseMeta["number"] = number
	}
	if section != "" {
		responseMeta["section"] = section
	}
	if school != "" {
		responseMeta["school"] = school
	}
	if search != "" {
		responseMeta["query"] = search
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	responseMeta["count"] = len(courses)
	responseMeta["courses"] = courses
	responseMeta["pagination"] = pagination

	c.JSON(http.StatusOK, responseMeta)
}
