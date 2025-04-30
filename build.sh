#!/bin/bash
set -e

# Print Go version and environment information for debugging
go version
echo "Current directory: $(pwd)"
echo "Directory contents: $(ls -la)"

# Force Go modules mode
export GO111MODULE=on

# Install all dependencies explicitly
# Note: 'go get' is generally discouraged for just installing dependencies in newer Go versions.
# 'go mod download' or letting the build command handle it is preferred.
# However, keeping 'go get' for now as it was in the original script.
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

# Removed the problematic sed commands for fixing imports
# echo "Fixing unused imports in problematic files..."
# sed -i ... (lines removed)

# Removed all 'cat > ... << EOF' blocks that were generating Go code.
# The build will now use the Go files directly from the repository.

echo "Ensuring necessary directories exist (though build should handle this)..."
mkdir -p internal/handlers
mkdir -p internal/models
mkdir -p internal/database
mkdir -p internal/middleware
mkdir -p internal/utils
mkdir -p internal/services # Added services dir just in case
mkdir -p cmd/api

# Build the application using the code from the repository
echo "Building application from repository source..."
# Use the main.go file directly for the build
CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o bridgetunes-api ./cmd/api/main.go

echo "Build finished."

