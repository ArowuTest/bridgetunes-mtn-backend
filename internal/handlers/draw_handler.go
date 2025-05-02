package handlers

import (
	"net/http"
	// "strconv" // Removed unused import
	"time"

	// Ensure this exact import path is used
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/utils"
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
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"}) 
		 return
	 }
	 draw, err := h.drawService.GetDrawByID(c, id)
	 if err != nil {
		 c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found: " + err.Error()  })
		 return
	 }
	 c.JSON(http.StatusOK, draw) 
}

// ScheduleDraw handles POST /draws/schedule
func (h *DrawHandler) ScheduleDraw(c *gin.Context) {
	 var request struct {
		 DrawDate       string `json:"draw_date" binding:"required"`
		 DrawType       string `json:"draw_type" binding:"required"`
		 EligibleDigits []int  `json:"eligible_digits"`
		 UseDefault     bool   `json:"use_default"`
	 }
	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()  })
		 return
	 }
	 drawDate, err := time.Parse("2006-01-02", request.DrawDate)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw date format (YYYY-MM-DD)  "})
		 return
	 }
	 if request.DrawType != "DAILY" && request.DrawType != "SATURDAY" {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw type (DAILY or SATURDAY)  "})
		 return
	 }
	 var eligibleDigits []int
	 if request.UseDefault {
		 eligibleDigits = utils.GetDefaultEligibleDigits(drawDate.Weekday())
	 } else {
		 eligibleDigits = request.EligibleDigits
	 }
	 draw, err := h.drawService.ScheduleDraw(c, drawDate, request.DrawType, eligibleDigits, request.UseDefault)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule draw: " + err.Error()  })
		 return
	 }
	 c.JSON(http.StatusCreated, draw) 
}

// ExecuteDraw handles POST /draws/:id/execute
func (h *DrawHandler) ExecuteDraw(c *gin.Context) {
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"}) 
		 return
	 }
	 _, err = h.drawService.ExecuteDraw(c, id)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute draw: " + err.Error()  })
		 return
	 }
	 c.JSON(http.StatusOK, gin.H{"message": "Draw executed successfully"}) 
}

// --- Methods matching DrawHandlerEnhanced and DrawService interface ---

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
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD)  "})
			 return
		 }
	 }
	 config, err := h.drawService.GetDrawConfig(c, date)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw config: " + err.Error()  })
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
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prize structure: " + err.Error()  })
		 return
	 }
	 c.JSON(http.StatusOK, structure) 
}

// Define the request structure outside the function
type UpdatePrizeStructureRequest struct {
	DrawType  string                 `json:"draw_type" binding:"required"`
	// Expects PrizeStructure from the models package
	Structure []models.PrizeStructure `json:"structure" binding:"required"` 
}

// UpdatePrizeStructure handles PUT /draws/prize-structure
func (h *DrawHandler) UpdatePrizeStructure(c *gin.Context) {
	 var request UpdatePrizeStructureRequest // Use the named struct type
	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()  })
		 return
	 }
	 // Pass the structure directly to the service (assuming it expects []models.PrizeStructure)
	 err := h.drawService.UpdatePrizeStructure(c, request.DrawType, request.Structure)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure: " + err.Error()  })
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
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD)  "})
			 return
		 }
	 }
	 if endDateStr != "" {
		 endDate, err = time.Parse("2006-01-02", endDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD)  "})
			 return
		 }
		 endDate = endDate.Add(24 * time.Hour)
	 }
	 draws, err := h.drawService.GetDraws(c, startDate, endDate)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draws: " + err.Error()  })
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
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw winners: " + err.Error()  })
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
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format (YYYY-MM-DD)  "})
			 return
		 }
	 }
	 if endDateStr != "" {
		 endDate, err = time.Parse("2006-01-02", endDateStr)
		 if err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format (YYYY-MM-DD)  "})
			 return
		 }
		 endDate = endDate.Add(24 * time.Hour)
	 }
	 history, err := h.drawService.GetJackpotHistory(c, startDate, endDate)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get jackpot history: " + err.Error()  })
		 return
	 }
	 c.JSON(http.StatusOK, history) 
}



