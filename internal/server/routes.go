package server

import (
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/internal/types"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", s.healthCheck)
	// Admin routes
	admin := router.Group("/admin")
	admin.Use(s.AuthMiddleware())
	admin.Use(s.RateLimitMiddleware())
	admin.Use(s.AdminMiddleware())
	{
		admin.POST("/apikeys", s.createAPIKey)
		admin.GET("/apikeys/:key", s.getAPIKey)
	}

	// API v1 routes (protected)
	v1 := router.Group("/api/v1")
	v1.Use(s.AuthMiddleware())
	v1.Use(s.RateLimitMiddleware())
	{
		// Course routes
		courses := v1.Group("/courses")
		{
			courses.GET("/", s.getAllCourses)
			courses.GET("/:term", s.getCoursesByTerm)
			courses.GET("/:term/prefix/:prefix", s.getCoursesByPrefix)
			courses.GET("/:term/prefix/:prefix/number/:number", s.getCoursesByNumber)
			courses.GET("/:term/school/:school", s.getCoursesBySchool)
			courses.GET("/:term/search", s.searchCourses)
		}

		terms := v1.Group("/terms")
		{
			terms.GET("/", s.getTerms)
		}

		// Schools routes
		schools := v1.Group("/schools")
		{
			schools.GET("/:term", s.getSchoolsByTerm)
		}
	}

	return router
}

// Health check endpoint
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "ACM API is running",
	})
}

// Get all courses (requires term parameter)
func (s *Server) getAllCourses(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Term parameter is required. Use /api/v1/courses/{term}",
	})
}

// Get courses by term
func (s *Server) getCoursesByTerm(c *gin.Context) {
	term := c.Param("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
		return
	}

	// Get query parameters for filtering
	prefix := c.Query("prefix")
	number := c.Query("number")
	school := c.Query("school")

	var courses []types.Course
	var err error

	// Apply filters based on query parameters
	if prefix != "" && number != "" {
		courses, err = s.db.QueryByCourseNumber(c.Request.Context(), term, prefix, number)
	} else if prefix != "" {
		courses, err = s.db.QueryByCoursePrefix(c.Request.Context(), term, prefix)
	} else if school != "" {
		courses, err = s.db.QueryBySchool(c.Request.Context(), term, school)
	} else {
		// Get all courses for the term
		courses, err = s.db.GetAllCoursesByTerm(c.Request.Context(), term)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"term":    term,
		"count":   len(courses),
		"courses": courses,
	})
}

// Get courses by prefix
func (s *Server) getCoursesByPrefix(c *gin.Context) {
	term := c.Param("term")
	prefix := c.Param("prefix")

	if term == "" || prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term and prefix parameters are required"})
		return
	}

	courses, err := s.db.QueryByCoursePrefix(c.Request.Context(), term, prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"term":    term,
		"prefix":  prefix,
		"count":   len(courses),
		"courses": courses,
	})
}

// Get courses by number
func (s *Server) getCoursesByNumber(c *gin.Context) {
	term := c.Param("term")
	prefix := c.Param("prefix")
	number := c.Param("number")

	if term == "" || prefix == "" || number == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term, prefix, and number parameters are required"})
		return
	}

	courses, err := s.db.QueryByCourseNumber(c.Request.Context(), term, prefix, number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"term":    term,
		"prefix":  prefix,
		"number":  number,
		"count":   len(courses),
		"courses": courses,
	})
}

// Get courses by school
func (s *Server) getCoursesBySchool(c *gin.Context) {
	term := c.Param("term")
	school := c.Param("school")

	if term == "" || school == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term and school parameters are required"})
		return
	}

	courses, err := s.db.QueryBySchool(c.Request.Context(), term, school)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"term":    term,
		"school":  school,
		"count":   len(courses),
		"courses": courses,
	})
}

// Search courses
func (s *Server) searchCourses(c *gin.Context) {
	term := c.Param("term")
	query := c.Query("q")

	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
		return
	}

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query parameter 'q' is required"})
		return
	}

	courses, err := s.db.SearchCourses(c.Request.Context(), term, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"term":    term,
		"query":   query,
		"count":   len(courses),
		"courses": courses,
	})
}

// Get schools by term
func (s *Server) getSchoolsByTerm(c *gin.Context) {
	term := c.Param("term")

	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
		return
	}

	schools, err := s.db.GetSchoolsByTerm(c.Request.Context(), term)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"term":    term,
		"count":   len(schools),
		"schools": schools,
	})
}

func (s *Server) getTerms(c *gin.Context) {
	terms, err := s.db.QueryAllTerms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(terms),
		"terms": terms,
	})
}

func (s *Server) createAPIKey(c *gin.Context) {
	var req struct {
		RateLimit     int    `json:"rate_limit" binding:"required"`
		WindowSeconds int    `json:"window_seconds" binding:"required"`
		IsAdmin       bool   `json:"is_admin"`
		ExpiresAt     string `json:"expires_at"` // ISO 8601 format
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

	// Parse expiration time
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

	key, err := s.db.GenerateAPIKey(
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

func (s *Server) getAPIKey(c *gin.Context) {
	key := c.Param("key")

	apiKey, err := s.db.GetAPIKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get API key"})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}
