#!/bin/bash
set -e

# Print Go version and environment information for debugging
go version
echo "Current directory: $(pwd)"
echo "Directory contents: $(ls -la)"

# Force Go modules mode
export GO111MODULE=on

# Install all dependencies explicitly
echo "Installing dependencies..."
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

# Fix unused imports in problematic files (without modifying existing code)
echo "Fixing unused imports in problematic files..."
sed -i 
'/encoding\\/json/d' pkg/mtnapi/client.go || echo "Warning: Could not fix imports in pkg/mtnapi/client.go"
sed -i 
'/context/d' internal/middleware/middleware.go || echo "Warning: Could not fix imports in internal/middleware/middleware.go"
sed -i 
'/time/d' internal/handlers/notification_handler.go || echo "Warning: Could not fix imports in internal/handlers/notification_handler.go"
sed -i 
'/time/d' internal/handlers/user_handler.go || echo "Warning: Could not fix imports in internal/handlers/user_handler.go"

# Create necessary directories with verbose output
echo "Creating necessary directories..."
mkdir -p internal/handlers && echo "Created internal/handlers"
mkdir -p internal/models && echo "Created internal/models"
mkdir -p internal/database && echo "Created internal/database"
mkdir -p internal/middleware && echo "Created internal/middleware"
mkdir -p internal/utils && echo "Created internal/utils"
mkdir -p cmd/api && echo "Created cmd/api"

# Create a backup of main.go if it exists
if [ -f cmd/api/main.go ]; then
    cp cmd/api/main.go cmd/api/main.go.bak
    echo "Backed up existing main.go"
fi

# Create User model
echo "Creating User model..."
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
    MSISDN        string             `json:"msisdn" bson:"msisdn"`
    Password      string             `json:"-" bson:"password"`
    FirstName     string             `json:"firstName" bson:"firstName"`
    LastName      string             `json:"lastName" bson:"lastName"`
    Role          string             `json:"role" bson:"role"`
    OptInStatus   bool               `json:"optInStatus" bson:"optInStatus"`
    OptInDate     time.Time          `json:"optInDate" bson:"optInDate"`
    OptInChannel  string             `json:"optInChannel" bson:"optInChannel"`
    OptOutDate    time.Time          `json:"optOutDate" bson:"optOutDate"`
    Points        int                `json:"points" bson:"points"`
    IsBlacklisted bool               `json:"isBlacklisted" bson:"isBlacklisted"`
    LastActivity  time.Time          `json:"lastActivity" bson:"lastActivity"`
    IsVerified    bool               `json:"isVerified" bson:"isVerified"`
    CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
    UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
    LastLoginAt   time.Time          `json:"lastLoginAt,omitempty" bson:"lastLoginAt,omitempty"`
}

// UserRegistration represents the data needed for user registration
type UserRegistration struct {
    Email     string `json:"email" binding:"required,email"`
    Phone     string `json:"phone" binding:"required"`
    MSISDN    string `json:"msisdn" binding:"required"`
    Password  string `json:"password" binding:"required,min=6"`
    FirstName string `json:"firstName" binding:"required"`
    LastName  string `json:"lastName" binding:"required"`
}

// UserLogin represents the data needed for user login
type UserLogin struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// AdminCreation represents the data needed for admin creation
type AdminCreation struct {
    Email     string `json:"email" binding:"required,email"`
    Phone     string `json:"phone" binding:"required"`
    MSISDN    string `json:"msisdn" binding:"required"`
    Password  string `json:"password" binding:"required,min=6"`
    FirstName string `json:"firstName" binding:"required"`
    LastName  string `json:"lastName" binding:"required"`
}

// TokenResponse represents the response for successful authentication
type TokenResponse struct {
    Token     string `json:"token"`
    ExpiresIn int64  `json:"expiresIn"`
    Role      string `json:"role"`
    UserID    string `json:"userId"`
}
EOF
echo "User model created"

# Create transaction model
echo "Creating transaction model..."
cat > internal/models/transaction.go << 'EOF'
package models

import (
    "time"
)

