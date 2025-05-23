package handlers

import (
	"net/http"
	// "strconv" // Removed unused import
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawHandlerEnhanced handles enhanced draw-related HTTP requests
type DrawHandlerEnhanced struct {
	 drawService services.DrawService // Use the interface type
}

// NewDrawHandlerEnhanced creates a new DrawHandlerEnhanced
func NewDrawHandlerEnhanced(drawService services.DrawService) *DrawHandlerEnhanced { // Use the interface type
	return &DrawHandlerEnhanced{
		 drawService: drawService,
	}
}

// GetDrawConfig handles GET /draws/config
func (h *DrawHandlerEnhanced) GetDrawConfig(c *gin.Context) {
	// Date parsing removed as it's no longer used by the service method

	// Get draw config from service
	 config, err := h.drawService.GetDrawConfig(c.Request.Context()) // Use c.Request.Context(), remove date argument
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw configuration: " + err.Error()})
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
	 structure, err := h.drawService.GetPrizeStructure(c.Request.Context(), drawType) // Use c.Request.Context()
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prize structure: " + err.Error()})
		 return
	 }

	 c.JSON(http.StatusOK, structure)
}

// UpdatePrizeStructure handles PUT /draws/prize-structure
func (h *DrawHandlerEnhanced) UpdatePrizeStructure(c *gin.Context) {
	// Parse request body
	 var request struct {
		 DrawType string `json:"draw_type" binding:"required"`
		 Prizes []models.Prize `json:"prizes" binding:"required"` // Changed to models.Prize
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
	 err := h.drawService.UpdatePrizeStructure(c.Request.Context(), request.DrawType, request.Prizes) // Use c.Request.Context()
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure: " + err.Error()})
		 return
	 }

	 c.JSON(http.StatusOK, gin.H{"message": "Prize structure updated successfully"})
}

// ScheduleDraw handles POST /draws
func (h *DrawHandlerEnhanced) ScheduleDraw(c *gin.Context) {
	// Parse request body
	 var request struct {
		 DrawDate       string   `json:"draw_date" binding:"required"`
		 DrawType       string   `json:"draw_type" binding:"required"`
		 EligibleDigits []int    `json:"eligible_digits"` // Allow empty if use_default is true
		 UseDefault     bool     `json:"use_default"`
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

	// Validate eligible digits if not using default
	 if !request.UseDefault && len(request.EligibleDigits) == 0 {
	 	 c.JSON(http.StatusBadRequest, gin.H{"error": "eligible_digits cannot be empty if use_default is false"})
	 	 return
	 }

	// Schedule draw - Pass arguments in correct order
	 draw, err := h.drawService.ScheduleDraw(c.Request.Context(), drawDate, request.DrawType, request.EligibleDigits, request.UseDefault) // Use c.Request.Context()
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule draw: " + err.Error()})
		 return
	 }

	 c.JSON(http.StatusCreated, draw)
}

// ExecuteDraw handles POST /draws/:id/execute
func (h *DrawHandlerEnhanced) ExecuteDraw(c *gin.Context) {
	// Parse ID from URL
	 drawIDStr := c.Param("id")
	 drawID, err := primitive.ObjectIDFromHex(drawIDStr)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }

	// Execute draw - Capture both return values
	 _, err = h.drawService.ExecuteDraw(c.Request.Context(), drawID) // Use c.Request.Context(), Capture draw object if needed later, otherwise use _
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute draw: " + err.Error()})
		 return
	 }

	 c.JSON(http.StatusOK, gin.H{"message": "Draw executed successfully"})
}

// GetDrawByID handles GET /draws/:id
func (h *DrawHandlerEnhanced) GetDrawByID(c *gin.Context) {
	// Parse ID from URL
	 drawIDStr := c.Param("id") // Renamed to avoid conflict if uncommenting below
	 drawID, err := primitive.ObjectIDFromHex(drawIDStr)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }

	// Get draw from service
	 draw, err := h.drawService.GetDrawByID(c.Request.Context(), drawID) // Use c.Request.Context(), Assuming DrawService has GetDrawByID
	 if err != nil {
	 	 c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found: " + err.Error()})
	 	 return
	 }
	 c.JSON(http.StatusOK, draw)
}

// GetWinnersByDrawID handles GET /draws/:id/winners
func (h *DrawHandlerEnhanced) GetWinnersByDrawID(c *gin.Context) {
	// Parse ID from URL
	 drawIDStr := c.Param("id") // Renamed to avoid conflict if uncommenting below
	 drawID, err := primitive.ObjectIDFromHex(drawIDStr)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }

	// Get winners from service
	 winners, err := h.drawService.GetWinnersByDrawID(c.Request.Context(), drawID) // Use c.Request.Context(), Assuming DrawService has GetWinnersByDrawID
	 if err != nil {
	 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get winners: " + err.Error()})
	 	 return
	 }
	 c.JSON(http.StatusOK, winners)
}

// GetDraws handles GET /draws
func (h *DrawHandlerEnhanced) GetDraws(c *gin.Context) {
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

	// Get draws from service
	 draws, err := h.drawService.GetDraws(c.Request.Context(), startDate, endDate) // Use c.Request.Context(), Assuming DrawService has GetDraws
	 if err != nil {
	 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draws: " + err.Error()})
	 	 return
	 }
	 c.JSON(http.StatusOK, draws)
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
	 history, err := h.drawService.GetJackpotHistory(c.Request.Context(), startDate, endDate) // Use c.Request.Context()
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get jackpot history: " + err.Error()})
		 return
	 }

	 c.JSON(http.StatusOK, history)
}


