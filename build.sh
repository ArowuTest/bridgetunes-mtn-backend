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
go get golang.org/x/crypto/bcrypt

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

# Fix the time.Now() assignment in user_service.go if it exists
echo "Checking for user_service.go to fix time.Now() assignment"
if [ -f internal/services/user_service.go ]; then
  # Create a backup of the original file
  cp internal/services/user_service.go internal/services/user_service.go.bak
  
  # Replace direct time.Now() assignment with pointer conversion if needed
  grep -q "user.OptOutDate = time.Now()" internal/services/user_service.go && \
  sed -i 's/user.OptOutDate = time.Now()/now := time.Now()\n\tuser.OptOutDate = \&now/g' internal/services/user_service.go && \
  echo "Modified user_service.go to convert time.Now() to a pointer"
fi

# Create auth directory if it doesn't exist
mkdir -p internal/auth

# Create JWT authentication helper
echo "Creating JWT authentication helper"
cat > internal/auth/jwt.go << 'EOF'
package auth

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateToken generates a JWT token for a user
func GenerateToken(userID string, role string) (string, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "bridgetunes-mtn-secure-jwt-secret-key-2025" // Default secret
	}

	// Set expiration time
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := jwt.MapClaims{
		"id":   userID,
		"role": role,
		"exp":  expirationTime.Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "bridgetunes-mtn-secure-jwt-secret-key-2025" // Default secret
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
EOF

# Create auth middleware
echo "Creating auth middleware"
mkdir -p internal/middleware

cat > internal/middleware/auth_middleware.go << 'EOF'
package middleware

import (
	"net/http"
	"strings"

	"github.com/bridgetunes/mtn-backend/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware authenticates requests
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if token is in correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in format: Bearer {token}"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user ID and role in context
		c.Set("userID", claims["id"])
		c.Set("userRole", claims["role"])
		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Check if user is admin
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
EOF

# Create CORS middleware if it doesn't exist
if [ ! -f internal/middleware/cors_middleware.go ]; then
  echo "Creating CORS middleware"
  cat > internal/middleware/cors_middleware.go << 'EOF'
package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
EOF
fi

# Create auth handlers
echo "Creating auth handlers"
mkdir -p internal/handlers

cat > internal/handlers/auth_handler.go << 'EOF'
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/auth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	db *mongo.Database
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *mongo.Database) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	// Parse request
	var req struct {
		Email       string `json:"email" binding:"required_without=Phone,omitempty,email"`
		Phone       string `json:"phone" binding:"required_without=Email,omitempty"`
		Password    string `json:"password" binding:"required,min=6"`
		FirstName   string `json:"firstName" binding:"required"`
		LastName    string `json:"lastName" binding:"required"`
		MSISDN      string `json:"msisdn"`
		OptInChannel string `json:"optInChannel"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	ctx := context.Background()
	var existingUser bson.M

	filter := bson.M{}
	if req.Email != "" {
		filter = bson.M{"email": req.Email}
	} else if req.Phone != "" {
		filter = bson.M{"phone": req.Phone}
	}

	err := h.db.Collection("users").FindOne(ctx, filter).Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	now := time.Now()
	user := bson.M{
		"email":        req.Email,
		"phone":        req.Phone,
		"password":     string(hashedPassword),
		"firstName":    req.FirstName,
		"lastName":     req.LastName,
		"role":         "user",
		"msisdn":       req.MSISDN,
		"optInStatus":  true,
		"optInDate":    now,
		"optInChannel": req.OptInChannel,
		"points":       0,
		"isBlacklisted": false,
		"lastActivity": now,
		"createdAt":    now,
		"updatedAt":    now,
	}

	result, err := h.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate token
	userID := result.InsertedID.(interface{})
	token, err := auth.GenerateToken(userID.(string), "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":           userID,
			"email":        req.Email,
			"phone":        req.Phone,
			"firstName":    req.FirstName,
			"lastName":     req.LastName,
			"role":         "user",
			"msisdn":       req.MSISDN,
			"optInStatus":  true,
			"optInDate":    now,
			"optInChannel": req.OptInChannel,
			"points":       0,
			"isBlacklisted": false,
			"lastActivity": now,
			"createdAt":    now,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	// Parse request
	var req struct {
		Email    string `json:"email" binding:"required_without=Phone,omitempty,email"`
		Phone    string `json:"phone" binding:"required_without=Email,omitempty"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	ctx := context.Background()
	var user bson.M

	filter := bson.M{}
	if req.Email != "" {
		filter = bson.M{"email": req.Email}
	} else if req.Phone != "" {
		filter = bson.M{"phone": req.Phone}
	}

	err := h.db.Collection("users").FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		}
		return
	}

	// Verify password
	storedPassword := user["password"].(string)
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login
	now := time.Now()
	_, err = h.db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": user["_id"]},
		bson.M{
			"$set": bson.M{
				"lastLoginAt":  now,
				"lastActivity": now,
				"updatedAt":    now,
			},
		},
	)

	// Generate token
	userID := user["_id"].(string)
	role := user["role"].(string)
	token, err := auth.GenerateToken(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// CreateAdmin creates an admin user
func (h *AuthHandler) CreateAdmin(c *gin.Context) {
	// Parse request
	var req struct {
		Email     string `json:"email" binding:"required_without=Phone,omitempty,email"`
		Phone     string `json:"phone" binding:"required_without=Email,omitempty"`
		Password  string `json:"password" binding:"required,min=6"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		MSISDN    string `json:"msisdn"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	ctx := context.Background()
	var existingUser bson.M

	filter := bson.M{}
	if req.Email != "" {
		filter = bson.M{"email": req.Email}
	} else if req.Phone != "" {
		filter = bson.M{"phone": req.Phone}
	}

	err := h.db.Collection("users").FindOne(ctx, filter).Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create admin user
	now := time.Now()
	admin := bson.M{
		"email":        req.Email,
		"phone":        req.Phone,
		"password":     string(hashedPassword),
		"firstName":    req.FirstName,
		"lastName":     req.LastName,
		"role":         "admin",
		"msisdn":       req.MSISDN,
		"optInStatus":  true,
		"optInDate":    now,
		"optInChannel": "admin",
		"points":       0,
		"isBlacklisted": false,
		"lastActivity": now,
		"createdAt":    now,
		"updatedAt":    now,
	}

	result, err := h.db.Collection("users").InsertOne(ctx, admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user"})
		return
	}

	// Generate token
	userID := result.InsertedID.(interface{})
	token, err := auth.GenerateToken(userID.(string), "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":           userID,
			"email":        req.Email,
			"phone":        req.Phone,
			"firstName":    req.FirstName,
			"lastName":     req.LastName,
			"role":         "admin",
			"msisdn":       req.MSISDN,
			"optInStatus":  true,
			"optInDate":    now,
			"optInChannel": "admin",
			"points":       0,
			"isBlacklisted": false,
			"lastActivity": now,
			"createdAt":    now,
		},
	})
}
EOF

