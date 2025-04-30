package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawServiceEnhanced handles draw logic with enhanced features
type DrawServiceEnhanced struct {
	 drawRepo   repositories.DrawRepository
	 userRepo   repositories.UserRepository
	 winnerRepo repositories.WinnerRepository
	 configRepo repositories.SystemConfigRepository
	 topupRepo  repositories.TopupRepository // Assuming TopupRepository exists
}

// NewDrawServiceEnhanced creates a new DrawServiceEnhanced
func NewDrawServiceEnhanced(
	 drawRepo repositories.DrawRepository,
	 userRepo repositories.UserRepository,
	 winnerRepo repositories.WinnerRepository,
	 configRepo repositories.SystemConfigRepository,
	 topupRepo repositories.TopupRepository,
) *DrawServiceEnhanced {
	 return &DrawServiceEnhanced{
		  drawRepo:   drawRepo,
		  userRepo:   userRepo,
		  winnerRepo: winnerRepo,
		  configRepo: configRepo,
		  topupRepo:  topupRepo,
	 }
}

// Constants for draw types and status
const (
	 DrawTypeDaily   = "DAILY"
	 DrawTypeWeekly  = "WEEKLY"
	 DrawStatusScheduled = "SCHEDULED"
	 DrawStatusInProgress = "IN_PROGRESS"
	 DrawStatusCompleted = "COMPLETED"
	 DrawStatusCancelled = "CANCELLED"

	 JackpotCategory     = "JACKPOT"
	 SecondCategory      = "SECOND"
	 ThirdCategory       = "THIRD"
	 ConsolationCategory = "CONSOLATION"

	 PrizeStructureDailyKey  = "prizeStructureDaily"
	 PrizeStructureWeeklyKey = "prizeStructureWeekly"

	 DefaultDailyJackpot  = 1000000.0
	 DefaultWeeklyJackpot = 3000000.0
)

// GetDrawConfig retrieves draw configuration for a specific date
func (s *DrawServiceEnhanced) GetDrawConfig(ctx context.Context, date time.Time) (map[string]interface{}, error) {
	 drawType := getDrawType(date)
	 prizeKey := PrizeStructureDailyKey
	 if drawType == DrawTypeWeekly {
		  prizeKey = PrizeStructureWeeklyKey
	 }

	 prizeConfig, err := s.configRepo.FindByKey(ctx, prizeKey)
	 if err != nil {
		  log.Printf("Error fetching prize structure config: %v", err)
		  // Fallback to default if config not found
		  prizeConfig = &models.SystemConfig{Value: getDefaultPrizeStructure(drawType)}
	 }

	 // Calculate current jackpot amount
	 jackpotAmount, err := s.calculateCurrentJackpot(ctx, date)
	 if err != nil {
		  return nil, fmt.Errorf("failed to calculate jackpot: %w", err)
	 }

	 // Recommended digits based on day of week
	 recommendedDigits := getRecommendedDigits(date)

	 return map[string]interface{}{
		  "prizeStructure":   prizeConfig.Value,
		  "currentJackpot":   jackpotAmount,
		  "recommendedDigits": recommendedDigits,
		  "drawType":         drawType,
	 }, nil
}

// GetPrizeStructure retrieves the prize structure for a given draw type
func (s *DrawServiceEnhanced) GetPrizeStructure(ctx context.Context, drawType string) ([]models.PrizeStructure, error) {
	 prizeKey := PrizeStructureDailyKey
	 if drawType == DrawTypeWeekly {
		  prizeKey = PrizeStructureWeeklyKey
	 }

	 config, err := s.configRepo.FindByKey(ctx, prizeKey)
	 if err != nil {
		  log.Printf("Error fetching prize structure config: %v", err)
		  // Return default structure if not found
		  return getDefaultPrizeStructure(drawType), nil
	 }

	 // Type assertion to convert interface{} to []models.PrizeStructure
	 prizeStructure, ok := config.Value.([]models.PrizeStructure)
	 if !ok {
		  log.Printf("Error: Invalid prize structure format in config for key %s", prizeKey)
		  return getDefaultPrizeStructure(drawType), fmt.Errorf("invalid prize structure format in config")
	 }

	 return prizeStructure, nil
}

