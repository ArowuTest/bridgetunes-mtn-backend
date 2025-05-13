package handlers

import (
	"net/http"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// SystemSettingsHandler handles system settings-related HTTP requests
type SystemSettingsHandler struct {
	settingsService services.SystemSettingsService
}

// NewSystemSettingsHandler creates a new SystemSettingsHandler
func NewSystemSettingsHandler(settingsService services.SystemSettingsService) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		settingsService: settingsService,
	}
}

// GetSettings handles GET /settings
func (h *SystemSettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsService.GetSettings(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateSettings handles PUT /settings
func (h *SystemSettingsHandler) UpdateSettings(c *gin.Context) {
	var settings models.SystemSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.settingsService.UpdateSettings(c, &settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateSMSGateway handles PUT /settings/sms-gateway
func (h *SystemSettingsHandler) UpdateSMSGateway(c *gin.Context) {
	var request struct {
		Gateway string `json:"gateway" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the user ID from the JWT token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := h.settingsService.UpdateSMSGateway(c, request.Gateway, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SMS gateway: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SMS gateway updated successfully"})
} 