// CreateDraw handles POST /draws
// Placeholder implementation
func (h *DrawHandler) CreateDraw(c *gin.Context) {
	// TODO: Implement logic to parse request body, call drawService.CreateDraw, and handle response/errors
	var draw models.Draw
	 if err := c.ShouldBindJSON(&draw); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		 return
	 }
	 // Assuming drawService.CreateDraw exists and takes context and *models.Draw
	 // newDraw, err := h.drawService.CreateDraw(c.Request.Context(), &draw)
	 // if err != nil {
	 // 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create draw: " + err.Error()})
	 // 	 return
	 // }
	 // c.JSON(http.StatusCreated, newDraw)
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "CreateDraw not implemented"})
}

// UpdateDraw handles PUT /draws/:id
// Placeholder implementation
func (h *DrawHandler) UpdateDraw(c *gin.Context) {
	// TODO: Implement logic to parse ID, parse request body, call drawService.UpdateDraw, and handle response/errors
	 _, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }
	 var draw models.Draw
	 if err := c.ShouldBindJSON(&draw); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		 return
	 }
	 // draw.ID = id // Set the ID from the path parameter
	 // Assuming drawService.UpdateDraw exists and takes context and *models.Draw
	 // updatedDraw, err := h.drawService.UpdateDraw(c.Request.Context(), &draw)
	 // if err != nil {
	 // 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update draw: " + err.Error()})
	 // 	 return
	 // }
	 // c.JSON(http.StatusOK, updatedDraw)
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "UpdateDraw not implemented"})
}

// DeleteDraw handles DELETE /draws/:id
// Placeholder implementation
func (h *DrawHandler) DeleteDraw(c *gin.Context) {
	// TODO: Implement logic to parse ID, call drawService.DeleteDraw, and handle response/errors
	 _, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }
	 // Assuming drawService.DeleteDraw exists and takes context and primitive.ObjectID
	 // err = h.drawService.DeleteDraw(c.Request.Context(), id)
	 // if err != nil {
	 // 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete draw: " + err.Error()})
	 // 	 return
	 // }
	 // c.JSON(http.StatusOK, gin.H{"message": "Draw deleted successfully"})
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "DeleteDraw not implemented"})
}

// GetWinners handles GET /draws/winners/:id
// Placeholder implementation - Note: routes.go has /draws/winners/:id but GetDrawWinners uses /draws/:id/winners
// Assuming the route is /draws/:id/winners based on GetDrawWinners implementation
func (h *DrawHandler) GetWinners(c *gin.Context) {
	// TODO: Implement logic to parse ID, call drawService.GetWinnersByDrawID, and handle response/errors
	 _, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }
	 // winners, err := h.drawService.GetWinnersByDrawID(c.Request.Context(), id)
	 // if err != nil {
	 // 	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get winners: " + err.Error()})
	 // 	 return
	 // }
	 // c.JSON(http.StatusOK, winners)
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetWinners not implemented"})
}

// GetDrawByDate handles GET /draws/date/:date
// Placeholder implementation
func (h *DrawHandler) GetDrawByDate(c *gin.Context) {
	// TODO: Implement logic to parse date, call drawService.GetDrawByDate, and handle response/errors
	 dateStr := c.Param("date")
	 _, err := time.Parse("2006-01-02", dateStr)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD)"})
		 return
	 }
	 // draw, err := h.drawService.GetDrawByDate(c.Request.Context(), date)
	 // if err != nil {
	 // 	 c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found for date: " + err.Error()})
	 // 	 return
	 // }
	 // c.JSON(http.StatusOK, draw)
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetDrawByDate not implemented"})
}

// GetDefaultDigitsForDay handles GET /draws/default-digits/:day
// Placeholder implementation
func (h *DrawHandler) GetDefaultDigitsForDay(c *gin.Context) {
	// TODO: Implement logic to parse day, call utils.GetDefaultEligibleDigits, and return digits
	 dayStr := c.Param("day") // e.g., "Monday", "Tuesday"
	 // Convert dayStr to time.Weekday if needed by the service/util function
	 // weekday, err := utils.ParseWeekday(dayStr) // Assuming a helper function exists
	 // if err != nil {
	 // 	 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day name"})
	 // 	 return
	 // }
	 // digits := utils.GetDefaultEligibleDigits(weekday)
	 // c.JSON(http.StatusOK, gin.H{"day": dayStr, "digits": digits})
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetDefaultDigitsForDay not implemented", "requested_day": dayStr})
}

