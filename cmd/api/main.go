package main

import (
	"context"
	"log"
	// "net/http" // Removed unused import
	"time"

	"github.com/bridgetunes/mtn-backend/internal/handlers"
	"github.com/bridgetunes/mtn-backend/internal/repositories/mongodb"
	"github.com/bridgetunes/mtn-backend/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use environment variables in production
	mongoURI := "mongodb://mongodb:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")

	// Initialize database
	// db := client.Database("bridgetunes") // Original - Assuming this is correct, but check if env var needed
    dbName := "bridgetunes" // Consider using an environment variable
    db := client.Database(dbName)

	// Initialize repositories
	userRepo := mongodb.NewUserRepository(db)
	// topupRepo := mongodb.NewTopupRepository(db) // Assuming this exists
	campaignRepo := mongodb.NewCampaignRepository(db) // Added based on previous errors
	notificationTemplateRepo := mongodb.NewNotificationTemplateRepository(db) // Added based on previous errors
	transactionRepo := mongodb.NewTransactionRepository(db) // Added based on previous errors
	// Assuming other repos might be needed based on services used
	// prizeRepo := mongodb.NewPrizeRepository(db)
	// winnerRepo := mongodb.NewWinnerRepository(db)
	// configRepo := mongodb.NewSystemConfigRepository(db)

	// Initialize services
	// Assuming AuthService and TransactionService exist and are needed
	// authService := services.NewAuthService(userRepo, /* other dependencies */)
	// transactionService := services.NewTransactionService(transactionRepo, /* other dependencies */)
	// drawService := services.NewDrawServiceEnhanced(
	// 	 drawRepo, userRepo, winnerRepo, configRepo, topupRepo,
	// )
	// campaignService := services.NewCampaignService(campaignRepo)
	// notificationService := services.NewNotificationService(notificationTemplateRepo)
	// userService := services.NewUserService(userRepo)

    // --- Placeholder for actual service initialization --- 
    // Need to confirm which services are actually used by handlers
    // Example: If DrawHandlerEnhanced needs DrawService
    drawRepo := mongodb.NewDrawRepository(db)
    winnerRepo := mongodb.NewWinnerRepository(db)
    configRepo := mongodb.NewSystemConfigRepository(db)
    topupRepo := mongodb.NewTopupRepository(db) // Assuming this exists and is needed
    var drawService services.DrawService // Use the interface type
    drawService = services.NewDrawServiceEnhanced(drawRepo, userRepo, winnerRepo, configRepo, topupRepo)
    // --- End Placeholder --- 

	// Initialize handlers
	// authHandler := handlers.NewAuthHandler(authService)
	// transactionHandler := handlers.NewTransactionHandler(transactionService)
	// drawHandler := handlers.NewDrawHandlerEnhanced(drawService)
	// campaignHandler := handlers.NewCampaignHandler(campaignService)
	// notificationHandler := handlers.NewNotificationHandler(notificationService)
	// userHandler := handlers.NewUserHandler(userService)

    // --- Placeholder for actual handler initialization --- 
    // Example: Initialize only the handlers needed for the defined routes
    drawHandler := handlers.NewDrawHandlerEnhanced(drawService)
    // --- End Placeholder --- 

	// Initialize router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins for now, restrict in production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	api := router.Group("/api/v1")
	{
		// Draw management routes
		// Assuming drawHandler is correctly initialized above
		// Check if these handler methods exist in drawHandler
		// draws := api.Group("/draws")
		// {
		// 	 draws.GET("/config", drawHandler.GetDrawConfig)
		// 	 draws.GET("/prize-structure", drawHandler.GetPrizeStructure)
		// 	 draws.PUT("/prize-structure", drawHandler.UpdatePrizeStructure)
		// 	 draws.POST("", drawHandler.ScheduleDraw)
		// 	 draws.GET("", drawHandler.GetDraws) // This is the method causing the interface error
		// 	 draws.GET("/:id", drawHandler.GetDrawByID)
		// 	 draws.POST("/:id/execute", drawHandler.ExecuteDraw)
		// 	 draws.GET("/:id/winners", drawHandler.GetDrawWinners)
		// 	 draws.GET("/jackpot-history", drawHandler.GetJackpotHistory)
		// }

        // --- Placeholder for actual routes --- 
        // Define only the routes that have corresponding handlers initialized
        draws := api.Group("/draws")
        {
            // Assuming these methods exist in drawHandler
            // draws.GET("/config", drawHandler.GetDrawConfig)
            // draws.GET("/prize-structure", drawHandler.GetPrizeStructure)
            // draws.PUT("/prize-structure", drawHandler.UpdatePrizeStructure)
            // draws.POST("", drawHandler.ScheduleDraw)
            draws.GET("", drawHandler.GetDraws) // Keep the route causing the error for now
            // draws.GET("/:id", drawHandler.GetDrawByID)
            // draws.POST("/:id/execute", drawHandler.ExecuteDraw)
            // draws.GET("/:id/winners", drawHandler.GetDrawWinners)
            // draws.GET("/jackpot-history", drawHandler.GetJackpotHistory)
        }
        // --- End Placeholder --- 
	}

	// Start server
	port := ":8080" // Consider using an environment variable
	log.Printf("Server running on port %s", port)
	// Use http.ListenAndServe directly if not using Gin's Run
	// if err := http.ListenAndServe(port, router); err != nil {
	// 	 log.Fatalf("Failed to start server: %v", err)
	// }
    if err := router.Run(port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

