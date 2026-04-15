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
			courses.GET("/general", handler.GetGeneralCourses)
			courses.GET("/general/:prefix/:number", handler.GetGeneralCourse)
			courses.GET("/sections", handler.GetSections)
			courses.GET("/sections/:prefix/:number/:section/:term", handler.GetSectionByParams)
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
	}

	return router
}
