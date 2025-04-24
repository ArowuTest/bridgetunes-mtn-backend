package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/services"
	"github.com/bridgetunes/mtn-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawHandler handles draw-related HTTP requests
type DrawHandler struct {
	drawService *services.DrawService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(drawService *services.DrawService) *DrawHandler {
	return &DrawHandler{
		drawService: drawService,
	}
}

// GetDrawByID handles GET /draws/:id
func (h *DrawHandler) GetDrawByID(c *gin.Context) {
	// Parse ID from URL
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Get draw from service
	draw, err := h.drawService.GetDrawByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
		return
	}

	c.JSON(http.StatusOK, draw)
}

// GetDrawByDate handles GET /draws/date/:date
func (h *DrawHandler) GetDrawByDate(c *gin.Context) {
	// Parse date from URL
	dateStr := c.Param("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD)"})
		return
	}

	// Get draw from service
	draw, err := h.drawService.GetDrawByDate(c, date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
		return
	}

	c.JSON(http.StatusOK, draw)
}

// GetDrawsByDateRange handles GET /draws/date-range
func (h *DrawHandler) GetDrawsByDateRange(c *gin.Context) {
	// Parse date range parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD)"})
		return
	}

	// Add one day to end date to include the end date in the range
	endDate = endDate.Add(24 * time.Hour)

	// Get draws from service
	draws, err := h.drawService.GetDrawsByDateRange(c, startDate, endDate, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draws"})
		return
	}

	c.JSON(http.StatusOK, draws)
}

// GetDrawsByStatus handles GET /draws/status/:status
func (h *DrawHandler) GetDrawsByStatus(c *gin.Context) {
	// Get status from URL
	status := c.Param("status")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get draws from service
	draws, err := h.drawService.GetDrawsByStatus(c, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draws"})
		return
	}

	c.JSON(http.StatusOK, draws)
}

// ScheduleDraw handles POST /draws/schedule
func (h *DrawHandler) ScheduleDraw(c *gin.Context) {
	// Parse request body
	var request struct {
		DrawDate       string `json:"draw_date" binding:"required"`
		DrawType       string `json:"draw_type" binding:"required"`
		EligibleDigits []int  `json:"eligible_digits"`
		UseDefault     bool   `json:"use_default"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse draw date
	drawDate, err := time.Parse("2006-01-02", request.DrawDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw date format (YYYY-MM-DD)"})
		return
	}

	// Validate draw type
	if request.DrawType != "DAILY" && request.DrawType != "WEEKLY" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw type (DAILY or WEEKLY)"})
		return
	}

	// Determine eligible digits
	var eligibleDigits []int
	if request.UseDefault {
		eligibleDigits = utils.GetDefaultEligibleDigits(drawDate.Weekday())
	} else {
		eligibleDigits = request.EligibleDigits
	}

	// Schedule draw
	draw, err := h.drawService.ScheduleDraw(c, drawDate, request.DrawType, eligibleDigits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule draw"})
		return
	}

	c.JSON(http.StatusCreated, draw)
}

// ExecuteDraw handles POST /draws/:id/execute
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	// Parse ID from URL
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Execute draw
	err = h.drawService.ExecuteDraw(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute draw"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Draw executed successfully"})
}

// GetDrawCount handles GET /draws/count
func (h *DrawHandler) GetDrawCount(c *gin.Context) {
	// Get draw count
	count, err := h.drawService.GetDrawCount(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// GetDefaultEligibleDigits handles GET /draws/default-digits/:day
func (h *DrawHandler) GetDefaultEligibleDigits(c *gin.Context) {
	// Parse day from URL
	dayStr := c.Param("day")
	
	// Convert day string to weekday
	var day time.Weekday
	switch dayStr {
	case "monday":
		day = time.Monday
	case "tuesday":
		day = time.Tuesday
	case "wednesday":
		day = time.Wednesday
	case "thursday":
		day = time.Thursday
	case "friday":
		day = time.Friday
	case "saturday":
		day = time.Saturday
	case "sunday":
		day = time.Sunday
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day"})
		return
	}

	// Get default eligible digits
	digits := h.drawService.GetDefaultEligibleDigits(day)

	c.JSON(http.StatusOK, gin.H{"day": dayStr, "eligible_digits": digits})
}
