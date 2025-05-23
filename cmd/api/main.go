package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/api/routes"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/config"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/handlers"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories" // Interface package
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"

	// Implementation packages
	"github.com/ArowuTest/bridgetunes-mtn-backend/pkg/mongodb"    // MongoDB client helper
	"github.com/ArowuTest/bridgetunes-mtn-backend/pkg/smsgateway" // SMS Gateway implementations

	// Import the specific repository implementation package
	// It's common practice to alias it to avoid name collisions if needed,
	// but here 'mongorepo' is clear enough.
	mongorepo "github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories/mongodb"
)

func main() {
	// Load configuration using the correct function name
	cfg, err := config.Load() // Changed from LoadConfig
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// *** ADDED LOGGING HERE ***
	log.Printf("[INFO] Effective MongoDB Database Name from Config: %s", cfg.MongoDB.Database)
	// *** ADDED JWT SECRET DEBUG LOGGING ***
	if cfg.JWT.Secret != "" {
		log.Printf("[DEBUG] JWT Secret loaded (first 5 chars): %s...", cfg.JWT.Secret[:5])
	} else {
		log.Println("[ERROR] JWT Secret IS EMPTY after config load!")
	}

	// Connect to MongoDB using the pkg helper
	mongoClient, err := mongodb.NewClient(cfg.MongoDB.URI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize Database (Get the specific database instance)
	// Use cfg.MongoDB.Database which was the intended field name
	db := mongoClient.Database(cfg.MongoDB.Database)

	// Initialize ALL Repositories using the implementation package
	var userRepo repositories.UserRepository = mongorepo.NewUserRepository(db)
	var drawRepo repositories.DrawRepository = mongorepo.NewDrawRepository(db)
	// var topupRepo repositories.TopupRepository = mongorepo.NewTopupRepository(db) // Commented out - Unused
	// var notificationRepo repositories.NotificationRepository = mongorepo.NewNotificationRepository(db) // Commented out - Unused
	var adminUserRepo repositories.AdminUserRepository = mongorepo.NewAdminUserRepository(db)
	var winnerRepo repositories.WinnerRepository = mongorepo.NewWinnerRepository(db) // Added Winner Repo
	// var templateRepo repositories.TemplateRepository = mongorepo.NewTemplateRepository(db) // Commented out - Unused
	// var campaignRepo repositories.CampaignRepository = mongorepo.NewCampaignRepository(db) // Commented out - Unused
	// Initialize Blacklist and SystemConfig repositories
	var blacklistRepo repositories.BlacklistRepository = mongorepo.NewBlacklistRepository(db) // Uncommented
	var systemConfigRepo repositories.SystemConfigRepository = mongorepo.NewSystemConfigRepository(db) // Uncommented
	// Initialize PointTransaction and JackpotRollover repositories (Added)
	var pointTransactionRepo repositories.PointTransactionRepository = mongorepo.NewPointTransactionRepository(db)
	var jackpotRolloverRepo repositories.JackpotRolloverRepository = mongorepo.NewJackpotRolloverRepository(db)
	var eventRepo repositories.EventRepository = mongorepo.NewEventRepository(db)

	// Initialize External Clients
	// mtnClient := mtnapi.NewClient(cfg.MTN.BaseURL, cfg.MTN.APIKey, cfg.MTN.APISecret, cfg.MTN.MockAPI) // Commented out - Currently unused

	// Initialize SMS Gateways
	var mtnGateway smsgateway.Gateway // Corrected typo
	var kodobeGateway smsgateway.Gateway
	if cfg.SMS.MockSMSGateway {
		mtnGateway = smsgateway.NewMockGateway("MTN_Mock")
		kodobeGateway = smsgateway.NewMockGateway("Kodobe_Mock")
	} else {
		// Pass the MockSMS flag (which is false here) to the constructors
		mtnGateway = smsgateway.NewMTNGateway(cfg.SMS.MTNGateway.BaseURL, cfg.SMS.MTNGateway.APIKey, cfg.SMS.MTNGateway.APISecret, false)
		kodobeGateway = smsgateway.NewKodobeGateway(cfg.SMS.KodobeGateway.BaseURL, cfg.SMS.KodobeGateway.APIKey, false)
	}

	// Initialize Services using Legacy constructors with ALL dependencies
	// Note: Ensure the service instances are stored with the correct type for dependency injection
	authService := services.NewAuthService(adminUserRepo, cfg.JWT.Secret, cfg.JWT.ExpiresIn) // Pass JWT secret and expiration
	legacyUserService := services.NewLegacyUserService(userRepo)
	// Pass blacklistRepo and systemConfigRepo to NewDrawService
	// Use correct constructor name: NewDrawService instead of NewLegacyDrawService
	drawServiceInstance := services.NewDrawService(drawRepo, userRepo, winnerRepo, blacklistRepo, systemConfigRepo, pointTransactionRepo, jackpotRolloverRepo) // Added missing pointTransactionRepo, jackpotRolloverRepo
	// Use correct constructor name: NewTopupService instead of NewLegacyTopupService
	// Pass correct arguments: userRepo, pointTransactionRepo, drawServiceInstance
	topupServiceInstance := services.NewTopupService(userRepo, pointTransactionRepo, drawServiceInstance)
	// Use correct arguments for NewLegacyNotificationService: userRepo, mtnGateway, kodobeGateway, cfg.SMS.DefaultGateway
	legacyNotificationService := services.NewLegacyNotificationService(
		userRepo, // Corrected: Pass userRepo
		mtnGateway,
		kodobeGateway,
		cfg.SMS.DefaultGateway,
	)
	eventServiceInstance := services.NewEventService(eventRepo)

	// Store services using interface types if handlers expect interfaces (recommended)
	var userService services.UserService = legacyUserService
	var drawService services.DrawService = drawServiceInstance // Use the new instance
	var topupService services.TopupService = topupServiceInstance // Use the new instance
	var notificationService services.NotificationService = legacyNotificationService
	var eventService services.EventServiceInterface = eventServiceInstance

	// Initialize Handlers (Assuming handlers accept interface types)
	authHandler := handlers.NewAuthHandler(authService)
	drawHandler := handlers.NewDrawHandler(drawService)
	topupHandler := handlers.NewTopupHandler(topupService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	userHandler := handlers.NewUserHandler(userService)
	eventHandler := handlers.NewEventHandler(eventService)
	// Add other handlers as needed

	// Create Handler Dependencies struct (Assuming it uses interface types)
	handlerDeps := routes.HandlerDependencies{
		AuthHandler:         authHandler,
		UserHandler:         userHandler,
		DrawHandler:         drawHandler,
		TopupHandler:        topupHandler,
		NotificationHandler: notificationHandler,
		EventHandler:        eventHandler,
		// Add other handlers here if they are defined in HandlerDependencies
		// Add missing handlers based on routes.go if needed
		// Example: BlacklistHandler, SystemConfigHandler, WinnerHandler
	}

	// Setup Router using the centralized function from routes package
	router := routes.SetupRouter(cfg, handlerDeps)

	// Determine port: Use PORT environment variable if set, otherwise use config
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	// Start the server
	srv := &http.Server{
		Addr:    ":" + port, // Use the determined port
		Handler: router,
	}

	log.Printf("Server starting on port %s", port) // Log the actual port being used

	// Run server in a goroutine so that it doesn't block
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}


