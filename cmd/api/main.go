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
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/database"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/handlers"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repository"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	// Import the MongoDB repository implementation
	mongorepo "github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories/mongodb"
	// "github.com/gin-contrib/cors" // No longer needed here, handled in routes.go
	// "github.com/gin-gonic/gin" // No longer needed here, handled in routes.go
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".") // Load from current directory or specify path
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to MongoDB
	mongoClient, err := database.ConnectDB(cfg)
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

	// Initialize Repositorie		userRepo := mongorepo.NewUserRepository(db)
		 drawRepo := mongorepo.NewDrawRepository(d		 topupRepo := mongorepo.NewTopupRepository(db)		notificationRepo := mongorepo.NewNotificationRepository(db)
	// Add other repositories as needed

	// Initialize Services
	// Assuming AuthService needs UserRepository and config for JWT
	authService := services.NewAuthService(userRepo, cfg)
	// Assuming DrawService needs DrawRepository
	drawService := services.NewDrawService(drawRepo)
	// Assuming TopupService needs TopupRepository
	topupService := services.NewTopupService(topupRepo)
	// Assuming NotificationService needs NotificationRepository
	notificationService := services.NewNotificationService(notificationRepo)
	// Assuming UserService needs UserRepository
	userService := services.NewUserService(userRepo)
	// Add other services as needed

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(authService)
	drawHandler := handlers.NewDrawHandler(drawService) // Use the standard DrawHandler
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

	// The context is used to inform the server it has 5 seconds to finish
	// the requests it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}



