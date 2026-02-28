package handlers

import (
	"net/http"
	"strings"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
)

// GetCourses fetches courses with optional filtering by query parameters.
// Query parameters:
//   - term: the term to query (required, e.g., "25s", "24f")
//
// Optional query parameters (can be combined):
//   - prefix: filters by course prefix (e.g., "cs")
//   - number: filters by course number (e.g., "1337") - works with or without prefix
//   - section: filters by section (e.g., "001") - works with or without prefix/number
//   - school: filters by school code (e.g., "ecs", "nsm")
//   - instructor: filters by instructor name (substring match)
//   - instructor_id: filters by instructor ID (substring match)
//   - days: filters by days of the week (e.g., "monday", "monday, wednesday")
//   - times: filters by time in 24h format (e.g., "14:00 - 14:50")
//   - times_12h: filters by time in 12h format (e.g., "2:00 PM - 2:50 PM")
//   - location: filters by location (e.g., "SCI_1.210", supports spaces or underscores)
//   - q: search query for title, topic, or instructor name
func (h *Handler) GetCourses(c *gin.Context) {
	term := strings.ToLower(strings.TrimSpace(c.Query("term")))
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term query parameter is required (e.g., ?term=25s)"})
		return
	}

	// Parse optional query parameters
	prefix := strings.ToLower(strings.TrimSpace(c.Query("prefix")))
	number := strings.ToLower(strings.TrimSpace(c.Query("number")))
	section := strings.ToLower(strings.TrimSpace(c.Query("section")))
	school := strings.ToLower(strings.TrimSpace(c.Query("school")))
	instructor := strings.TrimSpace(c.Query("instructor"))
	instructorID := strings.TrimSpace(c.Query("instructor_id"))
	days := strings.TrimSpace(c.Query("days"))
	times := strings.TrimSpace(c.Query("times"))
	times12h := strings.ToUpper(strings.TrimSpace(c.Query("times_12h")))
	location := strings.TrimSpace(c.Query("location"))
	// should change firestore to have building name and room number as separate values
	// plus the location parameter it already has
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
		Instructor:   instructor,
		InstructorID: instructorID,
		Days:         days,
		Times:        times,
		Times12h:     times12h,
		Location:     location,
		Search:       search,
		Limit:        params.Limit,
		Offset:       params.Offset,
	}

	// Build query object for response
	queryMeta := gin.H{
		"term":          term,
		"prefix":        prefix,
		"number":        number,
		"section":       section,
		"school":        school,
		"instructor":    instructor,
		"instructor_id": instructorID,
		"days":          days,
		"times":         times,
		"times_12h":     times12h,
		"location":      location,
		"search":        search,
	}

	courses, hasNext, err := h.db.QueryCourses(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination := buildPaginationMeta(params, len(courses), hasNext)

	c.JSON(http.StatusOK, gin.H{
		"count":      len(courses),
		"courses":    courses,
		"pagination": pagination,
		"query":      queryMeta,
	})
}
