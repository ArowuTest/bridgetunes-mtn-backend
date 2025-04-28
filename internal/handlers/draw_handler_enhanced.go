package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawHandlerEnhanced handles enhanced draw-related HTTP requests
type DrawHandlerEnhanced struct {
	 drawService *services.DrawServiceEnhanced
}

// NewDrawHandlerEnhanced creates a new DrawHandlerEnhanced
func NewDrawHandlerEnhanced(drawService *services.DrawServiceEnhanced) *DrawHandlerEnhanced {
	return &DrawHandlerEnhanced{
		 drawService: drawService,
	}
}

// GetDrawConfig handles GET /draws/config
func (h *DrawHandlerEnhanced) GetDrawConfig(c *gin.Context) {
	// Parse date from query parameter
	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	 if err != nil {
		 // Default to today if date is not provided or invalid
		 date = time.Now()
	 }

	// Get draw config from service
	config, err := h.drawService.GetDrawConfig(c, date)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw configuration"})
		 return
	 }

	 c.JSON(http.StatusOK, config)
}

// GetPrizeStructure handles GET /draws/prize-structure
func (h *DrawHandlerEnhanced) GetPrizeStructure(c *gin.Context) {
	// Get draw type from query parameter
	 drawType := c.Query("draw_type")
	 if drawType != "DAILY" && drawType != "WEEKLY" {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw type (DAILY or WEEKLY)"})
		 return
	 }

	// Get prize structure from service
	 structure, err := h.drawService.GetPrizeStructure(c, drawType)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prize structure"})
		 return
	 }

	 c.JSON(http.StatusOK, structure)
}

// UpdatePrizeStructure handles PUT /draws/prize-structure
func (h *DrawHandlerEnhanced) UpdatePrizeStructure(c *gin.Context) {
	// Parse request body
	 var request struct {
		 DrawType string `json:"draw_type" binding:"required"`
		 Prizes []models.PrizeStructure `json:"prizes" binding:"required"`
	 }
	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		 return
	 }

	// Validate draw type
	 if request.DrawType != "DAILY" && request.DrawType != "WEEKLY" {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw type (DAILY or WEEKLY)"})
		 return
	 }

	// Update prize structure
	 err := h.drawService.UpdatePrizeStructure(c, request.DrawType, request.Prizes)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure"})
		 return
	 }

	 c.JSON(http.StatusOK, gin.H{"message": "Prize structure updated successfully"})
}

// ScheduleDraw handles POST /draws
func (h *DrawHandlerEnhanced) ScheduleDraw(c *gin.Context) {
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
		 eligibleDigits = h.drawService.GetDefaultEligibleDigits(drawDate.Weekday())
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
func (h *DrawHandlerEnhanced) ExecuteDraw(c *gin.Context) {
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

// GetDrawByID handles GET /draws/:id
func (h *DrawHandlerEnhanced) GetDrawByID(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }

	// Get draw from service
	// Assuming DrawServiceEnhanced has a method GetDrawByID
	// draw, err := h.drawService.GetDrawByID(c, id)
	// if err != nil {
	// 	 c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
	// 	 return
	// }
	// c.JSON(http.StatusOK, draw)
	 c.JSON(http.StatusNotImplemented, gin.H{"error": "GetDrawByID not fully implemented in enhanced handler yet"})
}

// GetDrawWinners handles GET /draws/:id/winners
func (h *DrawHandlerEnhanced) GetDrawWinners(c *gin.Context) {
	// Parse ID from URL
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }

	// Get query parameters
	 category := c.Query("category")
	 maskMSISDN, _ := strconv.ParseBool(c.DefaultQuery("mask_msisdn", "true"))

	// Get winners from service
	// Assuming DrawServiceEnhanced has a method GetWinnersByDrawID
	// winners, err := h.drawService.GetWinnersByDrawID(c, id, category, maskMSISDN)
	// if err != nil {
	// 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get winners"})
	// 	 return
	// }
	// c.JSON(http.StatusOK, winners)
	 c.JSON(http.StatusNotImplemented, gin.H{"error": "GetDrawWinners not fully implemented in enhanced handler yet"})
}

// GetDraws handles GET /draws
func (h *DrawHandlerEnhanced) GetDraws(c *gin.Context) {
	// Parse query parameters
	 status := c.Query("status")
	 startDateStr := c.Query("start_date")
	 endDateStr := c.Query("end_date")
	 page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	 limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Parse dates
	 var startDate, endDate time.Time
	 var err error
	 if startDateStr != "" {
		 startDate, err = time.Parse("2006-01-02", startDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD)"})
			 return
		 }
	 }
	 if endDateStr != "" {
		 endDate, err = time.Parse("2006-01-02", endDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD)"})
			 return
		 }
		 // Add one day to end date to include the end date in the range
		 endDate = endDate.Add(24 * time.Hour)
	 }

	// Get draws from service
	// Assuming DrawServiceEnhanced has a method GetDraws
	// draws, err := h.drawService.GetDraws(c, status, startDate, endDate, page, limit)
	// if err != nil {
	// 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draws"})
	// 	 return
	// }
	// c.JSON(http.StatusOK, draws)
	 c.JSON(http.StatusNotImplemented, gin.H{"error": "GetDraws not fully implemented in enhanced handler yet"})
}

// GetJackpotHistory handles GET /draws/jackpot-history
func (h *DrawHandlerEnhanced) GetJackpotHistory(c *gin.Context) {
	// Parse query parameters
	 startDateStr := c.Query("start_date")
	 endDateStr := c.Query("end_date")

	// Parse dates
	 var startDate, endDate time.Time
	 var err error
	 if startDateStr != "" {
		 startDate, err = time.Parse("2006-01-02", startDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD)"})
			 return
		 }
	 }
	 if endDateStr != "" {
		 endDate, err = time.Parse("2006-01-02", endDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD)"})
			 return
		 }
		 // Add one day to end date to include the end date in the range
		 endDate = endDate.Add(24 * time.Hour)
	 }

	// Get jackpot history from service
	 history, err := h.drawService.GetJackpotHistory(c, startDate, endDate)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get jackpot history"})
		 return
	 }

	 c.JSON(http.StatusOK, history)
}

