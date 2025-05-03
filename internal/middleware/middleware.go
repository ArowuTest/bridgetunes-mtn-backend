package middleware

import (
	"errors"
	"fmt" // Added fmt import
	"log" // Added log import
	"net/http"
	"strings"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5" // Use the newer library
)

// JWTAuthMiddleware is a middleware for JWT authentication
func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Check if JWT secret is configured. If not, log a fatal error as middleware cannot function.
	 if cfg.JWT.Secret == "" {
	 	log.Fatal("[FATAL] JWTAuthMiddleware: JWT_SECRET is not configured in the environment variables!")
	 }
	 jwtSecret := []byte(cfg.JWT.Secret)

	 return func(c *gin.Context) {
	 	const BearerSchema = "Bearer "
	 	 authHeader := c.GetHeader("Authorization")
	 	 if authHeader == "" {
	 	 	log.Println("[WARN] JWTAuthMiddleware: Authorization header is missing")
	 	 	 c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
	 	 	return
	 	 }

	 	 if !strings.HasPrefix(authHeader, BearerSchema) {
	 	 	log.Println("[WARN] JWTAuthMiddleware: Authorization header format is invalid")
	 	 	 c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with Bearer "})
	 	 	return
	 	 }

	 	 tokenString := authHeader[len(BearerSchema):]

	 	 token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	 	 	// Validate the alg is what you expect:
	 	 	 if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	 	 	 	return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	 	 	 }
	 	 	return jwtSecret, nil
	 	 })

	 	 if err != nil {
	 	 	log.Printf("[WARN] JWTAuthMiddleware: Token parsing/validation failed: %v", err)
	 	 	// Handle specific errors like expiration
	 	 	 if errors.Is(err, jwt.ErrTokenExpired) {
	 	 	 	 c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
	 	 	 } else {
	 	 	 	 c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()}) // Include error details
	 	 	 }
	 	 	return
	 	 }

	 	 if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
	 	 	// Token is valid. Optionally set claims in context for downstream handlers.
	 	 	 c.Set("userID", claims["sub"])
	 	 	 c.Set("userEmail", claims["email"])
	 	 	 c.Set("userRole", claims["role"])
	 	 	log.Printf("[DEBUG] JWTAuthMiddleware: Token validated successfully for user %s", claims["email"])
	 	 	 c.Next() // Proceed to the next handler
	 	 } else {
	 	 	log.Println("[WARN] JWTAuthMiddleware: Token claims invalid or token is not valid")
	 	 	 c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
	 	 }
	 }
}

// CORSMiddleware is a middleware for CORS
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		// Ensure AllowedHosts configuration is correct and secure
		// Consider using a more robust CORS library if complex rules are needed
		// e.g., github.com/gin-contrib/cors
		 c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(cfg.Server.AllowedHosts, ","))
		 c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		 c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		 c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle preflight requests (OPTIONS method)
		 if c.Request.Method == "OPTIONS" {
			 c.AbortWithStatus(http.StatusNoContent) // Use 204 No Content for OPTIONS response
			 return
		}

		 c.Next()
	}
}

// RequestIDMiddleware is a middleware for adding a request ID to the context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header or generate a new one
		 requestID := c.GetHeader("X-Request-ID")
		 if requestID == "" {
			// Generate a unique request ID (consider using UUID library for better uniqueness)
			// Example using timestamp + IP might have collisions under high load
			// import "github.com/google/uuid"
			// requestID = uuid.New().String()
			// Using current format for now:
			 requestID = time.Now().Format("20060102150405.000000") + "-" + c.ClientIP()
		}
		// Set request ID in context and response header
		 c.Set("RequestID", requestID)
		 c.Writer.Header().Set("X-Request-ID", requestID)
		 c.Next()
	}
}

// LoggerMiddleware is a middleware for logging requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		 start := time.Now()

		// Process request
		 c.Next()

		// Calculate latency
		 latency := time.Since(start)

		// Retrieve request ID from context
		 // requestID, _ := c.Get("RequestID")
		 // requestIDStr := "unknown" // Removed as it was declared but not used
		 // if id, ok := requestID.(string); ok {
		 // 	 requestIDStr = id
		 // }

		// Log request details (Consider using a structured logger)
		// log.Printf(
		// 	"[%s] %s %s %d %s %s",
		// 	requestIDStr, // This would need to be fixed if logging is uncommented
		// 	latency,
		// 	 c.Request.Method,
		// 	 c.Writer.Status(),
		// 	 c.Request.URL.Path,
		// 	 c.ClientIP(),
		// )
		
		// Set response time header
		 c.Writer.Header().Set("X-Response-Time", latency.String())

		// Log errors if status code is >= 400
		 if c.Writer.Status() >= http.StatusBadRequest {
		 	 // Log errors associated with the context
		 	 // Consider logging the full error details
		 	 // log.Printf("[%s] Error: %s", requestIDStr, c.Errors.String()) // This would need requestIDStr
		 	 // Optionally set an error header (be cautious about exposing internal errors)
		 	 // c.Writer.Header().Set("X-Error", c.Errors.ByType(gin.ErrorTypePrivate).String())
		 }
		 
		 // Note: The original code set X-Request-ID again here, which is redundant
		 // as it was already set in RequestIDMiddleware. Removed the redundant line.
		 // c.Writer.Header().Set("X-Request-ID", requestID.(string))
	}
}


