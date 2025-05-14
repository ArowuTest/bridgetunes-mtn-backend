package handlers

import (
	"net/http"
	"strconv"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/dgrijalva/jwt-go" // Added missing import for JWT
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	// Use the interface type directly, not a pointer
	userService services.UserService
	// Consider adding AdminUserService if GetMe should return AdminUser
	// adminUserService services.AdminUserService
}

// NewUserHandler creates a new UserHandler
// Accept the interface type directly, not a pointer
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserByID handles GET /users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Get user from service
	// Method calls on interfaces work the same way
	 user, err := h.userService.GetUserByID(c, id)
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "User not found: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, user)
}

// GetUserByMSISDN handles GET /users/msisdn/:msisdn
func (h *UserHandler) GetUserByMSISDN(c *gin.Context) {
	// Get MSISDN from URL
	 msisdn := c.Param("msisdn")

	// Get user from service
	 user, err := h.userService.GetUserByMSISDN(c, msisdn)
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "User not found: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, user)
}

// GetAllUsers handles GET /users
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Parse pagination parameters
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get users from service
	 users, err := h.userService.GetAllUsers(c, page, limit)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, users)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Parse request body
	 var user models.User
	 if err := c.ShouldBindJSON(&user); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Create user
	 err := h.userService.CreateUser(c.Request.Context(), &user)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Parse request body
	 var user models.User
	 if err := c.ShouldBindJSON(&user); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Set ID
	 user.ID = id

	// Update user
	 err = h.userService.UpdateUser(c, &user)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Delete user
	 err = h.userService.DeleteUser(c, id)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// OptIn handles POST /users/opt-in
func (h *UserHandler) OptIn(c *gin.Context) {
	// Parse request body
	 var request struct {
		MSISDN  string `json:"msisdn" binding:"required"`
		Channel string `json:"channel" binding:"required"`
	}

	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Opt in user
	 err := h.userService.OptIn(c, request.MSISDN, request.Channel)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to opt in user: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, gin.H{"message": "User opted in successfully"})
}

// OptOut handles POST /users/opt-out
func (h *UserHandler) OptOut(c *gin.Context) {
	// Parse request body
	 var request struct {
		MSISDN string `json:"msisdn" binding:"required"`
	}

	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Opt out user
	 err := h.userService.OptOut(c, request.MSISDN)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to opt out user: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, gin.H{"message": "User opted out successfully"})
}

// GetUserCount handles GET /users/count
func (h *UserHandler) GetUserCount(c *gin.Context) {
	// Get user count
	 count, err := h.userService.GetUserCount(c)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user count: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, gin.H{"count": count})
}



// GetMe handles GET /users/me
// It retrieves the user ID from the JWT claims and fetches the user details.
// IMPORTANT: This currently assumes the JWT belongs to an AdminUser and fetches details using AdminUserRepository.
// If the JWT is for regular Users, this needs significant changes.
func (h *UserHandler) GetMe(c *gin.Context) {
	// Extract user ID from JWT claims (assuming it's stored as "sub")
	userIDClaim, exists := c.Get("claims") // Assuming claims are stored under "claims" by JWT middleware
	 if !exists {
		 c.JSON(http.StatusUnauthorized, gin.H{"error": "User claims not found in context"})
		 return
	}

	claims, ok := userIDClaim.(jwt.MapClaims)
	 if !ok {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid claims format"})
		 return
	}

	userIDStr, ok := claims["sub"].(string) // Assuming user ID is stored in the 'sub' claim
	 if !ok {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID (sub) not found in token claims"})
		 return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in token"})
		 return
	}

	// TODO: Clarify if /users/me should return AdminUser or User.
	// If AdminUser, we need AdminUserService injected here.
	// If User, the JWT needs to contain the User's ObjectID.
	// For now, assuming it's an AdminUser and calling a non-existent AdminUserService method.
	// This will need correction based on actual requirements.

	// Placeholder: Attempting to fetch AdminUser details (requires AdminUserService)
	/*
	adminUser, err := h.adminUserService.GetAdminUserByID(c, userID) // Assuming adminUserService exists and has this method
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found: " + err.Error() })
		 return
	}
	adminUser.Password = "" // Ensure password hash is not returned
	 c.JSON(http.StatusOK, adminUser)
	*/

	// Temporary fallback: Return a placeholder message until the logic is clarified
	// Use the parsed userID (ObjectID) in the response to fix the unused variable error
	 c.JSON(http.StatusOK, gin.H{
	 	"message": "GetMe endpoint needs clarification: Should return AdminUser or User?",
	 	"userID_from_token_str": userIDStr,
	 	"userID_from_token_obj": userID, // Use the userID variable here
	 })

	// Original code attempting to fetch a regular User - might be incorrect if JWT is for admins
	// user, err := h.userService.GetUserByID(c, userID)
	// if err != nil {
	// 	 c.JSON(http.StatusNotFound, gin.H{"error": "User not found: " + err.Error() })
	// 	 return
	// }
	// c.JSON(http.StatusOK, user)
}

// GetDashboardStats handles GET /dashboard/stats
// Placeholder implementation
func (h *UserHandler) GetDashboardStats(c *gin.Context) {
	// TODO: Implement actual logic to fetch dashboard statistics
	// This might involve calling methods on userService or other services/repositories.
	 c.JSON(http.StatusOK, gin.H{
	 	"message": "Dashboard stats endpoint reached (placeholder)",
	 	"total_users": 1234, // Example data
	 	"total_draws": 56,   // Example data
	 	"total_topups": 7890, // Example data
	 })
}



