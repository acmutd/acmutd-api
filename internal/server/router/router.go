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

	// Serve static assets
	router.Static("/assets", "frontend/dist/assets")

	router.GET("/swagger", handler.Default)
	router.GET("/openapi.yaml", handler.OpenAPISpec)
	router.GET("/health", handler.Health)

	// Serve SPA pages - serve index.html for React Router to handle client-side routing
	router.GET("/", func(c *gin.Context) {
		c.File("frontend/dist/index.html")
	})
	router.GET("/dashboard", func(c *gin.Context) {
		c.File("frontend/dist/index.html")
	})

	// Serve index.html for any unmatched route so React Router handles client-side navigation.
	router.NoRoute(func(c *gin.Context) {
		c.File("frontend/dist/index.html")
	})

	// Dashboard REST API — authenticated via Firebase ID token
	dash := router.Group("/dashboard/api/v1")
	dash.Use(mw.FirebaseAuth())
	{
		dash.GET("/users/me", handler.GetMe)
		dash.GET("/keys/me", handler.GetMyAPIKey)
		dash.POST("/keys/request", handler.RequestAPIKey)
		dash.POST("/keys/:keyId/regenerate", handler.RegenerateKey)
		dash.DELETE("/keys/:keyId", handler.RevokeKey)
		dash.DELETE("/keys/:keyId/permanent", handler.DeleteKey)
		dash.GET("/stats/daily", handler.GetMyDailyStats)
		dash.GET("/stats/requests", handler.GetMyRecentRequests)
		dash.GET("/config", handler.GetAppConfig)

		dashAdmin := dash.Group("")
		dashAdmin.Use(mw.DashboardAdmin())
		{
			dashAdmin.GET("/users", handler.ListUsers)
			dashAdmin.PUT("/users/:uid/ban", handler.BanUser)
			dashAdmin.PUT("/users/:uid/unban", handler.UnbanUser)
			dashAdmin.GET("/admin/keys", handler.ListAllKeys)
			dashAdmin.GET("/admin/keys/pending", handler.ListPendingKeys)
			dashAdmin.POST("/admin/keys/:keyId/approve", handler.ApproveKey)
			dashAdmin.POST("/admin/keys/:keyId/reject", handler.RejectKey)
			dashAdmin.POST("/admin/keys/:keyId/regenerate", handler.AdminRegenerateKey)
			dashAdmin.DELETE("/admin/keys/:keyId", handler.AdminRevokeKey)
			dashAdmin.DELETE("/admin/keys/:keyId/permanent", handler.AdminDeleteKey)
			dashAdmin.POST("/admin/keys", handler.AddKey)
			dashAdmin.PUT("/admin/keys/:keyId", handler.EditKey)
			dashAdmin.GET("/admin/stats/daily", handler.ListAllDailyStats)
			dashAdmin.GET("/admin/stats/requests", handler.ListAllRecentRequests)
			dashAdmin.PUT("/admin/config", handler.UpdateAppConfig)
			dashAdmin.GET("/admin/server/state", handler.GetInstanceState)
			dashAdmin.POST("/admin/server/start", handler.StartServer)
			dashAdmin.POST("/admin/server/stop", handler.StopServer)
			dashAdmin.POST("/admin/server/hackathon/enable", handler.EnableHackathonMode)
			dashAdmin.POST("/admin/server/hackathon/disable", handler.DisableHackathonMode)
			dashAdmin.GET("/admin/logs/scraper", handler.ListScraperLogs)
			dashAdmin.GET("/admin/logs/cron", handler.ListCronLogs)
			dashAdmin.GET("/admin/logs/server", handler.ListServerStateLogs)
			dashAdmin.GET("/admin/logs/promotion", handler.ListPromotionLogs)
		}
	}

	admin := router.Group("/admin")
	admin.Use(mw.Auth(), mw.RateLimit(), mw.Admin())
	{
		admin.POST("/apikeys", handler.CreateAPIKey)
		admin.GET("/apikeys/:key", handler.GetAPIKey)
	}

	v1 := router.Group("/api/v1")
	v1.Use(mw.Auth(), mw.RateLimit(), mw.Track())
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
