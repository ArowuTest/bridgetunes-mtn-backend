FROM golang:1.19-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/api

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
