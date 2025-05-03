package middleware

import (
	"errors" // Added missing import
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware creates a gin middleware for JWT authentication.
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
	 	 	 	 c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
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

