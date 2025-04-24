package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bridgetunes/mtn-backend/api/routes"
	"github.com/bridgetunes/mtn-backend/internal/config"
	"github.com/bridgetunes/mtn-backend/internal/repositories/mongodb"
	"github.com/bridgetunes/mtn-backend/internal/services"
	mongodbpkg "github.com/bridgetunes/mtn-backend/pkg/mongodb"
	"github.com/bridgetunes/mtn-backend/pkg/mtnapi"
	"github.com/bridgetunes/mtn-backend/pkg/smsgateway"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to MongoDB
	mongoClient, err := mongodbpkg.NewClient(cfg.MongoDB.URI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Get database
	db := mongoClient.Database(cfg.MongoDB.Database)

	// Initialize repositories
	userRepo := mongodb.NewUserRepository(db)
	topupRepo := mongodb.NewTopupRepository(db)
	drawRepo := mongodb.NewDrawRepository(db)
	winnerRepo := mongodb.NewWinnerRepository(db)
	notificationRepo := mongodb.NewNotificationRepository(db)
	templateRepo := mongodb.NewTemplateRepository(db)
	campaignRepo := mongodb.NewCampaignRepository(db)
	blacklistRepo := mongodb.NewBlacklistRepository(db)

	// Initialize MTN API client
	mtnClient := mtnapi.NewClient(
		cfg.MTN.BaseURL,
		cfg.MTN.APIKey,
		cfg.MTN.APISecret,
		cfg.MTN.MockAPI,
	)

	// Initialize SMS gateways
	mtnGateway := smsgateway.NewMTNGateway(
		cfg.SMS.MTN.BaseURL,
		cfg.SMS.MTN.APIKey,
		cfg.SMS.MTN.APISecret,
		cfg.SMS.MockSMS,
	)
	kodobeGateway := smsgateway.NewKodobeGateway(
		cfg.SMS.Kodobe.BaseURL,
		cfg.SMS.Kodobe.APIKey,
		cfg.SMS.MockSMS,
	)

	// Initialize services
	userService := services.NewUserService(userRepo)
	topupService := services.NewTopupService(topupRepo, userService, mtnClient)
	drawService := services.NewDrawService(drawRepo, userRepo, winnerRepo)
	notificationService := services.NewNotificationService(
		notificationRepo,
		templateRepo,
		campaignRepo,
		userRepo,
		mtnGateway,
		kodobeGateway,
		cfg.SMS.DefaultGateway,
	)

	// Initialize router
	router := routes.SetupRouter(cfg, mongoClient.client)

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
