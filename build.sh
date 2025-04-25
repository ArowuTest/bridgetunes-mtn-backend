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

# Fix the time.Now() assignment in user_service.go
echo "Fixing time.Now() assignment in user_service.go"
if [ -f internal/services/user_service.go ]; then
  # Create a backup of the original file
  cp internal/services/user_service.go internal/services/user_service.go.bak
  
  # Replace direct time.Now() assignment with pointer conversion
  sed -i 's/user.OptOutDate = time.Now()/now := time.Now()\n\tuser.OptOutDate = \&now/g' internal/services/user_service.go
  
  echo "Modified user_service.go to convert time.Now() to a pointer"
fi

# Create models directory if it doesn't exist
mkdir -p internal/models

# Create user model with all required fields
echo "Creating user model with all required fields"
cat > internal/models/user.go << 'EOF'
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email         string             `json:"email" bson:"email"`
	Phone         string             `json:"phone" bson:"phone"`
	Password      string             `json:"-" bson:"password"`
	FirstName     string             `json:"firstName" bson:"firstName"`
	LastName      string             `json:"lastName" bson:"lastName"`
	Role          string             `json:"role" bson:"role"`
	IsVerified    bool               `json:"isVerified" bson:"isVerified"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	LastLoginAt   *time.Time         `json:"lastLoginAt,omitempty" bson:"lastLoginAt,omitempty"`
	
	// Additional fields from existing model
	MSISDN        string             `json:"msisdn" bson:"msisdn"`
	OptInStatus   bool               `json:"optInStatus" bson:"optInStatus"`
	OptInDate     time.Time          `json:"optInDate" bson:"optInDate"`
	OptInChannel  string             `json:"optInChannel" bson:"optInChannel"`
	OptOutDate    *time.Time         `json:"optOutDate,omitempty" bson:"optOutDate,omitempty"`
	Points        int                `json:"points" bson:"points"`
	IsBlacklisted bool               `json:"isBlacklisted" bson:"isBlacklisted"`
	LastActivity  time.Time          `json:"lastActivity" bson:"lastActivity"`
}

// UserRegistration represents the data needed to register a new user
type UserRegistration struct {
	Email        string `json:"email" binding:"required_without=Phone,omitempty,email"`
	Phone        string `json:"phone" binding:"required_without=Email,omitempty"`
	Password     string `json:"password" binding:"required,min=6"`
	FirstName    string `json:"firstName" binding:"required"`
	LastName     string `json:"lastName" binding:"required"`
	MSISDN       string `json:"msisdn"`
	OptInChannel string `json:"optInChannel"`
}

// UserLogin represents the data needed to log in
type UserLogin struct {
	Email    string `json:"email" binding:"required_without=Phone,omitempty,email"`
	Phone    string `json:"phone" binding:"required_without=Email,omitempty"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user data returned to the client
type UserResponse struct {
	ID            string     `json:"id"`
	Email         string     `json:"email,omitempty"`
	Phone         string     `json:"phone,omitempty"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	Role          string     `json:"role"`
	MSISDN        string     `json:"msisdn,omitempty"`
	OptInStatus   bool       `json:"optInStatus"`
	OptInDate     time.Time  `json:"optInDate,omitempty"`
	OptInChannel  string     `json:"optInChannel,omitempty"`
	OptOutDate    *time.Time `json:"optOutDate,omitempty"`
	Points        int        `json:"points"`
	IsBlacklisted bool       `json:"isBlacklisted"`
	LastActivity  time.Time  `json:"lastActivity,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
EOF

# Create MongoDB client
echo "Creating MongoDB client"
mkdir -p pkg/mongodb

cat > pkg/mongodb/client.go << 'EOF'
package mongodb

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the MongoDB client
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient creates a new MongoDB client
func NewClient() (*Client, error) {
	// Get MongoDB URI from environment
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb+srv://fsanus20111:wXVTvRfaCtcd5W7t@cluster0.llhkakp.mongodb.net/bridgetunes?retryWrites=true&w=majority&appName=Cluster0"
	}

	// Get database name from environment
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "bridgetunes"
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

	// Return client
	return &Client{
		client:   client,
		database: client.Database(dbName),
	}, nil
}

// Close closes the MongoDB client
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}

// Database returns the MongoDB database
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Client returns the MongoDB client
func (c *Client) Client() *mongo.Client {
	return c.client
}
EOF

# Create repositories directory and user repository
echo "Creating user repository"
mkdir -p internal/repositories/mongodb

cat > internal/repositories/mongodb/user_repository.go << 'EOF'
package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const userCollection = "users"

// UserRepository handles user data operations
type UserRepository struct {
	client *mongodb.Client
	coll   *mongo.Collection
}

// NewUserRepository creates a new user repository
func NewUserRepository(client *mongodb.Client) *UserRepository {
	return &UserRepository{
		client: client,
		coll:   client.Database().Collection(userCollection),
	}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = r.coll.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// FindByPhone finds a user by phone
func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.coll.FindOne(ctx, bson.M{"phone": phone}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// FindByMSISDN finds a user by MSISDN
func (r *UserRepository) FindByMSISDN(ctx context.Context, msisdn string) (*models.User, error) {
	var user models.User
	err := r.coll.FindOne(ctx, bson.M{"msisdn": msisdn}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Set timestamps
	now := time.Now()
	user.Password = string(hashedPassword)
	user.CreatedAt = now
	user.UpdatedAt = now
	user.LastActivity = now
	
	// Set default role if not specified
	if user.Role == "" {
		user.Role = "user"
	}

	// Insert user
	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	// Set ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}

	return nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	_, err := r.coll.ReplaceOne(
		ctx,
		bson.M{"_id": user.ID},
		user,
	)

	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = r.coll.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"lastLoginAt": now,
				"lastActivity": now,
				"updatedAt": now,
			},
		},
	)

	return err
}

