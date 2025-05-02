package handlers

import (
	"net/http"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication related HTTP requests
type AuthHandler struct {
	// Use the interface type directly
	authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler
// Accept the interface type directly
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		// Correctly assign the passed service to the struct field
		// authService: authService, // Original incorrect line
		// Corrected line:
		 authService: authService,
	}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Bind JSON request body to the RegisterRequest struct
	if err := c.ShouldBindJSON(&req); err != nil {
		// If binding fails, return a bad request error
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		// More detailed error:
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the Register method on the authService
	user, err := h.authService.Register(c.Request.Context(), &req)
	// Check for errors from the service
	 if err != nil {
	 	// Return an internal server error (or specific error based on service logic)
	 	// c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
	 	// More detailed error:
	 	 c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	 	return
	 }

	// Return the newly created user (excluding password) with status 201 Created
	// Note: The service should ideally handle removing the password before returning
	 c.JSON(http.StatusCreated, user)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// Bind JSON request body to the LoginRequest struct
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
	// 	 return
	// }
	// More detailed error:
	 if err := c.ShouldBindJSON(&req); err != nil {
	 	 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	 	return
	 }

	// Call the Login method on the authService
	token, err := h.authService.Login(c.Request.Context(), &req)
	// Check for errors (e.g., invalid credentials)
	 if err != nil {
	 	// Return an unauthorized error
	 	// c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	 	// More detailed error:
	 	 c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	 	return
	 }

	// Return the JWT token with status 200 OK
	 c.JSON(http.StatusOK, gin.H{"token": token})
}

