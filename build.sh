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

# Build the application
cd ./cmd/api
go build -o ../../app .

