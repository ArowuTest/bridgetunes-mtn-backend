package handlers

import (
	"net/http"
	"strconv"
	// "time" // Removed unused import

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	// Use the interface type directly, not a pointer
	userService services.UserService
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
	 err := h.userService.CreateUser(c, &user)
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