// UpdatePrizeStructure updates the prize structure in the system config
func (s *DrawServiceEnhanced) UpdatePrizeStructure(ctx context.Context, drawType string, structure []models.PrizeStructure) error {
	 prizeKey := PrizeStructureDailyKey
	 if drawType == DrawTypeWeekly {
		  prizeKey = PrizeStructureWeeklyKey
	 }

	 config, err := s.configRepo.FindByKey(ctx, prizeKey)
	 if err != nil {
		  // If not found, create a new config entry
		  newConfig := &models.SystemConfig{
			   Key:         prizeKey,
			   Value:       structure,
			   Description: fmt.Sprintf("Prize structure for %s draws", drawType),
			   UpdatedAt:   time.Now(),
		  }
		  return s.configRepo.Create(ctx, newConfig)
	 }

	 // If found, update the existing config entry
	 config.Value = structure
	 config.UpdatedAt = time.Now()
	 return s.configRepo.Update(ctx, config)
}

// ScheduleDraw schedules a new draw
func (s *DrawServiceEnhanced) ScheduleDraw(ctx context.Context, drawDate time.Time, eligibleDigits []int, useDefaultDigits bool) (*models.Draw, error) {
	 drawType := getDrawType(drawDate)

	 if useDefaultDigits {
		  eligibleDigits = getRecommendedDigits(drawDate)
	 }

	 // Calculate jackpot amount for this draw
	 jackpotAmount, err := s.calculateCurrentJackpot(ctx, drawDate)
	 if err != nil {
		  return nil, fmt.Errorf("failed to calculate jackpot: %w", err)
	 }

	 // Get prize structure for this draw type
	 prizeStructure, err := s.GetPrizeStructure(ctx, drawType)
	 if err != nil {
		  // Log error but proceed with default structure
		  log.Printf("Error getting prize structure, using default: %v", err)
		  prizeStructure = getDefaultPrizeStructure(drawType)
	 }

	 // Create prize entries for the draw based on the structure
	 prizes := []models.Prize{}
	 for _, p := range prizeStructure {
		  for i := 0; i < p.Count; i++ {
			   prizes = append(prizes, models.Prize{
				    Category: p.Category,
				    Amount:   p.Amount,
			   })
		  }
	 }

	 // Find the jackpot prize entry and set its amount
	 for i := range prizes {
		  if prizes[i].Category == JackpotCategory {
			   prizes[i].Amount = jackpotAmount
			   break
		  }
	 }

	 // Get rollover info contributing to this jackpot
	 rolloverSource, err := s.getRolloverSource(ctx, drawDate)
	 if err != nil {
		  return nil, fmt.Errorf("failed to get rollover source: %w", err)
	 }

	 draw := &models.Draw{
		  DrawDate:       drawDate,
		  DrawType:       drawType,
		  EligibleDigits: eligibleDigits,
		  Status:         DrawStatusScheduled,
		  Prizes:         prizes,
		  JackpotAmount:  jackpotAmount,
		  RolloverSource: rolloverSource,
		  CreatedAt:      time.Now(),
		  UpdatedAt:      time.Now(),
	 }

	 err = s.drawRepo.Create(ctx, draw)
	 if err != nil {
		  return nil, fmt.Errorf("failed to create draw: %w", err)
	 }

	 return draw, nil
}

