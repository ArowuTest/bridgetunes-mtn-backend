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
	 DrawTypeDaily    = "DAILY"
	 DrawTypeWeekly   = "WEEKLY"
	 DrawStatusScheduled = "SCHEDULED"
	 DrawStatusInProgress = "IN_PROGRESS"
	 DrawStatusCompleted = "COMPLETED"
	 DrawStatusCancelled = "CANCELLED"
	 JackpotCategory   = "JACKPOT"
	 SecondCategory    = "SECOND"
	 ThirdCategory     = "THIRD"
	 ConsolationCategory = "CONSOLATION"
	 PrizeStructureDailyKey = "prizeStructureDaily"
	 PrizeStructureWeeklyKey = "prizeStructureWeekly"
	 DefaultDailyJackpot = 1000000.0
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
	 	 "prizeStructure":    prizeConfig.Value,
	 	 "currentJackpot":    jackpotAmount,
	 	 "recommendedDigits": recommendedDigits,
	 	 "drawType":          drawType,
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

	 	 if err != nil {
	 	 	 // Not enough eligible participants for this prize
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
	 	 	 MSISDN:        winner.MSISDN,
	 	 	 MaskedMSISDN:  maskedMSISDN,
	 	 	 DrawID:        draw.ID,
	 	 	 PrizeCategory: prize.Category,
	 	 	 PrizeAmount:   prize.Amount,
	 	 	 IsOptedIn:     winner.OptInStatus,
	 	 	 IsValid:       isValid,
	 	 	 Points:        winner.Points, // Assuming User model has Points
	 	 	 WinDate:       draw.DrawDate,
	 	 	 ClaimStatus:   "PENDING",
	 	 	 CreatedAt:     time.Now(),
	 	 	 UpdatedAt:     time.Now(),
	 	 }

	 	 err = s.winnerRepo.Create(ctx, winnerRecord)
	 	 if err != nil {
	 	 	 s.cancelDraw(ctx, draw, fmt.Sprintf("failed to save winner record: %v", err))
	 	 	 return nil, fmt.Errorf("failed to save winner record: %w", err)
	 	 }

	 	 // Link prize to winner and set validity
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

	 // TODO: Trigger notifications to winners

	 return draw, nil
}

// GetDraws retrieves a list of draws with pagination
func (s *DrawServiceEnhanced) GetDraws(ctx context.Context, page, limit int) ([]*models.Draw, error) {
	 // Simple implementation, add filtering/sorting as needed
	 return s.drawRepo.FindByDateRange(ctx, time.Time{}, time.Now().AddDate(10, 0, 0), page, limit)
}

// GetDrawByID retrieves a single draw by its ID
func (s *DrawServiceEnhanced) GetDrawByID(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error) {
	 return s.drawRepo.FindByID(ctx, drawID)
}

// GetDrawWinners retrieves the winners for a specific draw
func (s *DrawServiceEnhanced) GetDrawWinners(ctx context.Context, drawID primitive.ObjectID, page, limit int) ([]*models.Winner, error) {
	 return s.winnerRepo.FindByDrawID(ctx, drawID, page, limit)
}

// GetJackpotHistory retrieves the history of jackpot amounts and rollovers
func (s *DrawServiceEnhanced) GetJackpotHistory(ctx context.Context, page, limit int) ([]map[string]interface{}, error) {
	 // Fetch recent completed draws
	 draws, err := s.drawRepo.FindByStatus(ctx, DrawStatusCompleted, page, limit)
	 if err != nil {
	 	 return nil, fmt.Errorf("failed to fetch completed draws: %w", err)
	 }

	 history := []map[string]interface{}{}
	 for _, draw := range draws {
	 	 jackpotPrize := models.Prize{}
	 	 isValid := false
	 	 for _, p := range draw.Prizes {
	 	 	 if p.Category == JackpotCategory {
	 	 	 	 jackpotPrize = p
	 	 	 	 if p.IsValid != nil {
	 	 	 	 	 isValid = *p.IsValid
	 	 	 	 }
	 	 	 	 break
	 	 	 }
	 	 }
	 	 
	 	 entry := map[string]interface{}{
	 	 	 "drawId":        draw.ID.Hex(),
	 	 	 "drawDate":      draw.DrawDate,
	 	 	 "drawType":      draw.DrawType,
	 	 	 "jackpotAmount": jackpotPrize.Amount,
	 	 	 "isValid":       isValid,
	 	 	 "rolloverSource": draw.RolloverSource,
	 	 	 "rolloverTarget": draw.RolloverTarget,
	 	 }
	 	 history = append(history, entry)
	 }
	 return history, nil
}

// --- Helper Functions --- 

