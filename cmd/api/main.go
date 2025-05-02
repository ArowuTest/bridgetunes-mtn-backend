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
	// Import the MongoDB repository implementation
	
	// Implementation packages
	"github.com/ArowuTest/bridgetunes-mtn-backend/pkg/mongodb" // MongoDB client helper
	
	// Import the specific repository implementation package
	// It's common practice to alias it to avoid name collisions if needed,
	// but here 'mongorepo' is clear enough.
	mongorepo "github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories/mongodb"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".") // Load from current directory or specify path
	 if err != nil {
	 	log.Fatalf("Failed to load configuration: %v", err)
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
	 db := mongoClient.Database(cfg.MongoDB.DBName)

	// Initialize Repositories using the implementation package, assigning to interface types
	var userRepo repositories.UserRepository = mongorepo.NewUserRepository(db)
	var drawRepo repositories.DrawRepository = mongorepo.NewDrawRepository(db)
	var topupRepo repositories.TopupRepository = mongorepo.NewTopupRepository(db)
	var notificationRepo repositories.NotificationRepository = mongorepo.NewNotificationRepository(db)
	// *** Initialize the new AdminUserRepository ***
	var adminUserRepo repositories.AdminUserRepository = mongorepo.NewAdminUserRepository(db)
	// Add other repositories as needed

	// Initialize Services
	// *** Pass adminUserRepo to NewAuthService ***
	 authService := services.NewAuthService(adminUserRepo /*, cfg.JWT.Secret */) // Pass admin repo, uncomment JWT secret when implemented
	 drawService := services.NewDrawService(drawRepo)
	 topupService := services.NewTopupService(topupRepo)
	 notificationService := services.NewNotificationService(notificationRepo)
	 userService := services.NewUserService(userRepo)
	// Add other services as needed

	// Initialize Handlers
	 authHandler := handlers.NewAuthHandler(authService)
	 drawHandler := handlers.NewDrawHandler(drawService)
	 topupHandler := handlers.NewTopupHandler(topupService)
	 notificationHandler := handlers.NewNotificationHandler(notificationService)
	 userHandler := handlers.NewUserHandler(userService)
	// Add other handlers as needed

	// Create Handler Dependencies struct
	 handlerDeps := routes.HandlerDependencies{
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		DrawHandler:        drawHandler,
		TopupHandler:       topupHandler,
		NotificationHandler: notificationHandler,
		// Add other handlers here if they are defined in HandlerDependencies
	}

	// Setup Router using the centralized function from routes package
	 router := routes.SetupRouter(cfg, handlerDeps)

	// Start the server
	 srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)

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



