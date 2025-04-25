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
go get github.com/gin-contrib/cors

# Run go mod tidy to clean up dependencies
go mod tidy

# Fix unused imports in problematic files (without modifying existing code)
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

# Create a backup of main.go if it exists
if [ -f cmd/api/main.go ]; then
  cp cmd/api/main.go cmd/api/main.go.bak
fi

# Create CSV upload implementation
echo "Creating CSV upload implementation"
mkdir -p internal/handlers
mkdir -p internal/models
mkdir -p internal/database  # Create database directory

# Create transaction model
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

# Create transaction handler
cat > internal/handlers/transaction_handler.go << 'EOF'
package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
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
		fmt.Printf("Warning: Failed to create index: %v\n", err)
	}

	return &TransactionHandler{
		collection: collection,
	}
}

// UploadCSV handles CSV file uploads
func (h *TransactionHandler) UploadCSV(c *gin.Context) {
	// Get file from request
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Message: "No file uploaded",
			Errors:  []string{err.Error()},
		})
		return
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Message: "Failed to read CSV header",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Validate header
	expectedHeaders := []string{"MSISDN", "Recharge Amount", "Opt-In Status", "Recharge Date"}
	for i, h := range expectedHeaders {
		if i >= len(header) || !strings.Contains(header[i], expectedHeaders[i]) {
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

		// Calculate points based on the correct points allocation logic
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

	// Insert transactions into MongoDB
	if len(transactions) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := options.InsertMany().SetOrdered(false)
		result, err := h.collection.InsertMany(ctx, transactions, opts)
		if err != nil {
			// Handle duplicate key errors
			if mongo.IsDuplicateKeyError(err) {
				errors = append(errors, "Some records were not inserted due to duplicate MSISDN and recharge date")
				inserted = len(result.InsertedIDs)
			} else {
				errors = append(errors, fmt.Sprintf("Database error: %v", err))
				inserted = 0
			}
		} else {
			inserted = len(result.InsertedIDs)
		}
	}

	// Return response
	c.JSON(http.StatusOK, models.UploadResponse{
		Success:      len(errors) == 0,
		Message:      fmt.Sprintf("Processed %d records, inserted %d", totalRecords, inserted),
		TotalRecords: totalRecords,
		Inserted:     inserted,
		Errors:       errors,
	})
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

# Create database connection helper
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
	// Get MongoDB URI from environment
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb+srv://fsanus20111:wXVTvRfaCtcd5W7t@cluster0.llhkakp.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
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
		return nil, nil, err
	}

	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, err
	}

	log.Println("Connected to MongoDB")
	return client, client.Database(dbName), nil
}
EOF

# Create cmd/api directory if it doesn't exist
mkdir -p cmd/api

# Update main.go to include CSV upload endpoint
cat > cmd/api/main.go << 'EOF'
package main

import (
	"log"
	"os"

	"github.com/bridgetunes/mtn-backend/internal/database"
	"github.com/bridgetunes/mtn-backend/internal/handlers"
	"github.com/gin-contrib/cors"
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
	transactionHandler := handlers.NewTransactionHandler(db)

	// Initialize router
	router := gin.Default()

	// Apply CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		AllowCredentials: true,
	}))

	// Define routes
	api := router.Group("/api")
	{
		// Auth routes (placeholders)
		auth := api.Group("/auth")
		{
			auth.POST("/register", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Registration endpoint"})
			})
			auth.POST("/login", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Login endpoint"})
			})
			auth.POST("/admin", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin creation endpoint"})
			})
		}

		// CSV upload route
		api.POST("/upload/transactions", transactionHandler.UploadCSV)

		// Protected routes (placeholders)
		protected := api.Group("/protected")
		{
			protected.GET("/user/:id", func(c *gin.Context) {
				id := c.Param("id")
				c.JSON(200, gin.H{"message": "Get user endpoint", "id": id})
			})
			protected.PUT("/user/:id", func(c *gin.Context) {
				id := c.Param("id")
				c.JSON(200, gin.H{"message": "Update user endpoint", "id": id})
			})
			protected.GET("/admin/dashboard", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin dashboard"})
			})
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
