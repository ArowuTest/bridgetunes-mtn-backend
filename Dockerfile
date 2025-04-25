FROM golang:1.19-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod ./

# Download all dependencies and verify
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Force Go modules mode and run go mod tidy to ensure all dependencies are properly listed
RUN go mod tidy

# Build the application with explicit dependency resolution
RUN cd ./cmd/api && GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ../../main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