// Transaction represents a recharge transaction
type Transaction struct {
    MSISDN         string    `json:"msisdn" bson:"msisdn"`
    RechargeAmount float64   `json:"rechargeAmount" bson:"rechargeAmount"`
    OptInStatus    bool      `json:"optInStatus" bson:"optInStatus"`
    RechargeDate   time.Time `json:"rechargeDate" bson:"rechargeDate"`
    Points         int       `json:"points" bson:"points"`
    CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
}

// UploadResponse represents the response from the upload endpoint
type UploadResponse struct {
    Success      bool     `json:"success"`
    Message      string   `json:"message"`
    TotalRecords int      `json:"totalRecords"`
    Inserted     int      `json:"inserted"`
    Errors       []string `json:"errors"`
}
EOF
echo "Transaction model created"

# Create JWT utils
echo "Creating JWT utilities..."
cat > internal/utils/jwt.go << 'EOF'
package utils

import (
    "fmt"
    "os"
    "time"

    "github.com/dgrijalva/jwt-go"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// JWTClaims represents the claims in the JWT
type JWTClaims struct {
    UserID string `json:"userId"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

// GenerateToken generates a new JWT token
func GenerateToken(userID primitive.ObjectID, email, role string) (string, int64, error) {
    // Get JWT secret from environment
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "bridgetunes-mtn-secret-key" // Default secret key
    }

    // Set expiration time
    expirationTime := time.Now().Add(24 * time.Hour) // 24 hours
    expiresIn := expirationTime.Unix()

    // Create claims
    claims := &JWTClaims{
        UserID: userID.Hex(),
        Email:  email,
        Role:   role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expiresIn,
            IssuedAt:  time.Now().Unix(),
            Issuer:    "bridgetunes-mtn-backend",
        },
    }

    // Create token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign token
    tokenString, err := token.SignedString([]byte(jwtSecret))
    if err != nil {
        return "", 0, err
    }

    return tokenString, expiresIn, nil
}

// ValidateToken validates a JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
    // Get JWT secret from environment
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "bridgetunes-mtn-secret-key" // Default secret key
    }

    // Parse token
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(jwtSecret), nil
    })

    if err != nil {
        return nil, err
    }

    // Validate token and extract claims
    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}
EOF
echo "JWT utilities created"

# Create authentication middleware
echo "Creating authentication middleware..."
cat > internal/middleware/auth.go << 'EOF'
package middleware

import (
    "net/http"
    "strings"

    "github.com/bridgetunes/mtn-backend/internal/utils"
    "github.com/gin-gonic/gin"
)

// AuthMiddleware authenticates requests
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // Check if header starts with Bearer
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
            c.Abort()
            return
        }

        // Extract token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate token
        claims, err := utils.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // Set user context
        c.Set("userID", claims.UserID)
        c.Set("userEmail", claims.Email)
        c.Set("userRole", claims.Role)

        c.Next()
    }
}

// AdminMiddleware checks if the user is an admin
func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("userRole")
        if !exists || role.(string) != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
            c.Abort()
            return
        }
        c.Next()
    }
}
EOF
echo "Authentication middleware created"

# Create authentication handler
echo "Creating authentication handler..."
cat > internal/handlers/auth_handler.go << 'EOF'
package handlers

import (
    "net/http"

    "github.com/bridgetunes/mtn-backend/internal/models"
    "github.com/bridgetunes/mtn-backend/internal/services"
    "github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
    authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
    return &AuthHandler{
        authService: authService,
    }
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(c *gin.Context) {
    var req models.UserRegistration
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.authService.RegisterUser(c, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, user)
}

// LoginUser handles user login
func (h *AuthHandler) LoginUser(c *gin.Context) {
    var req models.UserLogin
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    tokenResponse, err := h.authService.LoginUser(c, req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, tokenResponse)
}

// CreateAdmin handles admin creation
func (h *AuthHandler) CreateAdmin(c *gin.Context) {
    var req models.AdminCreation
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    admin, err := h.authService.CreateAdmin(c, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, admin)
}
EOF
echo "Authentication handler created"

# Create transaction handler
echo "Creating transaction handler..."
cat > internal/handlers/transaction_handler.go << 'EOF'
package handlers

import (
    "net/http"

    "github.com/bridgetunes/mtn-backend/internal/services"
    "github.com/gin-gonic/gin"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
    transactionService *services.TransactionService
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(transactionService *services.TransactionService) *TransactionHandler {
    return &TransactionHandler{
        transactionService: transactionService,
    }
}

// UploadTransactions handles CSV upload for transactions
func (h *TransactionHandler) UploadTransactions(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error: " + err.Error()})
        return
    }

    // Open the uploaded file
    src, err := file.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file: " + err.Error()})
        return
    }
    defer src.Close()

    // Process the CSV file
    response, err := h.transactionService.ProcessCSVUpload(c, src)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process CSV: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, response)
}
EOF
echo "Transaction handler created"

# Create database connection helper
echo "Creating database connection helper..."
cat > internal/database/db.go << 'EOF'
package database

import (
    "context"
    "fmt"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

var DB *mongo.Database

// ConnectDB connects to the MongoDB database
func ConnectDB() error {
    mongoURI := os.Getenv("MONGO_URI")
    dbName := os.Getenv("DB_NAME")

    if mongoURI == "" {
        return fmt.Errorf("MONGO_URI environment variable not set")
    }
    if dbName == "" {
        return fmt.Errorf("DB_NAME environment variable not set")
    }

    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.NewClient(clientOptions)
    if err != nil {
        return fmt.Errorf("failed to create mongo client: %w", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err = client.Connect(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to mongo: %w", err)
    }

    // Ping the primary
    err = client.Ping(ctx, readpref.Primary())
    if err != nil {
        return fmt.Errorf("failed to ping mongo: %w", err)
    }

    DB = client.Database(dbName)
    fmt.Println("Successfully connected to MongoDB!")
    return nil
}
EOF
echo "Database connection helper created"

# Create main.go with authentication and CSV upload
echo "Creating main.go with authentication and CSV upload..."
cat > cmd/api/main.go << 'EOF'
package main

import (
    "log"
    "os"

    "github.com/bridgetunes/mtn-backend/internal/database"
    "github.com/bridgetunes/mtn-backend/internal/handlers"
    "github.com/bridgetunes/mtn-backend/internal/middleware"
    "github.com/bridgetunes/mtn-backend/internal/repositories"
    "github.com/bridgetunes/mtn-backend/internal/services"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
    // Load .env file
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: .env file not found, using environment variables")
    }

    // Connect to database
    if err := database.ConnectDB(); err != nil {
        log.Fatalf("Could not connect to the database: %v", err)
    }

    // Initialize Gin router
    router := gin.Default()

    // CORS configuration
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000", "https://your-frontend-domain.com"} // Add your frontend URL
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
    router.Use(cors.New(config))

    // Initialize repositories
    userRepo := repositories.NewUserRepository(database.DB)
    transactionRepo := repositories.NewTransactionRepository(database.DB)

    // Initialize services
    authService := services.NewAuthService(userRepo)
    transactionService := services.NewTransactionService(transactionRepo, userRepo)

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(authService)
    transactionHandler := handlers.NewTransactionHandler(transactionService)

    // Public routes
    router.POST("/register", authHandler.RegisterUser)
    router.POST("/login", authHandler.LoginUser)

    // Authenticated routes
    authGroup := router.Group("/api")
    authGroup.Use(middleware.AuthMiddleware())
    {
        // Admin routes
        adminGroup := authGroup.Group("/admin")
        adminGroup.Use(middleware.AdminMiddleware())
        {
            adminGroup.POST("/create-admin", authHandler.CreateAdmin)
            adminGroup.POST("/upload/transactions", transactionHandler.UploadTransactions)
            // Add other admin-specific routes here
        }

        // General authenticated routes (if any)
        // authGroup.GET("/profile", ...) 
    }

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default port
    }
    log.Printf("Server starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}
EOF
echo "Main application created with authentication and CSV upload"

# Build the application
echo "Building application..."
# Use the main.go file directly for the build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bridgetunes-api ./cmd/api/main.go

echo "Build finished."

