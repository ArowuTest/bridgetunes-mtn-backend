package handlers

import (
	"net/http"
	"strconv"
	"time"

	// Added missing import for models
	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/services"
	"github.com/bridgetunes/mtn-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
) 

// DrawHandler handles draw-related HTTP requests
type DrawHandler struct {
	// Use the interface type directly, not a pointer
	 drawService services.DrawService 
}

// NewDrawHandler creates a new DrawHandler
// Accept the interface type directly, not a pointer
func NewDrawHandler(drawService services.DrawService) *DrawHandler {
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
	 // Method calls on interfaces work the same way
	 draw, err := h.drawService.GetDrawByID(c, id)
	 if err != nil {
		 // Check if the error indicates not found, might need specific error handling from service
		 c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusOK, draw) 
}

// GetDrawByDate handles GET /draws/date/:date
// NOTE: This method is likely deprecated as it's not in the DrawService interface
// func (h *DrawHandler) GetDrawByDate(c *gin.Context) { ... }

// GetDrawsByDateRange handles GET /draws/date-range
// NOTE: This method is likely deprecated as it's not in the DrawService interface
// func (h *DrawHandler) GetDrawsByDateRange(c *gin.Context) { ... }

// GetDrawsByStatus handles GET /draws/status/:status
// NOTE: This method is likely deprecated as it's not in the DrawService interface
// func (h *DrawHandler) GetDrawsByStatus(c *gin.Context) { ... }

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
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	 }

	// Parse draw date
	 drawDate, err := time.Parse("2006-01-02", request.DrawDate)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw date format (YYYY-MM-DD) "})
		 return
	 }

	// Validate draw type
	 if request.DrawType != "DAILY" && request.DrawType != "SATURDAY" { // Assuming SATURDAY is the other type
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw type (DAILY or SATURDAY) "})
		 return
	 }

	// Determine eligible digits
	 var eligibleDigits []int
	 if request.UseDefault {
		 // This helper might need to be moved or accessed differently if it was part of the old service struct
		 eligibleDigits = utils.GetDefaultEligibleDigits(drawDate.Weekday())
	 } else {
		 eligibleDigits = request.EligibleDigits
	 }

	// Schedule draw
	 draw, err := h.drawService.ScheduleDraw(c, drawDate, request.DrawType, eligibleDigits, request.UseDefault) // Added useDefaultDigits argument
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule draw: " + err.Error() })
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
	 _, err = h.drawService.ExecuteDraw(c, id) // ExecuteDraw now returns (*models.Draw, error)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute draw: " + err.Error() })
		 return
	 }

	 c.JSON(http.StatusOK, gin.H{"message": "Draw executed successfully"}) 
}

// GetDrawCount handles GET /draws/count
// NOTE: This method is likely deprecated as it's not in the DrawService interface
// func (h *DrawHandler) GetDrawCount(c *gin.Context) { ... }

// GetDefaultEligibleDigits handles GET /draws/default-digits/:day
// NOTE: This method is likely deprecated as it's not in the DrawService interface
// func (h *DrawHandler) GetDefaultEligibleDigits(c *gin.Context) { ... }

// --- Added methods to match DrawHandlerEnhanced and DrawService interface ---

// GetDrawConfig handles GET /draws/config
func (h *DrawHandler) GetDrawConfig(c *gin.Context) {
	 dateStr := c.Query("date")
	 var date time.Time
	 var err error
	 if dateStr == "" {
		 date = time.Now()
	 } else {
		 date, err = time.Parse("2006-01-02", dateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD) "})
			 return
		 }
	 }

	 config, err := h.drawService.GetDrawConfig(c, date)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw config: " + err.Error() })
		 return
	 }
	 c.JSON(http.StatusOK, config) 
}

// GetPrizeStructure handles GET /draws/prize-structure
func (h *DrawHandler) GetPrizeStructure(c *gin.Context) {
	 drawType := c.Query("draw_type")
	 if drawType == "" {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Missing draw_type query parameter"}) 
		 return
	 }
	 structure, err := h.drawService.GetPrizeStructure(c, drawType)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prize structure: " + err.Error() })
		 return
	 }
	 c.JSON(http.StatusOK, structure) 
}

// UpdatePrizeStructure handles PUT /draws/prize-structure
func (h *DrawHandler) UpdatePrizeStructure(c *gin.Context) {
	 var request struct {
		 DrawType  string                 `json:"draw_type" binding:"required"`
		 Structure []models.PrizeStructure `json:"structure" binding:"required"`
	 }
	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error() })
		 return
	 }

	 err := h.drawService.UpdatePrizeStructure(c, request.DrawType, request.Structure)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure: " + err.Error() })
		 return
	 }
	 c.JSON(http.StatusOK, gin.H{"message": "Prize structure updated successfully"}) 
}

// GetDraws handles GET /draws
func (h *DrawHandler) GetDraws(c *gin.Context) {
	 startDateStr := c.Query("start_date")
	 endDateStr := c.Query("end_date")

	 var startDate, endDate time.Time
	 var err error
	 if startDateStr != "" {
		 startDate, err = time.Parse("2006-01-02", startDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD) "})
			 return
		 }
	 }
	 if endDateStr != "" {
		 endDate, err = time.Parse("2006-01-02", endDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD) "})
			 return
		 }
		 endDate = endDate.Add(24 * time.Hour)
	 }

	 draws, err := h.drawService.GetDraws(c, startDate, endDate)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draws: " + err.Error() })
		 return
	 }
	 c.JSON(http.StatusOK, draws) 
}

// GetDrawWinners handles GET /draws/:id/winners
func (h *DrawHandler) GetDrawWinners(c *gin.Context) {
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"}) 
		 return
	 }

	 winners, err := h.drawService.GetWinnersByDrawID(c, id)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw winners: " + err.Error() })
		 return
	 }
	 c.JSON(http.StatusOK, winners) 
}

// GetJackpotHistory handles GET /draws/jackpot-history
func (h *DrawHandler) GetJackpotHistory(c *gin.Context) {
	 startDateStr := c.Query("start_date")
	 endDateStr := c.Query("end_date")

	 var startDate, endDate time.Time
	 var err error
	 if startDateStr != "" {
		 startDate, err = time.Parse("2006-01-02", startDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD) "})
			 return
		 }
	 }
	 if endDateStr != "" {
		 endDate, err = time.Parse("2006-01-02", endDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD) "})
			 return
		 }
		 endDate = endDate.Add(24 * time.Hour)
	 }

	 history, err := h.drawService.GetJackpotHistory(c, startDate, endDate)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get jackpot history: " + err.Error() })
		 return
	 }
	 c.JSON(http.StatusOK, history) 
}

