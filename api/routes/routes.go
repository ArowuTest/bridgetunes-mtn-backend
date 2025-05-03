package routes

import (
	"log"
	"strings"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/config"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/handlers"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// HandlerDependencies holds all the handlers required by the router
type HandlerDependencies struct {
	AuthHandler        *handlers.AuthHandler
	UserHandler        *handlers.UserHandler
	DrawHandler        *handlers.DrawHandler // Assuming DrawHandlerEnhanced is now DrawHandler or similar
	TopupHandler       *handlers.TopupHandler
	NotificationHandler *handlers.NotificationHandler
	// Add other handlers as needed
}

// SetupRouter configures the Gin router with all application routes and middleware.
// It now accepts HandlerDependencies to allow for proper dependency injection.
func SetupRouter(cfg *config.Config, deps HandlerDependencies) *gin.Engine {
	// Set Gin mode based on config
	 if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New() // Use gin.New() instead of gin.Default() for more control

	// Logger middleware
	router.Use(gin.Logger())

	// Recovery middleware
	router.Use(gin.Recovery())

	// CORS Middleware - Use the configuration from cfg
	log.Printf("Configuring CORS with AllowedHosts: %v", cfg.Server.AllowedHosts)
	corsConfig := cors.Config{
		AllowOrigins:     cfg.Server.AllowedHosts, // Use AllowedHosts from config
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			// Allow requests from configured origins
			for _, allowedOrigin := range cfg.Server.AllowedHosts {
				 if strings.EqualFold(allowedOrigin, origin) || allowedOrigin == "*" {
					return true
				}
			}
			// Allow requests with no origin (like curl requests or mobile apps)
			// Be cautious with this in production if not needed.
			// if origin == "" {
			// 	return true
			// }
			log.Printf("CORS blocked origin: %s", origin)
			return false
		},
		// MaxAge: 12 * time.Hour, // Optional: Cache preflight response
	}
	router.Use(cors.New(corsConfig))

	// Add a simple handler for the root path ("/")
	router.GET("/", func(c *gin.Context) {
		 c.JSON(200, gin.H{"status": "Service is running"})
	})

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		 auth := public.Group("/auth")
		{
			 auth.POST("/login", deps.AuthHandler.Login)
			 auth.POST("/register", deps.AuthHandler.Register) // Assuming register handler exists
			// Add other public auth routes like refresh token if needed
		}

		// Add other public routes here if any
		// Example: public.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "UP"}) })
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(cfg)) // Apply JWT authentication middleware
	{
		 users := protected.Group("/users")
		{
			 users.GET("/me", deps.UserHandler.GetMe) // Example protected user route
			// Add other protected user routes
		}

		 draws := protected.Group("/draws")
		{
			 draws.POST("", deps.DrawHandler.CreateDraw)
			 draws.GET("", deps.DrawHandler.GetDraws)
			 draws.GET("/:id", deps.DrawHandler.GetDrawByID)
			 draws.PUT("/:id", deps.DrawHandler.UpdateDraw)
			 draws.DELETE("/:id", deps.DrawHandler.DeleteDraw)
			 draws.POST("/schedule", deps.DrawHandler.ScheduleDraw)
			 draws.POST("/execute/:id", deps.DrawHandler.ExecuteDraw)
			 draws.GET("/winners/:id", deps.DrawHandler.GetWinners)
			 draws.GET("/date/:date", deps.DrawHandler.GetDrawByDate) // The route causing 404 earlier
			 draws.GET("/default-digits/:day", deps.DrawHandler.GetDefaultDigitsForDay) // The route causing 404 earlier
			 draws.GET("/config", deps.DrawHandler.GetDrawConfig) // Assuming this handler exists
			 draws.GET("/prize-structure", deps.DrawHandler.GetPrizeStructure) // Assuming this handler exists
			// Add other draw routes
		}

		 topups := protected.Group("/topups")
		{
			 topups.POST("", deps.TopupHandler.CreateTopup)
			 topups.GET("", deps.TopupHandler.GetTopups)
			// Add other topup routes
		}

		 notifications := protected.Group("/notifications")
		{
			 notifications.GET("", deps.NotificationHandler.GetNotifications)
			// Add other notification routes
		}

		// Dashboard route
		 dashboard := protected.Group("/dashboard")
		{
			 dashboard.GET("/stats", deps.UserHandler.GetDashboardStats) // Add the dashboard stats route
		}
	}

	// Handle OPTIONS requests for preflight checks (CORS)
	// Gin-contrib/cors handles this automatically if configured correctly.

	// Route for 404 Not Found
	router.NoRoute(func(c *gin.Context) {
		 c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	return router
}



