package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
) 

// TopupHandler handles topup-related HTTP requests
type TopupHandler struct {
	// Use the interface type directly, not a pointer
	 topupService services.TopupService 
}

// NewTopupHandler creates a new TopupHandler
// Accept the interface type directly, not a pointer
func NewTopupHandler(topupService services.TopupService) *TopupHandler {
	return &TopupHandler{
		 topupService: topupService,
	}
}

// GetTopupByID handles GET /topups/:id
func (h *TopupHandler) GetTopupByID(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"}) 
		 return
	 }

	// Get topup from service
	 // Method calls on interfaces work the same way
	 topup, err := h.topupService.GetTopupByID(c, id)
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "Topup not found: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusOK, topup) 
}

// GetTopupsByMSISDN handles GET /topups/msisdn/:msisdn
func (h *TopupHandler) GetTopupsByMSISDN(c *gin.Context) {
	// Get MSISDN from URL
	 msisdn := c.Param("msisdn")

	// Parse pagination parameters
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get topups from service
	 topups, err := h.topupService.GetTopupsByMSISDN(c, msisdn, page, limit)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get topups: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusOK, topups) 
}

// GetTopupsByDateRange handles GET /topups/date-range
func (h *TopupHandler) GetTopupsByDateRange(c *gin.Context) {
	// Parse date range parameters
	 startDateStr := c.Query("start_date")
	 endDateStr := c.Query("end_date")

	// Parse pagination parameters
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse dates
	 startDate, err := time.Parse("2006-01-02", startDateStr)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD) "})
		 return
	 }

	 endDate, err := time.Parse("2006-01-02", endDateStr)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD) "})
		 return
	 }

	// Add one day to end date to include the end date in the range
	 endDate = endDate.Add(24 * time.Hour)

	// Get topups from service
	 topups, err := h.topupService.GetTopupsByDateRange(c, startDate, endDate, page, limit)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get topups: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusOK, topups) 
}

// CreateTopup handles POST /topups
func (h *TopupHandler) CreateTopup(c *gin.Context) {
	// Parse request body
	 var topup models.Topup
	 if err := c.ShouldBindJSON(&topup); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	 }

	// Create topup
	 err := h.topupService.CreateTopup(c, &topup)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create topup: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusCreated, topup) 
}

// ProcessTopups handles POST /topups/process
func (h *TopupHandler) ProcessTopups(c *gin.Context) {
	// Parse request body
	 var request struct {
		 StartDate string `json:"start_date" binding:"required"`
		 EndDate   string `json:"end_date" binding:"required"`
	 }
	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	 }

	// Parse dates
	 startDate, err := time.Parse("2006-01-02", request.StartDate)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD) "})
		 return
	 }

	 endDate, err := time.Parse("2006-01-02", request.EndDate)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD) "})
		 return
	 }

	// Add one day to end date to include the end date in the range
	 endDate = endDate.Add(24 * time.Hour)

	// Process topups
	 processed, err := h.topupService.ProcessTopups(c, startDate, endDate)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process topups: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusOK, gin.H{"message": "Topups processed successfully", "processed": processed}) 
}

// GetTopupCount handles GET /topups/count
func (h *TopupHandler) GetTopupCount(c *gin.Context) {
	// Get topup count
	 count, err := h.topupService.GetTopupCount(c)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get topup count: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusOK, gin.H{"count": count}) 
}


// GetTopups handles GET /topups
// Placeholder implementation - Assumes pagination and potentially filtering
func (h *TopupHandler) GetTopups(c *gin.Context) {
	// TODO: Implement logic to parse query params (pagination, filters), call topupService.GetTopups, and handle response/errors
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	 // Add other filters as needed (e.g., msisdn, date range)

	 // Assuming topupService.GetTopups exists and takes context, page, limit, and potentially filters
	 // topups, err := h.topupService.GetTopups(c.Request.Context(), page, limit /*, filters... */)
	 // if err != nil {
	 // 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get topups: " + err.Error()})
	 // 	 return
	 // }
	 // c.JSON(http.StatusOK, topups)
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetTopups not implemented", "page": page, "limit": limit})
}


