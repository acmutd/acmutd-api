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

// MAYBE TODO: Add term as a query parameter also but limit it so that you can't just get all courses in firestore cause that'll rack up costs

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
	term := strings.ToLower(strings.TrimSpace(c.Param("term")))
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
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
	times12h := strings.TrimSpace(c.Query("times_12h")) // PM needs to be capital in call, need to fix TODO
	location := strings.TrimSpace(c.Query("location"))
	// supports substring match of BUILDINGNAME_ROOMNUMBER, needs underscore between.
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
	if instructor != "" {
		responseMeta["instructor"] = instructor
	}
	if instructorID != "" {
		responseMeta["instructor_id"] = instructorID
	}
	if days != "" {
		responseMeta["days"] = days
	}
	if times != "" {
		responseMeta["times"] = times
	}
	if times12h != "" {
		responseMeta["times_12h"] = times12h
	}
	if location != "" {
		responseMeta["location"] = location
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