// selectWinnerFromPool selects a unique winner from the pool using points weighting
func (s *DrawServiceEnhanced) selectWinnerFromPool(pool []*models.User, alreadySelected map[string]bool) (*models.User, error) {
	 if len(pool) == 0 {
	 	 return nil, errors.New("participant pool is empty")
	 }

	 // Create weighted list (entries per user based on points)
	 weightedList := []string{}
	 eligibleUsers := []*models.User{}
	 for _, user := range pool {
	 	 if !alreadySelected[user.MSISDN] {
	 	 	 eligibleUsers = append(eligibleUsers, user)
	 	 	 points := user.Points
	 	 	 if points <= 0 {
	 	 	 	 points = 1 // Ensure at least one entry
	 	 	 }
	 	 	 for i := 0; i < points; i++ {
	 	 	 	 weightedList = append(weightedList, user.MSISDN)
	 	 	 }
	 	 }
	 }

	 if len(weightedList) == 0 {
	 	 return nil, errors.New("no eligible participants remaining in the pool")
	 }

	 // Select a random entry from the weighted list
	 randomIndex := rand.Intn(len(weightedList))
	 selectedMSISDN := weightedList[randomIndex]

	 // Find the corresponding user
	 for _, user := range eligibleUsers {
	 	 if user.MSISDN == selectedMSISDN {
	 	 	 return user, nil
	 	 }
	 }

	 return nil, errors.New("internal error: selected MSISDN not found in eligible users") // Should not happen
}

// calculateCurrentJackpot calculates the jackpot amount for a given draw date, considering rollovers
func (s *DrawServiceEnhanced) calculateCurrentJackpot(ctx context.Context, drawDate time.Time) (float64, error) {
	 drawType := getDrawType(drawDate)
	 baseJackpot := DefaultDailyJackpot
	 if drawType == DrawTypeWeekly {
	 	 baseJackpot = DefaultWeeklyJackpot
	 }

	 totalJackpot := baseJackpot

	 // Find previous draw(s) that might have rolled over
	 prevDrawDate := getPreviousDrawDate(drawDate)
	 prevDraw, err := s.drawRepo.FindByDate(ctx, prevDrawDate)
	 if err != nil {
	 	 // No previous draw found, just use base jackpot
	 	 log.Printf("No previous draw found for date %v, using base jackpot %f", prevDrawDate, baseJackpot)
	 	 return baseJackpot, nil
	 }

	 // Check if previous draw rolled over
	 if prevDraw.Status == DrawStatusCompleted {
	 	 jackpotRolledOver := false
	 	 rolloverAmount := 0.0
	 	 for _, p := range prevDraw.Prizes {
	 	 	 if p.Category == JackpotCategory {
	 	 	 	 if p.IsValid != nil && !*p.IsValid {
	 	 	 	 	 jackpotRolledOver = true
	 	 	 	 	 rolloverAmount = p.Amount // Rollover the full amount
	 	 	 	 }
	 	 	 	 break
	 	 	 }
	 	 }

	 	 if jackpotRolledOver {
	 	 	 // If previous was daily and current is weekly, add fixed amount (e.g., 1M)
	 	 	 if prevDraw.DrawType == DrawTypeDaily && drawType == DrawTypeWeekly {
	 	 	 	 totalJackpot += DefaultDailyJackpot // Add the daily jackpot amount
	 	 	 } else {
	 	 	 	 // Otherwise (daily->daily or weekly->weekly), add the full rollover amount
	 	 	 	 // Recursively check for further rollovers if needed (simplified here)
	 	 	 	 prevJackpot, err := s.calculateCurrentJackpot(ctx, prevDrawDate) // Get the actual jackpot of the previous draw
	 	 	 	 if err == nil {
	 	 	 	 	 totalJackpot += prevJackpot
	 	 	 	 } else {
	 	 	 	 	 log.Printf("Error calculating previous jackpot for rollover: %v", err)
	 	 	 	 	 totalJackpot += rolloverAmount // Fallback to prize amount
	 	 	 	 }
	 	 	 }
	 	 }
	 }

	 return totalJackpot, nil
}

// getRolloverSource finds draws that rolled over into the current draw date
func (s *DrawServiceEnhanced) getRolloverSource(ctx context.Context, drawDate time.Time) ([]models.RolloverInfo, error) {
	 sources := []models.RolloverInfo{}
	 prevDrawDate := getPreviousDrawDate(drawDate)
	 prevDraw, err := s.drawRepo.FindByDate(ctx, prevDrawDate)
	 if err != nil {
	 	 return sources, nil // No previous draw, no rollover source
	 }

	 if prevDraw.Status == DrawStatusCompleted {
	 	 jackpotRolledOver := false
	 	 rolloverAmount := 0.0
	 	 for _, p := range prevDraw.Prizes {
	 	 	 if p.Category == JackpotCategory {
	 	 	 	 if p.IsValid != nil && !*p.IsValid {
	 	 	 	 	 jackpotRolledOver = true
	 	 	 	 	 rolloverAmount = p.Amount
	 	 	 	 }
	 	 	 	 break
	 	 	 }
	 	 }

	 	 if jackpotRolledOver {
	 	 	 // Add this draw as a source
	 	 	 sources = append(sources, models.RolloverInfo{
	 	 	 	 SourceDrawID: prevDraw.ID,
	 	 	 	 Amount:       rolloverAmount, // Amount contributed
	 	 	 	 Reason:       "INVALID_WINNER",
	 	 	 })
	 	 	 // Recursively check previous draw's sources (simplified)
	 	 	 prevSources, _ := s.getRolloverSource(ctx, prevDrawDate)
	 	 	 sources = append(sources, prevSources...)
	 	 }
	 }
	 return sources, nil
}

