package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/services"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// DrawHandler handles draw-related HTTP requests
type DrawHandler struct {
	 drawService services.DrawService
}

// NewDrawHandler creates a new DrawHandler
func NewDrawHandler(drawService services.DrawService) *DrawHandler {
	return &DrawHandler{
		 drawService: drawService,
	}
}

// Helper function to parse weekday string (case-insensitive)
func parseWeekday(dayStr string) (time.Weekday, bool) {
	 dayStrLower := strings.ToLower(dayStr)
	 switch dayStrLower {
	 case "sunday":
		 return time.Sunday, true
	 case "monday":
		 return time.Monday, true
	 case "tuesday":
		 return time.Tuesday, true
	 case "wednesday":
		 return time.Wednesday, true
	 case "thursday":
		 return time.Thursday, true
	 case "friday":
		 return time.Friday, true
	 case "saturday":
		 return time.Saturday, true
	 default:
		 return 0, false
	 }
}

// GetDrawByID handles GET /draws/:id
func (h *DrawHandler) GetDrawByID(c *gin.Context) {
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }
	 // Assuming drawService.GetDrawByID exists and takes context and ID
	 draw, err := h.drawService.GetDrawByID(c.Request.Context(), id)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) {
			 c.JSON(http.StatusNotFound, gin.H{"error": "Draw not found"})
		 } else {
			 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draw: " + err.Error()})
		 }
		 return
	 }
	 c.JSON(http.StatusOK, draw)
}

// ScheduleDraw handles POST /draws/schedule
type ScheduleDrawRequest struct {
	DrawDate       string `json:"draw_date" binding:"required"`
	DrawType       string `json:"draw_type" binding:"required"` // DAILY or SATURDAY
	EligibleDigits []int  `json:"eligible_digits"`
	UseDefault     bool   `json:"use_default"`
}

func (h *DrawHandler) ScheduleDraw(c *gin.Context) {
	 var request ScheduleDrawRequest
	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		 return
	 }
	 drawDate, err := time.Parse("2006-01-02", request.DrawDate)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw date format (YYYY-MM-DD)"})
		 return
	 }
	 if request.DrawType != "DAILY" && request.DrawType != "SATURDAY" {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid draw type (DAILY or SATURDAY)"})
		 return
	 }

	 // Service layer now handles default digit logic based on UseDefault flag
	 draw, err := h.drawService.ScheduleDraw(c.Request.Context(), drawDate, request.DrawType, request.EligibleDigits, request.UseDefault)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule draw: " + err.Error()})
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
	 // ExecuteDraw now returns the updated draw object or error
	 executedDraw, err := h.drawService.ExecuteDraw(c.Request.Context(), id)
	 if err != nil {
		 // Return the draw object even on error, as it contains logs
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute draw: " + err.Error(), "draw_details": executedDraw})
		 return
	 }
	 c.JSON(http.StatusOK, gin.H{"message": "Draw executed successfully", "draw_details": executedDraw})
}

// GetDrawByDate handles GET /draws/date/:date
func (h *DrawHandler) GetDrawByDate(c *gin.Context) {
	 dateStr := c.Param("date")
	 date, err := time.Parse("2006-01-02", dateStr)
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format (YYYY-MM-DD)"})
		 return
	 }

	 draw, err := h.drawService.GetDrawByDate(c.Request.Context(), date)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) || strings.Contains(err.Error(), "no draw found") {
			 c.JSON(http.StatusNotFound, gin.H{"error": "No draw found for the specified date: " + dateStr})
		 } else {
			 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve draw data: " + err.Error()})
		 }
		 return
	 }
	 c.JSON(http.StatusOK, draw)
}

// GetDefaultDigitsForDay handles GET /draws/default-digits/:day
func (h *DrawHandler) GetDefaultDigitsForDay(c *gin.Context) {
	 dayStr := c.Param("day")
	 weekday, ok := parseWeekday(dayStr)
	 if !ok {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day name provided. Use full names like 'Monday', 'Tuesday', etc."})
		 return
	 }
	 digits := utils.GetDefaultEligibleDigits(weekday)
	 c.JSON(http.StatusOK, gin.H{"day": dayStr, "digits": digits})
}

