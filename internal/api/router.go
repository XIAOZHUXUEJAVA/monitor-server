package api

import (
	"monitor-server/internal/config"
	"monitor-server/internal/database"
	"monitor-server/internal/handler"
	"monitor-server/internal/middleware"
	"monitor-server/internal/service"
	"monitor-server/pkg/logger"

	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures the main router
func NewRouter(cfg *config.Config, logger *logger.Logger, db *database.DB) *gin.Engine {
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
	hostHandler := handler.NewHostHandler(db.DB)
	hostConfigHandler := handler.NewHostConfigHandler(db.DB)
	hostGroupHandler := handler.NewHostGroupHandler(db.DB)

	// Setup routes
	setupRoutes(router, monitorHandler, hostHandler, hostConfigHandler, hostGroupHandler)

	return router
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine, monitorHandler *handler.MonitorHandler, hostHandler *handler.HostHandler, hostConfigHandler *handler.HostConfigHandler, hostGroupHandler *handler.HostGroupHandler) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "monitor-server",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// System monitoring endpoints
		v1.GET("/cpu", monitorHandler.GetCPU)
		v1.GET("/memory", monitorHandler.GetMemory)
		v1.GET("/disk", monitorHandler.GetDisk)
		v1.GET("/network", monitorHandler.GetNetwork)
		v1.GET("/system", monitorHandler.GetSystem)
		v1.GET("/processes", monitorHandler.GetProcesses)

		// Host management endpoints
		hosts := v1.Group("/hosts")
		{
			hosts.POST("", hostHandler.CreateHost)
			hosts.GET("", hostHandler.GetHosts)
			hosts.GET("/stats", hostHandler.GetHostStats)
			hosts.PUT("/batch/status", hostHandler.BatchUpdateHostStatus)
			hosts.GET("/:id", hostHandler.GetHost)
			hosts.PUT("/:id", hostHandler.UpdateHost)
			hosts.DELETE("/:id", hostHandler.DeleteHost)
			
			// Host configuration endpoints
			hosts.GET("/:id/configs", hostConfigHandler.GetHostConfigs)
			hosts.GET("/:id/configs/:key", hostConfigHandler.GetHostConfigByKey)
			hosts.PUT("/:id/configs/:key", hostConfigHandler.UpdateHostConfigValue)
			
			// Host group relationships
			hosts.GET("/:id/groups", hostGroupHandler.GetHostGroupsForHost)
		}

		// Host configuration endpoints
		hostConfigs := v1.Group("/host-configs")
		{
			hostConfigs.POST("", hostConfigHandler.CreateHostConfig)
			hostConfigs.POST("/batch", hostConfigHandler.BatchCreateHostConfigs)
			hostConfigs.GET("/:id", hostConfigHandler.GetHostConfig)
			hostConfigs.PUT("/:id", hostConfigHandler.UpdateHostConfig)
			hostConfigs.DELETE("/:id", hostConfigHandler.DeleteHostConfig)
		}

		// Host group endpoints
		hostGroups := v1.Group("/host-groups")
		{
			hostGroups.POST("", hostGroupHandler.CreateHostGroup)
			hostGroups.GET("", hostGroupHandler.GetHostGroups)
			hostGroups.GET("/stats", hostGroupHandler.GetHostGroupStats)
			hostGroups.GET("/:id", hostGroupHandler.GetHostGroup)
			hostGroups.PUT("/:id", hostGroupHandler.UpdateHostGroup)
			hostGroups.DELETE("/:id", hostGroupHandler.DeleteHostGroup)
			
			// Host group member management
			hostGroups.GET("/:id/hosts", hostGroupHandler.GetGroupHosts)
			hostGroups.POST("/:id/hosts", hostGroupHandler.AddHostsToGroup)
			hostGroups.DELETE("/:id/hosts", hostGroupHandler.RemoveHostsFromGroup)
		}
	}

	// Legacy API routes (for backward compatibility)
	legacy := router.Group("/api")
	{
		legacy.GET("/cpu", monitorHandler.GetCPU)
		legacy.GET("/memory", monitorHandler.GetMemory)
		legacy.GET("/disk", monitorHandler.GetDisk)
		legacy.GET("/network", monitorHandler.GetNetwork)
		legacy.GET("/system", monitorHandler.GetSystem)
		legacy.GET("/processes", monitorHandler.GetProcesses)
	}
}