// ExecuteDraw executes a scheduled draw
func (s *DrawServiceEnhanced) ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error) {
	 draw, err := s.drawRepo.FindByID(ctx, drawID)
	 if err != nil {
		  return nil, fmt.Errorf("draw not found: %w", err)
	 }

	 if draw.Status != DrawStatusScheduled {
		  return nil, fmt.Errorf("draw is not in scheduled state (current state: %s)", draw.Status)
	 }

	 // Update status to IN_PROGRESS
	 draw.Status = DrawStatusInProgress
	 draw.UpdatedAt = time.Now()
	 if err := s.drawRepo.Update(ctx, draw); err != nil {
		  return nil, fmt.Errorf("failed to update draw status to in_progress: %w", err)
	 }

	 // Determine eligibility time range
	 startTime, endTime := getEligibilityTimeRange(draw.DrawDate)

	 // --- Get Participants ---
	 // Jackpot Pool: All users who recharged in the time range
	 jackpotParticipants, err := s.userRepo.FindByRechargeTimeRange(ctx, startTime, endTime)
	 if err != nil {
		  s.cancelDraw(ctx, draw, fmt.Sprintf("failed to get jackpot participants: %v", err))
		  return nil, fmt.Errorf("failed to get jackpot participants: %w", err)
	 }

	 // Consolation Pool: Opted-in users who recharged in the time range
	 consolationParticipants := []*models.User{}
	 for _, user := range jackpotParticipants {
		  if user.OptInStatus {
			   consolationParticipants = append(consolationParticipants, user)
		  }
	 }

	 draw.TotalParticipants = len(jackpotParticipants)
	 draw.OptedInParticipants = len(consolationParticipants)

	 // --- Select Winners ---
	 selectedWinners := make(map[string]bool) // Track MSISDNs already selected
	 rand.Seed(time.Now().UnixNano())

	 // Sort prizes: Jackpot first, then others
	 sort.SliceStable(draw.Prizes, func(i, j int) bool {
		  if draw.Prizes[i].Category == JackpotCategory {
			   return true
		  }
		  if draw.Prizes[j].Category == JackpotCategory {
			   return false
		  }
		  return draw.Prizes[i].Amount > draw.Prizes[j].Amount // Higher amounts first
	 })

	 jackpotRolloverNeeded := false
	 for i := range draw.Prizes {
		  prize := &draw.Prizes[i]
		  var winner *models.User
		  var err error

		  // Select from appropriate pool
		  if prize.Category == JackpotCategory {
			   winner, err = s.selectWinnerFromPool(jackpotParticipants, selectedWinners)
		  } else {
			   winner, err = s.selectWinnerFromPool(consolationParticipants, selectedWinners)
		  }

		  if err != nil { // Not enough eligible participants for this prize
			   log.Printf("Could not select winner for %s prize: %v", prize.Category, err)
			   if prize.Category == JackpotCategory {
				    jackpotRolloverNeeded = true
				    isValid := false
				    prize.IsValid = &isValid // Mark jackpot prize as invalid if no winner found
			   }
			   continue // Skip to next prize
		  }

		  // Mark winner as selected
		  selectedWinners[winner.MSISDN] = true

		  // Create Winner record
		  isValid := true
		  if prize.Category == JackpotCategory && !winner.OptInStatus {
			   isValid = false // Jackpot winner must be opted-in
			   jackpotRolloverNeeded = true
		  }

		  // Removed unused variable: winnerMSISDN := winner.MSISDN
		  maskedMSISDN := maskMSISDN(winner.MSISDN)

		  winnerRecord := &models.Winner{
			   MSISDN:       winner.MSISDN,
			   MaskedMSISDN: maskedMSISDN,
			   DrawID:       draw.ID,
			   PrizeCategory: prize.Category,
			   PrizeAmount:  prize.Amount,
			   IsOptedIn:    winner.OptInStatus,
			   IsValid:      isValid,
			   Points:       winner.Points, // Assuming User model has Points
			   WinDate:      draw.DrawDate,
			   ClaimStatus:  "PENDING",
			   CreatedAt:    time.Now(),
			   UpdatedAt:    time.Now(),
		  }

		  err = s.winnerRepo.Create(ctx, winnerRecord)
		  if err != nil {
			   s.cancelDraw(ctx, draw, fmt.Sprintf("failed to save winner record: %v", err))
			   return nil, fmt.Errorf("failed to save winner record: %w", err)
		  }

		  prize.WinnerID = winnerRecord.ID
		  prize.IsValid = &isValid
	 }

	 // Handle jackpot rollover if needed
	 if jackpotRolloverNeeded {
		  nextDrawDate := getNextDrawDate(draw.DrawDate)
		  nextDraw, err := s.drawRepo.FindByDate(ctx, nextDrawDate)
		  if err != nil {
			   // If next draw doesn't exist, log warning (rollover handled when scheduling)
			   log.Printf("Warning: Next draw on %v not found for rollover from draw %s", nextDrawDate, draw.ID.Hex())
		  } else {
			   // Mark current draw as rolling over to the next draw
			   draw.RolloverTarget = &nextDraw.ID
		  }
	 }

	 // Update draw status to COMPLETED
	 draw.Status = DrawStatusCompleted
	 draw.UpdatedAt = time.Now()
	 if err := s.drawRepo.Update(ctx, draw); err != nil {
		  // Log error but don't cancel the draw at this point
		  log.Printf("Error updating draw status to completed for draw %s: %v", draw.ID.Hex(), err)
	 }

	 // TODO: trigger notifications to winners

	 return draw, nil
}

