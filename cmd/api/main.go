package main

import (
	"context"
	"log"
	"os" // Import os package
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

	// --- Use environment variables --- 
	mongoURI := os.Getenv("MONGO_URI")
	 if mongoURI == "" {
		log.Println("WARNING: MONGO_URI environment variable not set. Using default.")
		// Fallback, but this likely won't work in Render if the env var isn't set correctly
		mongoURI = "mongodb://mongodb:27017" 
	}

	 dbName := os.Getenv("MONGO_DB_NAME")
	 if dbName == "" {
		log.Println("WARNING: MONGO_DB_NAME environment variable not set. Using default 'bridgetunes'.")
		 dbName = "bridgetunes"
	}

	 port := os.Getenv("PORT")
	 if port == "" {
		log.Println("WARNING: PORT environment variable not set. Using default ':8080'.")
		 port = "8080" // Render expects just the port number, not the colon
	}
	 port = ":" + port // Add the colon for Gin
	// --- End Environment Variables ---

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	 if err != nil {
		log.Fatalf("Failed to connect to MongoDB using URI %s: %v", mongoURI, err)
	}
	defer func() {
		 if err = client.Disconnect(context.Background()); err != nil {
			 log.Printf("Error disconnecting from MongoDB: %v", err)
		 }
	}()

	// Ping MongoDB to verify connection
	 if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")

	// Initialize database
    db := client.Database(dbName)

	// Initialize repositories needed for DrawService
	userRepo := mongodb.NewUserRepository(db)
	 drawRepo := mongodb.NewDrawRepository(db)
    winnerRepo := mongodb.NewWinnerRepository(db)
    configRepo := mongodb.NewSystemConfigRepository(db)
    topupRepo := mongodb.NewTopupRepository(db) // Assuming this exists and is needed by DrawServiceEnhanced

	// Initialize only the services needed by active handlers
    var drawService services.DrawService // Use the interface type
    drawService = services.NewDrawServiceEnhanced(drawRepo, userRepo, winnerRepo, configRepo, topupRepo)

	// Initialize only the handlers needed for the defined routes
    drawHandler := handlers.NewDrawHandlerEnhanced(drawService)

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
        draws := api.Group("/draws")
        {
            // Assuming these methods exist in drawHandler and match the service interface
            // draws.GET("/config", drawHandler.GetDrawConfig)
            // draws.GET("/prize-structure", drawHandler.GetPrizeStructure)
            // draws.PUT("/prize-structure", drawHandler.UpdatePrizeStructure)
            // draws.POST("", drawHandler.ScheduleDraw)
            draws.GET("", drawHandler.GetDraws)
            // draws.GET("/:id", drawHandler.GetDrawByID)
            // draws.POST("/:id/execute", drawHandler.ExecuteDraw)
            draws.GET("/:id/winners", drawHandler.GetWinnersByDrawID) // Use corrected method name
            draws.GET("/jackpot-history", drawHandler.GetJackpotHistory)
        }
	}

	// Start server
	log.Printf("Server running on port %s", port)
    if err := router.Run(port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}


