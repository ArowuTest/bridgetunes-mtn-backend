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

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	// Use the interface type directly, not a pointer
	notificationService services.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
// Accept the interface type directly, not a pointer
func NewNotificationHandler(notificationService services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotificationByID handles GET /notifications/:id
func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Get notification from service
	// Method calls on interfaces work the same way
	 notification, err := h.notificationService.GetNotificationByID(c, id)
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, notification)
}

// GetNotificationsByMSISDN handles GET /notifications/msisdn/:msisdn
func (h *NotificationHandler) GetNotificationsByMSISDN(c *gin.Context) {
	// Get MSISDN from URL
	 msisdn := c.Param("msisdn")

	// Parse pagination parameters
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get notifications from service
	 notifications, err := h.notificationService.GetNotificationsByMSISDN(c, msisdn, page, limit)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, notifications)
}

// GetNotificationsByCampaignID handles GET /notifications/campaign/:id
func (h *NotificationHandler) GetNotificationsByCampaignID(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Parse pagination parameters
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get notifications from service
	 notifications, err := h.notificationService.GetNotificationsByCampaignID(c, id, page, limit)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, notifications)
}

// GetNotificationsByStatus handles GET /notifications/status/:status
func (h *NotificationHandler) GetNotificationsByStatus(c *gin.Context) {
	// Get status from URL
	 status := c.Param("status")

	// Parse pagination parameters
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get notifications from service
	 notifications, err := h.notificationService.GetNotificationsByStatus(c, status, page, limit)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, notifications)
}

// SendSMS handles POST /notifications/send-sms
func (h *NotificationHandler) SendSMS(c *gin.Context) {
	// Parse request body
	 var request struct {
		MSISDN         string `json:"msisdn" binding:"required"`
		Content        string `json:"content" binding:"required"`
		NotificationType string `json:"notification_type" binding:"required"`
		CampaignID     string `json:"campaign_id"`
	}

	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Parse campaign ID
	 var campaignID primitive.ObjectID
	 var err error
	 if request.CampaignID != "" {
		 campaignID, err = primitive.ObjectIDFromHex(request.CampaignID)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID format"})
			 return
		}
	}

	// Send SMS
	 notification, err := h.notificationService.SendSMS(c, request.MSISDN, request.Content, request.NotificationType, campaignID)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send SMS: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, notification)
}

// CreateCampaign handles POST /notifications/campaigns
func (h *NotificationHandler) CreateCampaign(c *gin.Context) {
	// Parse request body
	 var campaign models.Campaign
	 if err := c.ShouldBindJSON(&campaign); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Create campaign
	 err := h.notificationService.CreateCampaign(c, &campaign)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusCreated, campaign)
}

// ExecuteCampaign handles POST /notifications/campaigns/:id/execute
func (h *NotificationHandler) ExecuteCampaign(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Execute campaign
	 err = h.notificationService.ExecuteCampaign(c, id)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute campaign: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, gin.H{"message": "Campaign executed successfully"})
}

// GetAllCampaigns handles GET /notifications/campaigns
// Placeholder implementation
func (h *NotificationHandler) GetAllCampaigns(c *gin.Context) {
	// TODO: Implement logic to fetch all campaigns, potentially with pagination
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	 // Example placeholder response
	 campaigns := []gin.H{
		{"id": "campaign1", "name": "Welcome Campaign", "status": "active"},
		{"id": "campaign2", "name": "Promo Campaign", "status": "inactive"},
	}
	 c.JSON(http.StatusOK, gin.H{
	 	"message": "GetAllCampaigns endpoint reached (placeholder)",
	 	"page": page,
	 	"limit": limit,
	 	"campaigns": campaigns,
	 })
}

// CreateTemplate handles POST /notifications/templates
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	// Parse request body
	 var template models.Template
	 if err := c.ShouldBindJSON(&template); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Create template
	 err := h.notificationService.CreateTemplate(c, &template)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusCreated, template)
}

// GetTemplateByID handles GET /notifications/templates/:id
func (h *NotificationHandler) GetTemplateByID(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Get template from service
	 template, err := h.notificationService.GetTemplateByID(c, id)
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "Template not found: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, template)
}

// GetTemplateByName handles GET /notifications/templates/name/:name
func (h *NotificationHandler) GetTemplateByName(c *gin.Context) {
	// Get name from URL
	 name := c.Param("name")

	// Get template from service
	 template, err := h.notificationService.GetTemplateByName(c, name)
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "Template not found: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, template)
}

// GetTemplatesByType handles GET /notifications/templates/type/:type
func (h *NotificationHandler) GetTemplatesByType(c *gin.Context) {
	// Get type from URL
	 templateType := c.Param("type")

	// Parse pagination parameters
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get templates from service
	 templates, err := h.notificationService.GetTemplatesByType(c, templateType, page, limit)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get templates: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, templates)
}

// UpdateTemplate handles PUT /notifications/templates/:id
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Parse request body
	 var template models.Template
	 if err := c.ShouldBindJSON(&template); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	}

	// Set the ID on the template object from the URL parameter
	 template.ID = id

	// Update template - Removed the 'id' argument from the call as it's part of the template object
	 err = h.notificationService.UpdateTemplate(c, &template)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, template)
}

// DeleteTemplate handles DELETE /notifications/templates/:id
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	}

	// Delete template
	 err = h.notificationService.DeleteTemplate(c, id)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template: " + err.Error() })
		 return
	}

	 c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

// GetAllTemplates handles GET /notifications/templates
// Placeholder implementation
func (h *NotificationHandler) GetAllTemplates(c *gin.Context) {
	// TODO: Implement logic to fetch all templates, potentially with pagination
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	 // Example placeholder response
	 templates := []gin.H{
		{"id": "template1", "name": "Welcome SMS", "type": "SMS"},
		{"id": "template2", "name": "Winner Notification", "type": "SMS"},
	}
	 c.JSON(http.StatusOK, gin.H{
	 	"message": "GetAllTemplates endpoint reached (placeholder)",
	 	"page": page,
	 	"limit": limit,
	 	"templates": templates,
	 })
}


// GetNotifications handles GET /notifications
// Placeholder implementation - Assumes pagination and potentially filtering
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// TODO: Implement logic to parse query params (pagination, filters like status, msisdn, campaign_id), call notificationService.GetNotifications, and handle response/errors
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	 // Add other filters as needed

	 // Assuming notificationService.GetNotifications exists and takes context, page, limit, and potentially filters
	 // notifications, err := h.notificationService.GetNotifications(c.Request.Context(), page, limit /*, filters... */)
	 // if err != nil {
	 // 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications: " + err.Error()})
	 // 	 return
	 // }
	 // c.JSON(http.StatusOK, notifications)
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetNotifications not implemented", "page": page, "limit": limit})
}



