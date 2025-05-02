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
	// "github.com/ArowuTest/bridgetunes-mtn-backend/internal/database" // Replaced by pkg/mongodb
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/handlers"
	// "github.com/ArowuTest/bridgetunes-mtn-backend/internal/repository" // Changed to plural below
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	// Import the MongoDB repository implementation
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories" // Interface package
	mongorepo "github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories/mongodb" // Implementation package
	mongodb "github.com/ArowuTest/bridgetunes-mtn-backend/pkg/mongodb" // MongoDB client helper
	// "github.com/gin-contrib/cors" // No longer needed here, handled in routes.go
	// "github.com/gin-gonic/gin" // No longer needed here, handled in routes.go
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".") // Load from current directory or specify path
	
	// Check if config loading was successful
	// It's important to handle this error before using cfg
	// Note: The original code had this check, ensure it's still present or re-add if necessary.
	// If LoadConfig already logs and exits on error, this might be redundant.
	// However, explicit check is safer.
	 if err != nil {
	 	log.Fatalf("Failed to load configuration: %v", err)
	 }


	// Connect to MongoDB using the pkg helper
	// Note: Ensure cfg.MongoDB.URI is correctly defined in your config struct and loaded
	mongoClient, err := mongodb.NewClient(cfg.MongoDB.URI)
	
	// Check if MongoDB connection was successful
	// It's important to handle this error before proceeding
	 if err != nil {
	 	log.Fatalf("Failed to connect to MongoDB: %v", err)
	 }
	defer func() {
		// Use context.Background() or context.TODO() for disconnection
		// Consider a timeout context if disconnection might hang
		 if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize Database (Get the specific database instance)
	// Ensure cfg.MongoDB.DBName is correctly defined and loaded
	 db := mongoClient.Database(cfg.MongoDB.DBName)

	// Initialize Repositories using the implementation package, assigning to interface types
	var userRepo repositories.UserRepository = mongorepo.NewUserRepository(db)
	var drawRepo repositories.DrawRepository = mongorepo.NewDrawRepository(db)
	var topupRepo repositories.TopupRepository = mongorepo.NewTopupRepository(db)
	var notificationRepo repositories.NotificationRepository = mongorepo.NewNotificationRepository(db)
	// Add other repositories as needed, ensuring they use the correct interfaces and implementations

	// Initialize Services (These should be fine as they depend on the interfaces)
	// Ensure NewAuthService, NewDrawService, etc., exist and accept the correct repository interface types
	 authService := services.NewAuthService(userRepo, cfg)
	 drawService := services.NewDrawService(drawRepo)
	 topupService := services.NewTopupService(topupRepo)
	 notificationService := services.NewNotificationService(notificationRepo)
	 userService := services.NewUserService(userRepo)
	// Add other services as needed

	// Initialize Handlers (These should be fine as they depend on services)
	// Ensure NewAuthHandler, NewDrawHandler, etc., exist and accept the correct service types
	 authHandler := handlers.NewAuthHandler(authService)
	 drawHandler := handlers.NewDrawHandler(drawService) // Use the standard DrawHandler
	 topupHandler := handlers.NewTopupHandler(topupService)
	 notificationHandler := handlers.NewNotificationHandler(notificationService)
	 userHandler := handlers.NewUserHandler(userService)
	// Add other handlers as needed

	// Create Handler Dependencies struct
	// Ensure routes.HandlerDependencies struct definition matches the handlers being passed
	 handlerDeps := routes.HandlerDependencies{
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		DrawHandler:        drawHandler,
		TopupHandler:       topupHandler,
		NotificationHandler: notificationHandler,
		// Add other handlers here if they are defined in HandlerDependencies
	}

	// Setup Router using the centralized function from routes package
	// Ensure routes.SetupRouter accepts cfg and handlerDeps correctly
	 router := routes.SetupRouter(cfg, handlerDeps)

	// Start the server
	// Ensure cfg.Server.Port is correctly defined and loaded
	 srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)

	// Run server in a goroutine so that it doesn't block
	go func() {
		// Check for http.ErrServerClosed specifically to avoid logging it as a fatal error on graceful shutdown
		 if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// Notify on SIGINT (Ctrl+C) and SIGTERM (kill/system shutdown)
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