// GetDrawWinners handles GET /draws/:id/winners
func (h *DrawHandler) GetDrawWinners(c *gin.Context) {
	 id, err := primitive.ObjectIDFromHex(c.Param("id"))
	 if err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		 return
	 }
	 winners, err := h.drawService.GetWinnersByDrawID(c.Request.Context(), id)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get draw winners: " + err.Error()})
		 return
	 }
	 c.JSON(http.StatusOK, winners)
}

// GetJackpotStatus handles GET /draws/jackpot-status
func (h *DrawHandler) GetJackpotStatus(c *gin.Context) {
	 status, err := h.drawService.GetJackpotStatus(c.Request.Context())
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get jackpot status: " + err.Error()})
		 return
	 }
	 c.JSON(http.StatusOK, status)
}

// GetPrizeStructure handles GET /draws/prize-structure
func (h *DrawHandler) GetPrizeStructure(c *gin.Context) {
	 drawType := c.Query("draw_type")
	 if drawType == "" {
		 c.JSON(http.StatusBadRequest, gin.H{"error": "Missing draw_type query parameter"})
		 return
	 }
	 structure, err := h.drawService.GetPrizeStructure(c.Request.Context(), drawType)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prize structure: " + err.Error()})
		 return
	 }
	 c.JSON(http.StatusOK, structure)
}

// UpdatePrizeStructure handles PUT /draws/prize-structure
type UpdatePrizeStructureRequest struct {
	DrawType  string          `json:"draw_type" binding:"required"`
	Structure []models.Prize `json:"structure" binding:"required"` // Expecting []models.Prize
}

func (h *DrawHandler) UpdatePrizeStructure(c *gin.Context) {
	 var request UpdatePrizeStructureRequest
	 if err := c.ShouldBindJSON(&request); err != nil {
		 c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		 return
	 }
	 err := h.drawService.UpdatePrizeStructure(c.Request.Context(), request.DrawType, request.Structure)
	 if err != nil {
		 c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prize structure: " + err.Error()})
		 return
	 }
	 c.JSON(http.StatusOK, gin.H{"message": "Prize structure updated successfully"})
}

// --- Placeholder/Unused Handlers (Review if needed) ---

// GetDrawConfig (Service method is placeholder, keep handler as placeholder)
func (h *DrawHandler) GetDrawConfig(c *gin.Context) {
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetDrawConfig endpoint not fully implemented"})
}

// GetDraws (Service method is placeholder, keep handler as placeholder)
func (h *DrawHandler) GetDraws(c *gin.Context) {
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetDraws endpoint not fully implemented"})
}

// GetJackpotHistory (Service method is placeholder, keep handler as placeholder)
func (h *DrawHandler) GetJackpotHistory(c *gin.Context) {
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetJackpotHistory endpoint not fully implemented"})
}

// CreateDraw (Service method is placeholder, keep handler as placeholder)
func (h *DrawHandler) CreateDraw(c *gin.Context) {
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "CreateDraw endpoint not fully implemented"})
}

// UpdateDraw (Service method is placeholder, keep handler as placeholder)
func (h *DrawHandler) UpdateDraw(c *gin.Context) {
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "UpdateDraw endpoint not fully implemented"})
}

// DeleteDraw (Service method is placeholder, keep handler as placeholder)
func (h *DrawHandler) DeleteDraw(c *gin.Context) {
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "DeleteDraw endpoint not fully implemented"})
}

// GetWinners (Redundant with GetDrawWinners, keep as placeholder)
func (h *DrawHandler) GetWinners(c *gin.Context) {
	 c.JSON(http.StatusNotImplemented, gin.H{"message": "GetWinners endpoint not implemented (use /draws/:id/winners)"})
}


