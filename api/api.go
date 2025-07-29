package api

import (
	"net/http"
	"time"

	"github.com/acmutd/acmutd-api/firebase"
	"github.com/acmutd/acmutd-api/types"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

const (
	apiKeyCacheTTL    = 5 * time.Minute
	rateLimitCacheTTL = 1 * time.Minute
)

type API struct {
	db          *firebase.Firestore
	router      *gin.Engine
	apiKeyCache *cache.Cache
	rateLimiter *RateLimiter
}

func NewAPI(db *firebase.Firestore) *API {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	rl := NewRateLimiter()
	rl.StartCleanup(1 * time.Minute)

	return &API{
		db:          db,
		router:      router,
		apiKeyCache: cache.New(apiKeyCacheTTL, 10*time.Minute),
		rateLimiter: rl,
	}
}

func (api *API) SetupRoutes() {
	// Health check endpoint
	api.router.GET("/health", api.healthCheck)
	// Admin routes
	admin := api.router.Group("/admin")
	admin.Use(api.AuthMiddleware())
	admin.Use(api.RateLimitMiddleware())
	admin.Use(api.AdminMiddleware())
	{
		admin.POST("/apikeys", api.createAPIKey)
		admin.GET("/apikeys/:key", api.getAPIKey)
	}

	// API v1 routes (protected)
	v1 := api.router.Group("/api/v1")
	v1.Use(api.AuthMiddleware())
	v1.Use(api.RateLimitMiddleware())
	{
		// Course routes
		courses := v1.Group("/courses")
		{
			courses.GET("/", api.getAllCourses)
			courses.GET("/:term", api.getCoursesByTerm)
			courses.GET("/:term/prefix/:prefix", api.getCoursesByPrefix)
			courses.GET("/:term/prefix/:prefix/number/:number", api.getCoursesByNumber)
			courses.GET("/:term/school/:school", api.getCoursesBySchool)
			courses.GET("/:term/search", api.searchCourses)
		}

		terms := v1.Group("/terms")
		{
			terms.GET("/", api.getTerms)
		}

		// Schools routes
		schools := v1.Group("/schools")
		{
			schools.GET("/:term", api.getSchoolsByTerm)
		}
	}
}

func (api *API) Run(addr string) error {
	return api.router.Run(addr)
}

// Health check endpoint
func (api *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "ACM API is running",
	})
}

// Get all courses (requires term parameter)
func (api *API) getAllCourses(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Term parameter is required. Use /api/v1/courses/{term}",
	})
}

// Get courses by term
func (api *API) getCoursesByTerm(c *gin.Context) {
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
		courses, err = api.db.QueryByCourseNumber(c.Request.Context(), term, prefix, number)
	} else if prefix != "" {
		courses, err = api.db.QueryByCoursePrefix(c.Request.Context(), term, prefix)
	} else if school != "" {
		courses, err = api.db.QueryBySchool(c.Request.Context(), term, school)
	} else {
		// Get all courses for the term
		courses, err = api.db.GetAllCoursesByTerm(c.Request.Context(), term)
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
func (api *API) getCoursesByPrefix(c *gin.Context) {
	term := c.Param("term")
	prefix := c.Param("prefix")

	if term == "" || prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term and prefix parameters are required"})
		return
	}

	courses, err := api.db.QueryByCoursePrefix(c.Request.Context(), term, prefix)
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
func (api *API) getCoursesByNumber(c *gin.Context) {
	term := c.Param("term")
	prefix := c.Param("prefix")
	number := c.Param("number")

	if term == "" || prefix == "" || number == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term, prefix, and number parameters are required"})
		return
	}

	courses, err := api.db.QueryByCourseNumber(c.Request.Context(), term, prefix, number)
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
func (api *API) getCoursesBySchool(c *gin.Context) {
	term := c.Param("term")
	school := c.Param("school")

	if term == "" || school == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term and school parameters are required"})
		return
	}

	courses, err := api.db.QueryBySchool(c.Request.Context(), term, school)
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
func (api *API) searchCourses(c *gin.Context) {
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

	courses, err := api.db.SearchCourses(c.Request.Context(), term, query)
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
func (api *API) getSchoolsByTerm(c *gin.Context) {
	term := c.Param("term")

	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term parameter is required"})
		return
	}

	schools, err := api.db.GetSchoolsByTerm(c.Request.Context(), term)
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

func (api *API) getTerms(c *gin.Context) {
	terms, err := api.db.QueryAllTerms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(terms),
		"terms": terms,
	})
}

func (api *API) createAPIKey(c *gin.Context) {
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

	key, err := api.db.GenerateAPIKey(
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

func (api *API) getAPIKey(c *gin.Context) {
	key := c.Param("key")

	apiKey, err := api.db.GetAPIKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get API key"})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}
