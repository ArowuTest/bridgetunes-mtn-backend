package middleware

import (
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
		// Ensure header retrieval is case-insensitive if needed, though GetHeader is typically sufficient
		// Consider trimming whitespace from the header value
		 authHeader := c.GetHeader("Authorization")
		 if authHeader == "" {
			// Provide a clear error message
			// Use standard HTTP status codes
			 c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			 c.Abort()
			 return
		}

		// Check if the Authorization header has the correct format
		// Ensure splitting handles potential extra spaces correctly (strings.Fields might be an alternative)
		// Current split logic is standard and generally okay.
		 parts := strings.Split(authHeader, " ")
		 if len(parts) != 2 || parts[0] != "Bearer" {
			// Provide a clear error message about the expected format
			 c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			 c.Abort()
			 return
		}

		// Extract the token
		// Consider trimming whitespace from the token string itself
		 tokenString := parts[1]

		// Parse and validate the token
		// Ensure the secret key retrieval from config is robust
		// Handle potential errors during parsing gracefully
		 token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method to prevent algorithm downgrade attacks
			 if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				// Log the unexpected signing method for security monitoring
				// log.Printf("Unexpected signing method: %v", token.Header["alg"])
				 return nil, errors.New("unexpected signing method")
			}
			// Ensure cfg.JWT.Secret is not empty
			 if cfg.JWT.Secret == "" {
			 	 // Log this critical configuration error
			 	 // log.Println("Error: JWT Secret is not configured")
			 	 return nil, errors.New("JWT secret not configured")
			 }
			 return []byte(cfg.JWT.Secret), nil
		})

		// Handle parsing errors (e.g., malformed token, signature mismatch)
		 if err != nil {
		 	 // Log the specific error for debugging
		 	 // log.Printf("Token parsing error: %v", err)
			 c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			 c.Abort()
			 return
		}

		// Check if the token is valid and extract claims
		// Ensure claims extraction is safe (type assertion)
		 if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check if the token is expired
			// Validate the type of the 'exp' claim (should be float64 for standard JWT)
			 if exp, ok := claims["exp"].(float64); ok {
				// Compare expiration time with current time
				 if time.Unix(int64(exp), 0).Before(time.Now()) {
					 c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is expired"})
					 c.Abort()
					 return
				}
			} else {
				// Handle case where 'exp' claim is missing or not a number
				 c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (expiration)"})
				 c.Abort()
				 return
			}

			// Set the claims in the context for downstream handlers
			// Consider setting specific claims like user ID, role directly if needed
			// e.g., c.Set("userID", claims["user_id"])
			// e.g., c.Set("userRole", claims["role"])
			 c.Set("claims", claims)
			 c.Next() // Proceed to the next handler
		} else {
			// Handle invalid token (e.g., signature invalid, claims invalid)
			 c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			 c.Abort()
			 return
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
		 requestID, _ := c.Get("RequestID")
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

