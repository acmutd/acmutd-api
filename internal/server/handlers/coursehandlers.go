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

// All 16 Query Parameter Combinations (all use QueryCourses with different filters)
// #	prefix | number | section | school | Result
// 1	-	-	-	-	All courses for term
// 2	-	-	-	✓	Filter by school
// 3	-	-	✓	-	Error: Section requires prefix and number
// 4	-	-	✓	✓	Error: Section requires prefix and number
// 5	-	✓	-	-	Error: Number requires prefix
// 6	-	✓	-	✓	Error: Number requires prefix
// 7	-	✓	✓	-	Error: Section requires prefix and number
// 8	-	✓	✓	✓	Error: Section requires prefix and number
// 9	✓	-	-	-	Filter by prefix
// 10	✓	-	-	✓	Filter by prefix (school ignored)
// 11	✓	-	✓	-	Error: Section requires prefix and number
// 12	✓	-	✓	✓	Error: Section requires prefix and number
// 13	✓	✓	-	-	Filter by prefix and number
// 14	✓	✓	-	✓	Filter by prefix and number (school ignored)
// 15	✓	✓	✓	-	Single course by section
// 16	✓	✓	✓	✓	Single course by section (school ignored)
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

	// Validate parameter dependencies
	if section != "" && (prefix == "" || number == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Section parameter requires both prefix and number parameters"})
		return
	}
	if number != "" && prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number parameter requires prefix parameter"})
		return
	}

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

	// Handle section lookup (returns single course)
	if section != "" {
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
