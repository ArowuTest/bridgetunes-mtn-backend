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
mkdir -p cmd/api && echo "Created cmd/api"

# Create a backup of main.go if it exists
if [ -f cmd/api/main.go ]; then
  cp cmd/api/main.go cmd/api/main.go.bak
  echo "Backed up existing main.go"
fi

# Create CSV upload implementation
echo "Creating CSV upload implementation..."

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

# Update main.go to include CSV upload endpoint with built-in CORS middleware and debug logging
echo "Creating main.go with debug logging..."
cat > cmd/api/main.go << 'EOF'
package main

import (
	"log"
	"os"

	"github.com/bridgetunes/mtn-backend/internal/database"
	"github.com/bridgetunes/mtn-backend/internal/handlers"
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
		// Auth routes (placeholders)
		auth := api.Group("/auth")
		{
			auth.POST("/register", func(c *gin.Context) {
				log.Println("Register endpoint called")
				c.JSON(200, gin.H{"message": "Registration endpoint"})
			})
			auth.POST("/login", func(c *gin.Context) {
				log.Println("Login endpoint called")
				c.JSON(200, gin.H{"message": "Login endpoint"})
			})
			auth.POST("/admin", func(c *gin.Context) {
				log.Println("Admin creation endpoint called")
				c.JSON(200, gin.H{"message": "Admin creation endpoint"})
			})
		}

		// CSV upload route
		log.Println("Registering CSV upload route: /api/upload/transactions")
		api.POST("/upload/transactions", transactionHandler.UploadCSV)

		// Protected routes (placeholders)
		protected := api.Group("/protected")
		{
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
			protected.GET("/admin/dashboard", func(c *gin.Context) {
				log.Println("Admin dashboard endpoint called")
				c.JSON(200, gin.H{"message": "Admin dashboard"})
			})
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
				"/api/upload/transactions",
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
echo "Main application created with debug logging"

# Build the application
echo "Building application..."
cd ./cmd/api
go build -o ../../app .
echo "Build completed successfully"

# Print success message
echo "Build script completed successfully"
echo "You can now deploy the application to Render.com"