// VerifyPassword verifies a password against a hash
func (r *UserRepository) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
EOF

# Create auth handler
echo "Creating auth handler"
mkdir -p internal/handlers

cat > internal/handlers/auth_handler.go << 'EOF'
package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories/mongodb"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userRepo *mongodb.UserRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *mongodb.UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	ctx := context.Background()
	var existingUser *models.User
	var err error

	if req.Email != "" {
		existingUser, err = h.userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
			return
		}
	}

	if req.Phone != "" {
		existingUser, err = h.userRepo.FindByPhone(ctx, req.Phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already registered"})
			return
		}
	}

	if req.MSISDN != "" {
		existingUser, err = h.userRepo.FindByMSISDN(ctx, req.MSISDN)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MSISDN already registered"})
			return
		}
	}

	// Create user
	now := time.Now()
	user := &models.User{
		Email:        req.Email,
		Phone:        req.Phone,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         "user", // Default role is user
		MSISDN:       req.MSISDN,
		OptInStatus:  true,
		OptInDate:    now,
		OptInChannel: req.OptInChannel,
		Points:       0,
		IsBlacklisted: false,
		LastActivity: now,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:           user.ID.Hex(),
			Email:        user.Email,
			Phone:        user.Phone,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Role:         user.Role,
			MSISDN:       user.MSISDN,
			OptInStatus:  user.OptInStatus,
			OptInDate:    user.OptInDate,
			OptInChannel: user.OptInChannel,
			Points:       user.Points,
			IsBlacklisted: user.IsBlacklisted,
			LastActivity: user.LastActivity,
			CreatedAt:    user.CreatedAt,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	ctx := context.Background()
	var user *models.User
	var err error

	if req.Email != "" {
		user, err = h.userRepo.FindByEmail(ctx, req.Email)
	} else if req.Phone != "" {
		user, err = h.userRepo.FindByPhone(ctx, req.Phone)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email or phone is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if !h.userRepo.VerifyPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login
	if err := h.userRepo.UpdateLastLogin(ctx, user.ID.Hex()); err != nil {
		// Non-critical error, just log it
		// log.Printf("Failed to update last login: %v", err)
	}

	// Generate token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:           user.ID.Hex(),
			Email:        user.Email,
			Phone:        user.Phone,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Role:         user.Role,
			MSISDN:       user.MSISDN,
			OptInStatus:  user.OptInStatus,
			OptInDate:    user.OptInDate,
			OptInChannel: user.OptInChannel,
			OptOutDate:   user.OptOutDate,
			Points:       user.Points,
			IsBlacklisted: user.IsBlacklisted,
			LastActivity: user.LastActivity,
			CreatedAt:    user.CreatedAt,
		},
	})
}

// CreateAdmin creates an admin user (for development/testing)
func (h *AuthHandler) CreateAdmin(c *gin.Context) {
	var req models.UserRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	ctx := context.Background()
	var existingUser *models.User
	var err error

	if req.Email != "" {
		existingUser, err = h.userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
			return
		}
	}

	if req.Phone != "" {
		existingUser, err = h.userRepo.FindByPhone(ctx, req.Phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone already registered"})
			return
		}
	}

	if req.MSISDN != "" {
		existingUser, err = h.userRepo.FindByMSISDN(ctx, req.MSISDN)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MSISDN already registered"})
			return
		}
	}

	// Create admin user
	now := time.Now()
	user := &models.User{
		Email:        req.Email,
		Phone:        req.Phone,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         "admin", // Set role as admin
		MSISDN:       req.MSISDN,
		OptInStatus:  true,
		OptInDate:    now,
		OptInChannel: req.OptInChannel,
		Points:       0,
		IsBlacklisted: false,
		LastActivity: now,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user"})
		return
	}

	// Generate token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:           user.ID.Hex(),
			Email:        user.Email,
			Phone:        user.Phone,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Role:         user.Role,
			MSISDN:       user.MSISDN,
			OptInStatus:  user.OptInStatus,
			OptInDate:    user.OptInDate,
			OptInChannel: user.OptInChannel,
			OptOutDate:   user.OptOutDate,
			Points:       user.Points,
			IsBlacklisted: user.IsBlacklisted,
			LastActivity: user.LastActivity,
			CreatedAt:    user.CreatedAt,
		},
	})
}

// generateToken generates a JWT token for a user
func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "bridgetunes-mtn-secure-jwt-secret-key-2025" // Default secret
	}

	// Set expiration time
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := jwt.MapClaims{
		"id":   user.ID.Hex(),
		"role": user.Role,
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
EOF

# Create auth middleware
echo "Creating auth middleware"
mkdir -p internal/middleware

cat > internal/middleware/auth_middleware.go << 'EOF'
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
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

		// Get JWT secret from environment
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "bridgetunes-mtn-secure-jwt-secret-key-2025" // Default secret
		}

		// Parse token
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Validate token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Set user ID and role in context
			c.Set("userID", claims["id"])
			c.Set("userRole", claims["role"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
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

# Create CORS middleware
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

# Create main.go with authentication routes
echo "Creating main.go with authentication routes"
mkdir -p cmd/api

cat > cmd/api/main.go << 'EOF'
package main

import (
	"log"
	"os"

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
	defer mongoClient.Close()

	// Initialize repositories
	userRepo := mongodb.NewUserRepository(mongoClient)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo)

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
			auth.POST("/admin", authHandler.CreateAdmin) // For creating admin users (can be protected in production)
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