# Create MongoDB connection helper
echo "Creating MongoDB connection helper"
mkdir -p internal/database

cat > internal/database/mongodb.go << 'EOF'
package database

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongoDB connects to MongoDB
func ConnectMongoDB() (*mongo.Client, error) {
	// Get MongoDB URI from environment
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb+srv://fsanus20111:wXVTvRfaCtcd5W7t@cluster0.llhkakp.mongodb.net/bridgetunes?retryWrites=true&w=majority&appName=Cluster0"
	}

	// Create client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// GetDatabase gets the MongoDB database
func GetDatabase(client *mongo.Client) *mongo.Database {
	// Get database name from environment
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "bridgetunes"
	}

	return client.Database(dbName)
}
EOF

# Update main.go to include authentication routes
echo "Updating main.go to include authentication routes"
mkdir -p cmd/api

# Check if main.go exists and create a backup
if [ -f cmd/api/main.go ]; then
  cp cmd/api/main.go cmd/api/main.go.bak
fi

cat > cmd/api/main.go << 'EOF'
package main

import (
	"log"
	"os"

	"github.com/bridgetunes/mtn-backend/internal/database"
	"github.com/bridgetunes/mtn-backend/internal/handlers"
	"github.com/bridgetunes/mtn-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Connect to MongoDB
	mongoClient, err := database.ConnectMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(nil)

	// Get database
	db := database.GetDatabase(mongoClient)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)

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
			auth.POST("/admin", authHandler.CreateAdmin) // For creating admin users
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/:id", func(c *gin.Context) {
					id := c.Param("id")
					c.JSON(200, gin.H{"message": "Get user endpoint", "id": id})
				})
				users.PUT("/:id", func(c *gin.Context) {
					id := c.Param("id")
					c.JSON(200, gin.H{"message": "Update user endpoint", "id": id})
				})
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				// Add admin-only routes here
				admin.GET("/dashboard", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Admin dashboard"})
				})
			}
		}

		// Health check route
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "API is running"})
		})
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	router.Run(":" + port)
}
EOF

# Build the application
cd ./cmd/api
go build -o ../../app .