// GetDraws retrieves a list of draws within a time range
func (s *DrawServiceEnhanced) GetDraws(ctx context.Context, startTime, endTime time.Time) ([]*models.Draw, error) {
	 // Implementation updated to match interface: uses startTime, endTime
	 // Assuming FindByDateRange takes (ctx, startTime, endTime) and handles pagination internally or doesn't need it.
	 // If FindByDateRange requires pagination, its signature or the interface might need adjustment.
	 log.Printf("Fetching draws between %v and %v", startTime, endTime)
	 draws, err := s.drawRepo.FindByDateRange(ctx, startTime, endTime)
	 if err != nil {
		  log.Printf("Error fetching draws by date range: %v", err)
		  return nil, fmt.Errorf("failed to fetch draws by date range: %w", err)
	 }
	 return draws, nil
}

// GetDrawByID retrieves a single draw by its ID
func (s *DrawServiceEnhanced) GetDrawByID(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error) {
	 return s.drawRepo.FindByID(ctx, drawID)
}

// GetDrawWinners retrieves the winners for a specific draw
func (s *DrawServiceEnhanced) GetDrawWinners(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	 return s.winnerRepo.FindByDrawID(ctx, drawID)
}

// GetJackpotHistory retrieves the history of jackpot amounts
func (s *DrawServiceEnhanced) GetJackpotHistory(ctx context.Context, limit int) ([]models.JackpotHistoryEntry, error) {
	 // 1. Get completed draws with Jackpot prizes, sorted by date descending
	 draws, err := s.drawRepo.FindCompletedWithJackpot(ctx, limit)
	 if err != nil {
		  return nil, fmt.Errorf("failed to get completed jackpot draws: %w", err)
	 }

	 history := []models.JackpotHistoryEntry{}
	 for _, draw := range draws {
		  var jackpotPrize *models.Prize
		  for i := range draw.Prizes {
			   if draw.Prizes[i].Category == JackpotCategory {
				    jackpotPrize = &draw.Prizes[i]
				    break
			   }
		  }

		  if jackpotPrize == nil {
			   log.Printf("Warning: Draw %s completed but no jackpot prize found?", draw.ID.Hex())
			   continue
		  }

		  entry := models.JackpotHistoryEntry{
			   DrawDate:     draw.DrawDate,
			   JackpotAmount: jackpotPrize.Amount,
			   Won:          jackpotPrize.WinnerID != primitive.NilObjectID && jackpotPrize.IsValid != nil && *jackpotPrize.IsValid,
		  }

		  if entry.Won {
			   winner, err := s.winnerRepo.FindByID(ctx, jackpotPrize.WinnerID)
			   if err != nil {
				    log.Printf("Error fetching winner %s for draw %s: %v", jackpotPrize.WinnerID.Hex(), draw.ID.Hex(), err)
				    // Still include history entry, just without winner info
			   } else {
				    entry.WinnerMSISDN = winner.MaskedMSISDN
			   }
		  }
		  history = append(history, entry)
	 }

	 return history, nil
}

