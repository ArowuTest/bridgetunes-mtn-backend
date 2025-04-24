package routes

import (
	"github.com/bridgetunes/mtn-backend/internal/config"
	"github.com/bridgetunes/mtn-backend/internal/handlers"
	"github.com/bridgetunes/mtn-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRouter sets up the router
func SetupRouter(cfg *config.Config, mongoClient *mongo.Client) *gin.Engine {
	// Create router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// Create handlers
	userHandler := handlers.NewUserHandler(nil) // Will be initialized in dependency injection
	topupHandler := handlers.NewTopupHandler(nil)
	drawHandler := handlers.NewDrawHandler(nil)
	notificationHandler := handlers.NewNotificationHandler(nil)

	// Public routes
	public := router.Group("/api/v1")
	{
		// Health check
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})

		// Auth routes
		auth := public.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Login endpoint (to be implemented)",
				})
			})
		}

		// Opt-in/opt-out routes
		public.POST("/opt-in", userHandler.OptIn)
		public.POST("/opt-out", userHandler.OptOut)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(cfg))
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("", userHandler.GetAllUsers)
			users.GET("/count", userHandler.GetUserCount)
			users.GET("/:id", userHandler.GetUserByID)
			users.GET("/msisdn/:msisdn", userHandler.GetUserByMSISDN)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Topup routes
		topups := protected.Group("/topups")
		{
			topups.GET("", topupHandler.GetTopupsByDateRange)
			topups.GET("/count", topupHandler.GetTopupCount)
			topups.GET("/:id", topupHandler.GetTopupByID)
			topups.GET("/msisdn/:msisdn", topupHandler.GetTopupsByMSISDN)
			topups.POST("", topupHandler.CreateTopup)
			topups.POST("/process", topupHandler.ProcessTopups)
		}

		// Draw routes
		draws := protected.Group("/draws")
		{
			draws.GET("", drawHandler.GetDrawsByDateRange)
			draws.GET("/count", drawHandler.GetDrawCount)
			draws.GET("/:id", drawHandler.GetDrawByID)
			draws.GET("/date/:date", drawHandler.GetDrawByDate)
			draws.GET("/status/:status", drawHandler.GetDrawsByStatus)
			draws.GET("/default-digits/:day", drawHandler.GetDefaultEligibleDigits)
			draws.POST("/schedule", drawHandler.ScheduleDraw)
			draws.POST("/:id/execute", drawHandler.ExecuteDraw)
		}

		// Notification routes
		notifications := protected.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotificationsByStatus)
			notifications.GET("/count", notificationHandler.GetNotificationCount)
			notifications.GET("/:id", notificationHandler.GetNotificationByID)
			notifications.GET("/msisdn/:msisdn", notificationHandler.GetNotificationsByMSISDN)
			notifications.GET("/campaign/:id", notificationHandler.GetNotificationsByCampaignID)
			notifications.GET("/status/:status", notificationHandler.GetNotificationsByStatus)
			notifications.POST("/send-sms", notificationHandler.SendSMS)

			// Campaign routes
			campaigns := notifications.Group("/campaigns")
			{
				campaigns.GET("/count", notificationHandler.GetCampaignCount)
				campaigns.POST("", notificationHandler.CreateCampaign)
				campaigns.POST("/:id/execute", notificationHandler.ExecuteCampaign)
			}

			// Template routes
			templates := notifications.Group("/templates")
			{
				templates.GET("", notificationHandler.GetAllTemplates)
				templates.GET("/count", notificationHandler.GetTemplateCount)
				templates.GET("/:id", notificationHandler.GetTemplateByID)
				templates.GET("/name/:name", notificationHandler.GetTemplateByName)
				templates.GET("/type/:type", notificationHandler.GetTemplatesByType)
				templates.POST("", notificationHandler.CreateTemplate)
				templates.PUT("/:id", notificationHandler.UpdateTemplate)
				templates.DELETE("/:id", notificationHandler.DeleteTemplate)
			}
		}
	}

	return router
}
