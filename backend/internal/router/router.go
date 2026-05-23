package router

import (
	"nas-partner/backend/internal/config"
	"nas-partner/backend/internal/database"
	ddnshandler "nas-partner/backend/internal/ddns/handler"
	"nas-partner/backend/internal/ddns/scheduler"
	"nas-partner/backend/internal/handler"
	"nas-partner/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())

	database.Init(cfg.DBPath)
	handler.SetConfig(cfg)

	api := r.Group("/api")
	{
		api.GET("/health", handler.Health)
		api.GET("/auth/status", handler.AuthStatus)
		api.POST("/auth/init", handler.AuthInit)
		api.POST("/auth/login", handler.AuthLogin)

		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg))
		{
			protected.GET("/me", handler.Me)

			ddns := protected.Group("/ddns")
			{
				ddns.GET("", ddnshandler.List)
				ddns.GET("/logs/latest", ddnshandler.ListLatestLogs)
				ddns.GET("/:id", ddnshandler.Get)
				ddns.POST("", ddnshandler.Create)
				ddns.PUT("/:id", ddnshandler.Update)
				ddns.DELETE("/:id", ddnshandler.Delete)
				ddns.POST("/:id/toggle", ddnshandler.Toggle)
				ddns.POST("/:id/run", ddnshandler.Run)
				ddns.GET("/net-interfaces", ddnshandler.ListNetInterfaces)
				ddns.POST("/test-ip", ddnshandler.TestIP)
				ddns.GET("/:id/logs", ddnshandler.ListRunLogs)
					ddns.DELETE("/:id/logs", ddnshandler.ClearLogs)
					ddns.POST("/logs/cleanup", ddnshandler.CleanupLogs)
			}
		}
	}

	// Start DDNS scheduler
	scheduler.Default.Start()

	return r
}
