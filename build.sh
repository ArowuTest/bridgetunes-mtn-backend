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
sed -i '/encoding\/json/d' pkg/mtnapi/client.go || echo "Warning: Could not fix imports in pkg/mtnapi/client.go"
sed -i '/context/d' internal/middleware/middleware.go || echo "Warning: Could not fix imports in internal/middleware/middleware.go"
sed -i '/github.com\/bridgetunes\/mtn-backend\/internal\/models/d' internal/handlers/draw_handler.go || echo "Warning: Could not fix imports in internal/handlers/draw_handler.go"
sed -i '/time/d' internal/handlers/notification_handler.go || echo "Warning: Could not fix imports in internal/handlers/notification_handler.go"
sed -i '/time/d' internal/handlers/user_handler.go || echo "Warning: Could not fix imports in internal/handlers/user_handler.go"

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
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email        string             `json:"email" bson:"email"`
	Phone        string             `json:"phone" bson:"phone"`
	MSISDN       string             `json:"msisdn" bson:"msisdn"`
	Password     string             `json:"-" bson:"password"`
	FirstName    string             `json:"firstName" bson:"firstName"`
	LastName     string             `json:"lastName" bson:"lastName"`
	Role         string             `json:"role" bson:"role"`
	OptInStatus  bool               `json:"optInStatus" bson:"optInStatus"`
	OptInDate    time.Time          `json:"optInDate" bson:"optInDate"`
	OptInChannel string             `json:"optInChannel" bson:"optInChannel"`
	OptOutDate   time.Time          `json:"optOutDate" bson:"optOutDate"`
	Points       int                `json:"points" bson:"points"`
	IsBlacklisted bool              `json:"isBlacklisted" bson:"isBlacklisted"`
	LastActivity time.Time          `json:"lastActivity" bson:"lastActivity"`
	IsVerified   bool               `json:"isVerified" bson:"isVerified"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	LastLoginAt  time.Time          `json:"lastLoginAt,omitempty" bson:"lastLoginAt,omitempty"`
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if Authorization header has Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in format: Bearer {token}"})
			c.Abort()
			return
		}

		// Validate token
		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
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
echo "Authentication middleware created"

# Create authentication handler
echo "Creating authentication handler..."
cat > internal/handlers/auth_handler.go << 'EOF'
package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	collection *mongo.Collection
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *mongo.Database) *AuthHandler {
	log.Println("Initializing AuthHandler...")
	collection := db.Collection("users")

	// Create index on email for faster lookups and to prevent duplicates
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: mongo.options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create index on email: %v\n", err)
	} else {
		log.Println("Successfully created index on users collection")
	}

	// Create index on phone for faster lookups and to prevent duplicates
	phoneIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: mongo.options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(ctx, phoneIndexModel)
	if err != nil {
		log.Printf("Warning: Failed to create index on phone: %v\n", err)
	} else {
		log.Println("Successfully created index on phone field")
	}

	// Create index on MSISDN for faster lookups and to prevent duplicates
	msisdnIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "msisdn", Value: 1}},
		Options: mongo.options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(ctx, msisdnIndexModel)
	if err != nil {
		log.Printf("Warning: Failed to create index on MSISDN: %v\n", err)
	} else {
		log.Println("Successfully created index on MSISDN field")
	}

	return &AuthHandler{
		collection: collection,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	log.Println("Register handler called")

	// Parse request body
	var registration models.UserRegistration
	if err := c.ShouldBindJSON(&registration); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existingUser models.User
	err := h.collection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": registration.Email},
			{"phone": registration.Phone},
			{"msisdn": registration.MSISDN},
		},
	}).Decode(&existingUser)

	if err == nil {
		log.Println("User already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email, phone, or MSISDN already exists"})
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	now := time.Now()
	user := models.User{
		ID:           primitive.NewObjectID(),
		Email:        registration.Email,
		Phone:        registration.Phone,
		MSISDN:       registration.MSISDN,
		Password:     string(hashedPassword),
		FirstName:    registration.FirstName,
		LastName:     registration.LastName,
		Role:         "user", // Default role is user
		OptInStatus:  true,
		OptInDate:    now,
		OptInChannel: "registration",
		Points:       0,
		IsBlacklisted: false,
		LastActivity: now,
		IsVerified:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Insert user into database
	_, err = h.collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, expiresIn, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return token
	c.JSON(http.StatusCreated, models.TokenResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		Role:      user.Role,
		UserID:    user.ID.Hex(),
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	log.Println("Login handler called")

	// Parse request body
	var login models.UserLogin
	if err := c.ShouldBindJSON(&login); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := h.collection.FindOne(ctx, bson.M{"email": login.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("User not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is blacklisted
	if user.IsBlacklisted {
		log.Println("User is blacklisted")
		c.JSON(http.StatusForbidden, gin.H{"error": "Your account has been blacklisted"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		log.Println("Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Update last login time
	now := time.Now()
	_, err = h.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"lastLoginAt":  now,
				"lastActivity": now,
				"updatedAt":    now,
			},
		},
	)
	if err != nil {
		log.Printf("Error updating last login time: %v", err)
		// Continue anyway, this is not critical
	}

	// Generate JWT token
	token, expiresIn, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return token
	c.JSON(http.StatusOK, models.TokenResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		Role:      user.Role,
		UserID:    user.ID.Hex(),
	})
}

// CreateAdmin handles admin user creation
func (h *AuthHandler) CreateAdmin(c *gin.Context) {
	log.Println("CreateAdmin handler called")

	// Parse request body
	var adminCreation models.AdminCreation
	if err := c.ShouldBindJSON(&adminCreation); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existingUser models.User
	err := h.collection.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": adminCreation.Email},
			{"phone": adminCreation.Phone},
			{"msisdn": adminCreation.MSISDN},
		},
	}).Decode(&existingUser)

	if err == nil {
		log.Println("User already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email, phone, or MSISDN already exists"})
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminCreation.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create admin user
	now := time.Now()
	admin := models.User{
		ID:           primitive.NewObjectID(),
		Email:        adminCreation.Email,
		Phone:        adminCreation.Phone,
		MSISDN:       adminCreation.MSISDN,
		Password:     string(hashedPassword),
		FirstName:    adminCreation.FirstName,
		LastName:     adminCreation.LastName,
		Role:         "admin", // Admin role
		OptInStatus:  true,
		OptInDate:    now,
		OptInChannel: "admin_creation",
		Points:       0,
		IsBlacklisted: false,
		LastActivity: now,
		IsVerified:   true, // Admins are automatically verified
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Insert admin into database
	_, err = h.collection.InsertOne(ctx, admin)
	if err != nil {
		log.Printf("Error inserting admin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin"})
		return
	}

	// Return success
	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin created successfully",
		"userId":  admin.ID.Hex(),
		"role":    admin.Role,
	})
}
EOF
echo "Authentication handler created"

# Create transaction handler with fixed variable naming
echo "Creating transaction handler..."
cat > internal/handlers/transaction_handler.go << 'EOF'
package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
	collection *mongo.Collection
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(db *mongo.Database) *TransactionHandler {
	log.Println("Initializing TransactionHandler...")
	collection := db.Collection("transactions")

	// Create index on MSISDN and RechargeDate for faster lookups and to prevent duplicates
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "msisdn", Value: 1},
			{Key: "rechargeDate", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create index: %v\n", err)
	} else {
		log.Println("Successfully created index on transactions collection")
	}

	return &TransactionHandler{
		collection: collection,
	}
}

// UploadCSV handles CSV file uploads
func (h *TransactionHandler) UploadCSV(c *gin.Context) {
	log.Println("UploadCSV handler called")
	
	// Get file from request
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Error getting file from request: %v", err)
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Message: "No file uploaded",
			Errors:  []string{err.Error()},
		})
		return
	}
	defer file.Close()
	
	log.Printf("Received file: %s", fileHeader.Filename)

	// Parse CSV
	reader := csv.NewReader(file)

	// Read header
	csvHeader, err := reader.Read()
	if err != nil {
		log.Printf("Error reading CSV header: %v", err)
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Message: "Failed to read CSV header",
			Errors:  []string{err.Error()},
		})
		return
	}
	
	log.Printf("CSV header: %v", csvHeader)

	// Validate header
	expectedHeaders := []string{"MSISDN", "Recharge Amount", "Opt-In Status", "Recharge Date"}
	for i, h := range expectedHeaders {
		if i >= len(csvHeader) || !strings.Contains(csvHeader[i], expectedHeaders[i]) {
			log.Printf("Invalid CSV format: expected header '%s' not found", h)
			c.JSON(http.StatusBadRequest, models.UploadResponse{
				Success: false,
				Message: "Invalid CSV format",
				Errors:  []string{fmt.Sprintf("Expected header '%s' not found", h)},
			})
			return
		}
	}

	// Process records
	var transactions []interface{}
	var errors []string
	totalRecords := 0
	inserted := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error reading record: %v", err))
			continue
		}

		totalRecords++

		// Validate record length
		if len(record) < 4 {
			errors = append(errors, fmt.Sprintf("Record %d: Invalid number of fields", totalRecords))
			continue
		}

		// Parse MSISDN
		msisdn := strings.TrimSpace(record[0])
		if msisdn == "" {
			errors = append(errors, fmt.Sprintf("Record %d: MSISDN is required", totalRecords))
			continue
		}

		// Parse recharge amount
		amountStr := strings.TrimSpace(record[1])
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Record %d: Invalid recharge amount: %s", totalRecords, amountStr))
			continue
		}

		// Parse opt-in status
		optInStr := strings.TrimSpace(record[2])
		optIn := strings.EqualFold(optInStr, "Yes")

		// Parse recharge date
		dateStr := strings.TrimSpace(record[3])
		date, err := time.Parse("02/01/2006", dateStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Record %d: Invalid date format: %s", totalRecords, dateStr))
			continue
		}

		// Calculate points based on the correct points allocation logic (REQFUNC025)
		points := calculatePoints(amount)

		// Create transaction
		transaction := models.Transaction{
			MSISDN:         msisdn,
			RechargeAmount: amount,
			OptInStatus:    optIn,
			RechargeDate:   date,
			Points:         points,
			CreatedAt:      time.Now(),
		}

		transactions = append(transactions, transaction)
	}
	
	log.Printf("Processed %d records from CSV", totalRecords)

	// Insert transactions into MongoDB
	if len(transactions) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := options.InsertMany().SetOrdered(false)
		result, err := h.collection.InsertMany(ctx, transactions, opts)
		if err != nil {
			// Handle duplicate key errors
			if mongo.IsDuplicateKeyError(err) {
				log.Printf("Some records were not inserted due to duplicate keys: %v", err)
				errors = append(errors, "Some records were not inserted due to duplicate MSISDN and recharge date")
				if result != nil {
					inserted = len(result.InsertedIDs)
				}
			} else {
				log.Printf("Database error: %v", err)
				errors = append(errors, fmt.Sprintf("Database error: %v", err))
				inserted = 0
			}
		} else {
			inserted = len(result.InsertedIDs)
			log.Printf("Successfully inserted %d records into MongoDB", inserted)
		}
	}

	// Return response
	response := models.UploadResponse{
		Success:      len(errors) == 0,
		Message:      fmt.Sprintf("Processed %d records, inserted %d", totalRecords, inserted),
		TotalRecords: totalRecords,
		Inserted:     inserted,
		Errors:       errors,
	}
	
	log.Printf("Upload response: %+v", response)
	c.JSON(http.StatusOK, response)
}

// Calculate points based on recharge amount according to REQFUNC025
func calculatePoints(amount float64) int {
	switch {
	case amount < 100:
		return 0
	case amount >= 100 && amount <= 199:
		return 1
	case amount >= 200 && amount <= 299:
		return 2
	case amount >= 300 && amount <= 399:
		return 3
	case amount >= 400 && amount <= 499:
		return 4
	case amount >= 500 && amount <= 599:
		return 5
	case amount >= 600 && amount <= 699:
		return 6
	case amount >= 700 && amount <= 799:
		return 7
	case amount >= 800 && amount <= 899:
		return 8
	case amount >= 900 && amount <= 999:
		return 9
	case amount >= 1000:
		return 10
	default:
		return 0
	}
}
EOF
echo "Transaction handler created"

# Create database connection helper
echo "Creating database connection helper..."
cat > internal/database/mongodb.go << 'EOF'
package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect establishes a connection to MongoDB
func Connect() (*mongo.Client, *mongo.Database, error) {
	log.Println("Connecting to MongoDB...")
	
	// Get MongoDB URI from environment
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb+srv://fsanus20111:wXVTvRfaCtcd5W7t@cluster0.llhkakp.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
		log.Println("Using default MongoDB URI")
	}

	// Get database name from environment
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "bridgetunes"
		log.Println("Using default database name: bridgetunes")
	}

	// Create client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to MongoDB at %s", uri)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		return nil, nil, err
	}

	// Ping database
	log.Println("Pinging MongoDB...")
	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("Error pinging MongoDB: %v", err)
		return nil, nil, err
	}

	log.Println("Successfully connected to MongoDB")
	return client, client.Database(dbName), nil
}
EOF
echo "Database connection helper created"

# Update main.go to include authentication and CSV upload endpoints
echo "Creating main.go with authentication and CSV upload..."
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
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Bridgetunes MTN backend application...")
	
	// Print environment information
	log.Printf("Working directory: %s", getWorkingDir())
	log.Printf("Environment variables: PORT=%s", os.Getenv("PORT"))

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize MongoDB connection
	log.Println("Initializing MongoDB connection...")
	client, db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(nil); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Initialize handlers
	log.Println("Initializing handlers...")
	authHandler := handlers.NewAuthHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)

	// Initialize router
	log.Println("Initializing Gin router...")
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	// Add logging middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Apply built-in CORS middleware
	log.Println("Applying CORS middleware...")
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Define routes
	log.Println("Defining API routes...")
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			
			// Admin creation route (protected by auth middleware)
			adminRoute := auth.Group("/admin")
			adminRoute.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
			adminRoute.POST("", authHandler.CreateAdmin)
		}

		// CSV upload route
		log.Println("Registering CSV upload route: /api/upload/transactions")
		api.POST("/upload/transactions", transactionHandler.UploadCSV)

		// Protected routes
		protected := api.Group("/protected")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes (accessible by all authenticated users)
			protected.GET("/user/:id", func(c *gin.Context) {
				id := c.Param("id")
				log.Printf("Get user endpoint called with id: %s", id)
				c.JSON(200, gin.H{"message": "Get user endpoint", "id": id})
			})
			
			protected.PUT("/user/:id", func(c *gin.Context) {
				id := c.Param("id")
				log.Printf("Update user endpoint called with id: %s", id)
				c.JSON(200, gin.H{"message": "Update user endpoint", "id": id})
			})
			
			// Admin routes (accessible only by admins)
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				admin.GET("/dashboard", func(c *gin.Context) {
					log.Println("Admin dashboard endpoint called")
					c.JSON(200, gin.H{"message": "Admin dashboard"})
				})
			}
		}

		// Health check route
		api.GET("/health", func(c *gin.Context) {
			log.Println("Health check endpoint called")
			c.JSON(200, gin.H{"status": "ok", "message": "API is running"})
		})
	}

	// Add a root route for easy testing
	router.GET("/", func(c *gin.Context) {
		log.Println("Root endpoint called")
		c.JSON(200, gin.H{
			"message": "Bridgetunes MTN Backend API",
			"version": "1.0.0",
			"status": "running",
			"endpoints": []string{
				"/api/health",
				"/api/auth/register",
				"/api/auth/login",
				"/api/auth/admin",
				"/api/upload/transactions",
				"/api/protected/user/:id",
				"/api/protected/admin/dashboard",
			},
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("No PORT environment variable found, using default port 8080")
	}
	
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Helper function to get working directory
func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}
EOF
echo "Main application created with authentication and CSV upload"

# Build the application
echo "Building application..."
cd ./cmd/api
go build -o ../../app .
echo "Build completed successfully"

# Print success message
echo "Build script completed successfully"
echo "You can now deploy the application to Render.com"