// --- Helper Functions ---

// getDrawType determines if a date falls on a weekly draw day (e.g., Sunday)
func getDrawType(date time.Time) string {
	 if date.Weekday() == time.Sunday {
		  return DrawTypeWeekly
	 }
	 return DrawTypeDaily
}

// getRecommendedDigits returns the recommended eligible digits based on the day of the week
func getRecommendedDigits(date time.Time) []int {
	 switch date.Weekday() {
	 case time.Monday:
		  return []int{0, 1}
	 case time.Tuesday:
		  return []int{2, 3}
	 case time.Wednesday:
		  return []int{4, 5}
	 case time.Thursday:
		  return []int{6, 7}
	 case time.Friday:
		  return []int{8, 9}
	 case time.Saturday:
		  return []int{0, 1, 2, 3, 4} // Example for Saturday
	 case time.Sunday:
		  return []int{5, 6, 7, 8, 9} // Example for Sunday (Weekly Draw)
	 default:
		  return []int{} // Should not happen
	 }
}

// getDefaultPrizeStructure returns a default prize structure based on draw type
func getDefaultPrizeStructure(drawType string) []models.PrizeStructure {
	 if drawType == DrawTypeWeekly {
		  return []models.PrizeStructure{
			   {Category: JackpotCategory, Amount: DefaultWeeklyJackpot, Count: 1},
			   {Category: SecondCategory, Amount: 50000, Count: 5},
			   {Category: ThirdCategory, Amount: 10000, Count: 20},
			   {Category: ConsolationCategory, Amount: 1000, Count: 100},
		  }
	 }
	 // Default Daily
	 return []models.PrizeStructure{
		  {Category: JackpotCategory, Amount: DefaultDailyJackpot, Count: 1},
		  {Category: SecondCategory, Amount: 10000, Count: 10},
		  {Category: ThirdCategory, Amount: 5000, Count: 50},
		  {Category: ConsolationCategory, Amount: 500, Count: 200},
	 }
}

// calculateCurrentJackpot calculates the jackpot amount for a given draw date,
// considering rollovers from previous draws.
func (s *DrawServiceEnhanced) calculateCurrentJackpot(ctx context.Context, drawDate time.Time) (float64, error) {
	 drawType := getDrawType(drawDate)
	 defaultJackpot := DefaultDailyJackpot
	 if drawType == DrawTypeWeekly {
		  defaultJackpot = DefaultWeeklyJackpot
	 }

	 // Find the most recent completed draw *before* the current drawDate
	 previousDraw, err := s.drawRepo.FindMostRecentCompletedBefore(ctx, drawDate)
	 if err != nil {
		  if errors.Is(err, mongo.ErrNoDocuments) {
			   // No previous completed draw, use default jackpot
			   return defaultJackpot, nil
		  } else {
			   return 0, fmt.Errorf("failed to find previous completed draw: %w", err)
		  }
	 }

	 // Check if the previous draw's jackpot was won and valid
	 previousJackpotWon := false
	 previousJackpotAmount := 0.0
	 for _, prize := range previousDraw.Prizes {
		  if prize.Category == JackpotCategory {
			   previousJackpotAmount = prize.Amount
			   if prize.WinnerID != primitive.NilObjectID && prize.IsValid != nil && *prize.IsValid {
				    previousJackpotWon = true
			   }
			   break
		  }
	 }

	 if previousJackpotWon {
		  // Previous jackpot was won, start with default for the current draw type
		  return defaultJackpot, nil
	 } else {
		  // Previous jackpot was not won (or invalid), rollover the amount
		  // Rollover logic might need refinement based on specific business rules
		  // (e.g., does daily rollover to daily, weekly to weekly?)
		  // Simple rollover: Add previous jackpot amount to current default
		  return defaultJackpot + previousJackpotAmount, nil
	 }
}

