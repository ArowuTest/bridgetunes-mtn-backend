package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware is a middleware for JWT authentication
func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the Authorization header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check if the token is expired
			if exp, ok := claims["exp"].(float64); ok {
				if time.Unix(int64(exp), 0).Before(time.Now()) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is expired"})
					c.Abort()
					return
				}
			}

			// Set the claims in the context
			c.Set("claims", claims)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}

// CORSMiddleware is a middleware for CORS
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(cfg.Server.AllowedHosts, ","))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware is a middleware for adding a request ID to the context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = time.Now().Format("20060102150405") + "-" + c.ClientIP()
		}
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

		// Log request
		requestID, _ := c.Get("RequestID")
		c.Writer.Header().Set("X-Response-Time", latency.String())

		// Log to console
		if c.Writer.Status() >= 400 {
			c.Writer.Header().Set("X-Error", c.Errors.String())
		}

		// Log request details
		c.Writer.Header().Set("X-Request-ID", requestID.(string))
	}
}
