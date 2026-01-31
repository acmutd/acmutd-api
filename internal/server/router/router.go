package router

import (
	"net/http"

	"github.com/acmutd/acmutd-api/internal/server/handlers"
	"github.com/acmutd/acmutd-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// New wires handlers and middleware into an HTTP router.
func New(handler *handlers.Handler, mw *middleware.Manager) http.Handler {
	router := gin.Default()

	router.GET("/health", handler.Health)

	admin := router.Group("/admin")
	admin.Use(mw.Auth(), mw.RateLimit(), mw.Admin())
	{
		admin.POST("/apikeys", handler.CreateAPIKey)
		admin.GET("/apikeys/:key", handler.GetAPIKey)
	}

	v1 := router.Group("/api/v1")
	v1.Use(mw.Auth(), mw.RateLimit())
	{
		courses := v1.Group("/courses")
		{
			courses.GET("/", handler.GetAllCourses)
			courses.GET("/:term", handler.GetCoursesByTerm)
			courses.GET("/:term/prefix/:prefix", handler.GetCoursesByPrefix)
			courses.GET("/:term/prefix/:prefix/number/:number", handler.GetCoursesByNumber)
			courses.GET("/:term/search", handler.SearchCourses)
		}

		terms := v1.Group("/terms")
		{
			terms.GET("/", handler.GetTerms)
		}

		professors := v1.Group("/professors")
		{
			professors.GET("/id/:id", handler.GetProfessorByID)
			professors.GET("/name/:name", handler.GetProfessorsByName)
		}

		grades := v1.Group("/grades")
		{
			grades.GET("/prof/id/:id", handler.GetGradesByProfID)
			grades.GET("/prof/name/:name", handler.GetGradesByProfName)
			grades.GET("/term/:term", handler.GetGradesByTerm)
			grades.GET("/prefix/:prefix", handler.GetGradesByPrefix)
			grades.GET("/prefix/:prefix/number/:number", handler.GetGradesByPrefixAndNumber)
			grades.GET("/prefix/:prefix/number/:number/term/:term/section/:section", handler.GetGradesBySection)
		}
	}

	return router
}