// getRolloverSource finds the ID of the draw whose potential rollover contributes to the current draw
func (s *DrawServiceEnhanced) getRolloverSource(ctx context.Context, drawDate time.Time) (*primitive.ObjectID, error) {
	 previousDraw, err := s.drawRepo.FindMostRecentCompletedBefore(ctx, drawDate)
	 if err != nil {
		  if errors.Is(err, mongo.ErrNoDocuments) {
			   return nil, nil // No previous draw, no rollover source
		  } else {
			   return nil, fmt.Errorf("failed to find previous completed draw for rollover source: %w", err)
		  }
	 }
	 return &previousDraw.ID, nil
}

// getEligibilityTimeRange calculates the start and end time for participant eligibility
// based on the draw date and type.
func getEligibilityTimeRange(drawDate time.Time) (time.Time, time.Time) {
	 year, month, day := drawDate.Date()
	 drawDayStart := time.Date(year, month, day, 0, 0, 0, 0, drawDate.Location())

	 if getDrawType(drawDate) == DrawTypeWeekly {
		  // Weekly draw (Sunday): Eligible from start of previous Sunday to end of Saturday before draw
		  startOfWeek := drawDayStart.AddDate(0, 0, -7)
		  endOfEligibility := drawDayStart.Add(-1 * time.Second) // End of Saturday
		  return startOfWeek, endOfEligibility
	 }

	 // Daily draw: Eligible from start of previous day to end of day before draw
	 startOfPreviousDay := drawDayStart.AddDate(0, 0, -1)
	 endOfEligibility := drawDayStart.Add(-1 * time.Second) // End of previous day
	 return startOfPreviousDay, endOfEligibility
}

// selectWinnerFromPool selects a random winner from the participant pool,
// ensuring they haven't been selected already.
func (s *DrawServiceEnhanced) selectWinnerFromPool(participants []*models.User, selectedWinners map[string]bool) (*models.User, error) {
	 eligiblePool := []*models.User{}
	 for _, p := range participants {
		  if !selectedWinners[p.MSISDN] {
			   eligiblePool = append(eligiblePool, p)
		  }
	 }

	 if len(eligiblePool) == 0 {
		  return nil, errors.New("no eligible participants remaining in the pool")
	 }

	 winnerIndex := rand.Intn(len(eligiblePool))
	 return eligiblePool[winnerIndex], nil
}

// cancelDraw updates the draw status to CANCELLED and logs the reason
func (s *DrawServiceEnhanced) cancelDraw(ctx context.Context, draw *models.Draw, reason string) {
	 log.Printf("Cancelling draw %s: %s", draw.ID.Hex(), reason)
	 draw.Status = DrawStatusCancelled
	 draw.UpdatedAt = time.Now()
	 // Add reason to draw notes or a dedicated field if available
	 // draw.Notes = reason
	 if err := s.drawRepo.Update(ctx, draw); err != nil {
		  log.Printf("Error updating draw %s status to cancelled: %v", draw.ID.Hex(), err)
	 }
}

// getNextDrawDate calculates the date of the next draw
func getNextDrawDate(currentDrawDate time.Time) time.Time {
	 return currentDrawDate.AddDate(0, 0, 1) // Simple assumption: next draw is always the next day
}

// maskMSISDN masks a phone number (e.g., 23480xxxx1234)
func maskMSISDN(msisdn string) string {
	 if len(msisdn) < 7 { // Need at least 7 digits to mask reasonably
		  return msisdn // Return original if too short
	 }
	 // Example: Keep first 5 and last 4 digits (adjust as needed)
	 prefixLength := 5
	 suffixLength := 4
	 maskedPart := ""
	 for i := 0; i < len(msisdn)-(prefixLength+suffixLength); i++ {
		  maskedPart += "x"
	 }
	 return msisdn[:prefixLength] + maskedPart + msisdn[len(msisdn)-suffixLength:]
}

