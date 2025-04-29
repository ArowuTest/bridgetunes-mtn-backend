# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# --- NEW: Remove potential cached modules before downloading ---
RUN rm -rf /go/pkg/mod

# Download dependencies
# Using go mod download & verify ensures dependencies are fetched and consistent
RUN go mod download && go mod verify

# Copy the entire source code
# Ensure you don't have a .dockerignore file excluding necessary code
COPY . .

# Explicitly run go mod tidy to ensure module consistency
RUN go mod tidy

# Clean Go build cache before building
RUN go clean -cache

# Build the application from the root directory using the full package path
# Output the binary to /app/bridgetunes-api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bridgetunes-api github.com/bridgetunes/mtn-backend/cmd/api

# Stage 2: Create the final lightweight image
FROM alpine:latest

WORKDIR /app

# Install necessary packages like ca-certificates for HTTPS and tzdata for timezones
RUN apk --no-cache add ca-certificates tzdata

# Copy only the built binary from the builder stage
COPY --from=builder /app/bridgetunes-api .

# Expose the port the application runs on
EXPOSE 8080

# Define the command to run the application
# Use the binary name specified in the build stage
CMD ["./bridgetunes-api"]
