package main

import (
	"context"
	"log"
	"net/http"
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
	db := client.Database("bridgetunes")
	
	// Initialize repositories
	userRepo := mongodb.NewUserRepository(db)
	topupRepo := mongodb.NewTopupRepository(db)
	drawRepo := mongodb.NewDrawRepository(db)
	winnerRepo := mongodb.NewWinnerRepository(db)
	configRepo := mongodb.NewSystemConfigRepository(db)
	
	// Initialize services
	drawService := services.NewDrawServiceEnhanced(
		drawRepo,
		userRepo,
		winnerRepo,
		configRepo,
		topupRepo,
	)
	
	// Initialize handlers
	drawHandler := handlers.NewDrawHandlerEnhanced(drawService)
	
	// Initialize router
	router := gin.Default()
	
	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
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
			draws.GET("/config", drawHandler.GetDrawConfig)
			draws.GET("/prize-structure", drawHandler.GetPrizeStructure)
			draws.PUT("/prize-structure", drawHandler.UpdatePrizeStructure)
			draws.POST("", drawHandler.ScheduleDraw)
			draws.GET("", drawHandler.GetDraws)
			draws.GET("/:id", drawHandler.GetDrawByID)
			draws.POST("/:id/execute", drawHandler.ExecuteDraw)
			draws.GET("/:id/winners", drawHandler.GetDrawWinners)
			draws.GET("/jackpot-history", drawHandler.GetJackpotHistory)
		}
	}
	
	// Start server
	port := ":8080"
	log.Printf("Server running on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

