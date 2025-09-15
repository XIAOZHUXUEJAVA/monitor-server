package api

import (
	"monitor-server/internal/config"
	"monitor-server/internal/handler"
	"monitor-server/internal/middleware"
	"monitor-server/internal/service"
	"monitor-server/pkg/logger"

	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures the main router
func NewRouter(cfg *config.Config, logger *logger.Logger) *gin.Engine {
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add logging middleware
	router.Use(middleware.Logging(logger))

	// Add CORS middleware
	router.Use(middleware.CORS(cfg.CORS))

	// Initialize services
	monitorService := service.NewMonitorService()

	// Initialize handlers
	monitorHandler := handler.NewMonitorHandler(monitorService, logger)

	// Setup routes
	setupRoutes(router, monitorHandler)

	return router
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine, monitorHandler *handler.MonitorHandler) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "monitor-server",
		})
	})

	// API v1 routes
	v1 := router.Group("/api")
	{
		// System monitoring endpoints
		v1.GET("/cpu", monitorHandler.GetCPU)
		v1.GET("/memory", monitorHandler.GetMemory)
		v1.GET("/disk", monitorHandler.GetDisk)
		v1.GET("/network", monitorHandler.GetNetwork)
		v1.GET("/system", monitorHandler.GetSystem)
		v1.GET("/processes", monitorHandler.GetProcesses)
	}
}