// cancelDraw updates draw status to CANCELLED with a reason
func (s *DrawServiceEnhanced) cancelDraw(ctx context.Context, draw *models.Draw, reason string) {
	 log.Printf("Cancelling draw %s: %s", draw.ID.Hex(), reason)
	 draw.Status = DrawStatusCancelled
	 // Optionally add reason to draw model
	 draw.UpdatedAt = time.Now()
	 if err := s.drawRepo.Update(ctx, draw); err != nil {
	 	 log.Printf("Error updating draw status to cancelled for draw %s: %v", draw.ID.Hex(), err)
	 }
}

// getDrawType determines if a date falls on a Saturday (WEEKLY) or weekday (DAILY)
func getDrawType(date time.Time) string {
	 if date.Weekday() == time.Saturday {
	 	 return DrawTypeWeekly
	 }
	 return DrawTypeDaily
}

// getRecommendedDigits returns the default eligible digits based on the day of the week
func getRecommendedDigits(date time.Time) []int {
	 switch date.Weekday() {
	 case time.Monday: return []int{0, 1}
	 case time.Tuesday: return []int{2, 3}
	 case time.Wednesday: return []int{4, 5}
	 case time.Thursday: return []int{6, 7}
	 case time.Friday: return []int{8, 9}
	 case time.Saturday: return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} // All digits for weekly
	 case time.Sunday: return []int{} // No draws on Sunday
	 default: return []int{}
	 }
}

// getDefaultPrizeStructure returns a default prize structure if config is missing
func getDefaultPrizeStructure(drawType string) []models.PrizeStructure {
	 if drawType == DrawTypeWeekly {
	 	 return []models.PrizeStructure{
	 	 	 {Category: JackpotCategory, Amount: DefaultWeeklyJackpot, Count: 1},
	 	 	 {Category: SecondCategory, Amount: 1000000, Count: 1},
	 	 	 {Category: ThirdCategory, Amount: 500000, Count: 1},
	 	 	 {Category: ConsolationCategory, Amount: 100000, Count: 7},
	 	 }
	 }
	 return []models.PrizeStructure{
	 	 {Category: JackpotCategory, Amount: DefaultDailyJackpot, Count: 1},
	 	 {Category: SecondCategory, Amount: 350000, Count: 1},
	 	 {Category: ThirdCategory, Amount: 150000, Count: 1},
	 	 {Category: ConsolationCategory, Amount: 75000, Count: 7},
	 }
}

// getEligibilityTimeRange calculates the start and end time for participant eligibility
func getEligibilityTimeRange(drawDate time.Time) (time.Time, time.Time) {
	 // Cut-off time is 6 PM on the draw date
	 year, month, day := drawDate.Date()
	 endTime := time.Date(year, month, day, 18, 0, 0, 0, drawDate.Location())
	 
	 // Start time is 6 PM on the previous draw date
	 prevDrawDate := getPreviousDrawDate(drawDate)
	 prevYear, prevMonth, prevDay := prevDrawDate.Date()
	 startTime := time.Date(prevYear, prevMonth, prevDay, 18, 0, 0, 0, prevDrawDate.Location())
	 
	 return startTime, endTime
}

// getPreviousDrawDate finds the date of the immediately preceding draw (skipping Sunday)
func getPreviousDrawDate(currentDate time.Time) time.Time {
	 prevDate := currentDate.AddDate(0, 0, -1)
	 if prevDate.Weekday() == time.Sunday {
	 	 prevDate = prevDate.AddDate(0, 0, -1) // Skip Sunday, go to Saturday
	 }
	 return prevDate
}

// getNextDrawDate finds the date of the next draw (skipping Sunday)
func getNextDrawDate(currentDate time.Time) time.Time {
	 nextDate := currentDate.AddDate(0, 0, 1)
	 if nextDate.Weekday() == time.Sunday {
	 	 nextDate = nextDate.AddDate(0, 0, 1) // Skip Sunday, go to Monday
	 }
	 return nextDate
}

// maskMSISDN masks the phone number (first 3, last 3 visible)
func maskMSISDN(msisdn string) string {
	 if len(msisdn) < 6 {
	 	 return msisdn // Not long enough to mask properly
	 }
	 prefix := msisdn[:3]
	 suffix := msisdn[len(msisdn)-3:]
	 return prefix + "*****" + suffix
}

// Helper function to convert string slice to int slice
func stringSliceToIntSlice(s []string) ([]int, error) {
	 var intSlice []int
	 for _, str := range s {
	 	 i, err := strconv.Atoi(str)
	 	 if err != nil {
	 	 	 return nil, fmt.Errorf("invalid digit '%s': %w", str, err)
	 	 }
	 	 intSlice = append(intSlice, i)
	 }
	 return intSlice, nil
}

