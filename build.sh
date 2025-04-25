#!/bin/bash
set -e

# Print Go version
go version

# Force Go modules mode
export GO111MODULE=on

# Install all dependencies explicitly
go get github.com/gin-gonic/gin
go get github.com/joho/godotenv
go get github.com/dgrijalva/jwt-go
go get github.com/spf13/viper
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
go get go.mongodb.org/mongo-driver/mongo/options

# Run go mod tidy to clean up dependencies
go mod tidy

# Fix unused imports in problematic files
echo "Fixing unused imports in pkg/mtnapi/client.go"
sed -i '/encoding\/json/d' pkg/mtnapi/client.go || true

echo "Fixing unused imports in internal/middleware/middleware.go"
sed -i '/context/d' internal/middleware/middleware.go || true

echo "Fixing unused imports in internal/handlers/draw_handler.go"
sed -i '/github.com\/bridgetunes\/mtn-backend\/internal\/models/d' internal/handlers/draw_handler.go || true

echo "Fixing unused imports in internal/handlers/notification_handler.go"
sed -i '/time/d' internal/handlers/notification_handler.go || true

echo "Fixing unused imports in internal/handlers/user_handler.go"
sed -i '/time/d' internal/handlers/user_handler.go || true

# Fix main.go issues
echo "Fixing main.go issues"
# Create a temporary file with fixed content
cat > cmd/api/main.go.fixed << 'EOF'
package main

import (
	"log"
	"os"

	"github.com/bridgetunes/mtn-backend/internal/config"
	"github.com/bridgetunes/mtn-backend/internal/handlers"
	"github.com/bridgetunes/mtn-backend/internal/middleware"
	"github.com/bridgetunes/mtn-backend/internal/repositories/mongodb"
	"github.com/bridgetunes/mtn-backend/pkg/mongodb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize MongoDB connection
	mongoClient, err := mongodb.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Initialize repositories
	userRepo := mongodb.NewUserRepository(mongoClient)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	userHandler := handlers.NewUserHandler(userRepo)

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Define routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// User routes
		user := api.Group("/users")
		{
			user.GET("/:id", userHandler.GetUser)
			user.PUT("/:id", userHandler.UpdateUser)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
EOF

# Replace the original file with the fixed one
mv cmd/api/main.go.fixed cmd/api/main.go

# Build the application
cd ./cmd/api
go build -o ../../app .

