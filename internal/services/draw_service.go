package services

import (
	"context"
	"encoding/json" // Added for prize structure parsing
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings" // Added for prize structure parsing
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"
)

// Compile-time check to ensure DrawServiceImpl implements DrawService
var _ DrawService = (*DrawServiceImpl)(nil)

// DrawServiceImpl handles draw-related business logic according to redesign plan
type DrawServiceImpl struct {
	 drawRepo             repositories.DrawRepository
	 userRepo             repositories.UserRepository
	 winnerRepo           repositories.WinnerRepository
	 blacklistRepo        repositories.BlacklistRepository
	 systemConfigRepo     repositories.SystemConfigRepository
	 pointTransactionRepo repositories.PointTransactionRepository
	 jackpotRolloverRepo  repositories.JackpotRolloverRepository
	 // userService          UserService // Might be needed for AllocatePointsForTopup
}

// NewDrawService creates a new DrawServiceImpl
func NewDrawService(
	 drawRepo repositories.DrawRepository,
	 userRepo repositories.UserRepository,
	 winnerRepo repositories.WinnerRepository,
	 blacklistRepo repositories.BlacklistRepository,
	 systemConfigRepo repositories.SystemConfigRepository,
	 pointTransactionRepo repositories.PointTransactionRepository,
	 jackpotRolloverRepo repositories.JackpotRolloverRepository,
	 // userService UserService,
) *DrawServiceImpl {
	return &DrawServiceImpl{
		 drawRepo:             drawRepo,
		 userRepo:             userRepo,
		 winnerRepo:           winnerRepo,
		 blacklistRepo:        blacklistRepo,
		 systemConfigRepo:     systemConfigRepo,
		 pointTransactionRepo: pointTransactionRepo,
		 jackpotRolloverRepo:  jackpotRolloverRepo,
		 // userService:          userService,
	}
}

// --- Core Draw Lifecycle Methods (Refactored & Refined) ---

// ScheduleDraw schedules a new draw, incorporating configuration and rollover logic
func (s *DrawServiceImpl) ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int, useDefaultDigits bool) (*models.Draw, error) {
	 // 1. Check if a draw already exists for this date
	 existingDraw, err := s.drawRepo.FindByDate(ctx, drawDate)
	 if err == nil && existingDraw != nil {
		 slog.Warn("Attempted to schedule draw for date with existing draw", "date", drawDate)
		 return existingDraw, errors.New("a draw already exists for this date")
	 }
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("Failed to check for existing draw", "error", err, "date", drawDate)
		 return nil, fmt.Errorf("failed to check for existing draw: %w", err)
	 }

	 // 2. Determine final eligible digits
	 finalEligibleDigits := eligibleDigits
	 if useDefaultDigits {
		 finalEligibleDigits = utils.GetDefaultEligibleDigits(drawDate.Weekday())
	 }

	 // 3. Fetch Prize Structure from System Config
	 prizes, err := s.GetPrizeStructure(ctx, drawType)
	 if err != nil {
		 return nil, err // Error already logged in GetPrizeStructure
	 }

	 // 4. Fetch Base Jackpot Amount from System Config
	 baseJackpotKey := "base_jackpot_" + strings.ToUpper(drawType)
	 baseJackpotConfig, err := s.systemConfigRepo.FindByKey(ctx, baseJackpotKey)
	 if err != nil {
		 slog.Error("Failed to fetch base jackpot config", "error", err, "key", baseJackpotKey)
		 return nil, fmt.Errorf("failed to fetch base jackpot config %s: %w", baseJackpotKey, err)
	 }
	 baseJackpotAmount, ok := baseJackpotConfig.Value.(float64) // Assuming stored as float64
	 if !ok {
		 slog.Error("Invalid base jackpot amount format in config", "key", baseJackpotKey, "valueType", fmt.Sprintf("%T", baseJackpotConfig.Value))
		 return nil, fmt.Errorf("invalid base jackpot amount format in config %s", baseJackpotKey)
	 }

	 // 5. Calculate Rollover Amount *into* this draw
	 accumulatedRollover := 0.0
	 rollovers, err := s.jackpotRolloverRepo.FindRolloversByDestinationDate(ctx, drawDate)
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("Failed to fetch incoming rollovers for draw date", "error", err, "date", drawDate)
		 // Decide if this is fatal. For now, log and continue with 0 rollover.
	 } else if err == nil {
		 for _, rollover := range rollovers {
			 accumulatedRollover += rollover.RolloverAmount
		 }
	 }

	 // 6. Calculate Final Jackpot Amount
	 calculatedJackpot := baseJackpotAmount + accumulatedRollover

	 // 7. Create the Draw object
	 draw := &models.Draw{
		 DrawDate:                drawDate,
		 DrawType:                strings.ToUpper(drawType),
		 EligibleDigits:          finalEligibleDigits,
		 UseDefaultDigits:        useDefaultDigits,
		 Status:                  models.DrawStatusScheduled,
		 Prizes:                  prizes, // Use prizes from config
		 BaseJackpotAmount:       baseJackpotAmount,
		 RolloverAmount:          accumulatedRollover,
		 CalculatedJackpotAmount: calculatedJackpot,
		 RolloverExecuted:        false,
		 CreatedAt:               time.Now(),
		 UpdatedAt:               time.Now(),
	 }

	 // 8. Save the Draw
	 err = s.drawRepo.Create(ctx, draw)
	 if err != nil {
		 slog.Error("Failed to create draw in repository", "error", err)
		 return nil, fmt.Errorf("failed to save scheduled draw: %w", err)
	 }

	 slog.Info("Draw scheduled successfully", "drawId", draw.ID, "date", drawDate, "type", drawType, "jackpot", calculatedJackpot)
	 return draw, nil
}

// ExecuteDraw executes a scheduled draw, including eligibility, selection, validation, and rollover
func (s *DrawServiceImpl) ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error) {
	 // 1. Get the Draw & Basic Validation
	 draw, err := s.drawRepo.FindByID(ctx, drawID)
	 if err != nil {
		 slog.Error("ExecuteDraw: Failed to find draw", "error", err, "drawId", drawID)
		 return nil, fmt.Errorf("draw not found: %w", err)
	 }

	 if draw.Status != models.DrawStatusScheduled {
		 slog.Warn("ExecuteDraw: Attempted to execute draw not in SCHEDULED state", "drawId", drawID, "status", draw.Status)
		 return draw, fmt.Errorf("draw is not in SCHEDULED state (current: %s)", draw.Status)
	 }

	 // 2. Update Draw Status to EXECUTING
	 draw.Status = models.DrawStatusExecuting
	 draw.ExecutionStartTime = time.Now()
	 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: Starting execution", time.Now().Format(time.RFC3339)))
	 err = s.drawRepo.Update(ctx, draw)
	 if err != nil {
		 slog.Error("ExecuteDraw: Failed to update draw status to EXECUTING", "error", err, "drawId", drawID)
		 // Attempt to return the original draw object on failure
		 originalDraw, _ := s.drawRepo.FindByID(ctx, drawID)
		 if originalDraw == nil {
			 originalDraw = draw // Fallback to the modified object if find fails
		 }
		 return originalDraw, fmt.Errorf("failed to mark draw as executing: %w", err)
	 }

	 // Defer status update on failure/completion
	 defer func() {
		 if r := recover(); r != nil {
			 draw.Status = models.DrawStatusFailed
			 draw.ErrorMessage = fmt.Sprintf("Panic during execution: %v", r)
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: PANIC: %v", time.Now().Format(time.RFC3339), r))
			 slog.Error("ExecuteDraw: Panic recovered", "panic", r, "drawId", drawID)
		 } else if err != nil {
			 draw.Status = models.DrawStatusFailed
			 draw.ErrorMessage = err.Error()
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: ERROR: %s", time.Now().Format(time.RFC3339), err.Error()))
			 slog.Error("ExecuteDraw: Execution failed", "error", err, "drawId", drawID)
		 } else {
			 draw.Status = models.DrawStatusCompleted
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: Execution completed successfully", time.Now().Format(time.RFC3339)))
			 slog.Info("ExecuteDraw: Execution completed", "drawId", drawID)
		 }
		 draw.ExecutionEndTime = time.Now()
		 updateErr := s.drawRepo.Update(ctx, draw)
		 if updateErr != nil {
			 slog.Error("ExecuteDraw: CRITICAL: Failed to update final draw status", "error", updateErr, "drawId", drawID, "finalStatusAttempt", draw.Status)
			 // If the final update fails, the error from the main execution (err) should be returned
			 if err == nil {
				 err = fmt.Errorf("failed to update final draw status: %w", updateErr)
			 }
		 }
	 }()

	 // 3. Determine Eligibility Time Windows
	 eligibilityCutoff := time.Date(draw.DrawDate.Year(), draw.DrawDate.Month(), draw.DrawDate.Day(), 18, 0, 0, 0, draw.DrawDate.Location())
	 var eligibilityStart time.Time
	 if draw.DrawType == "SATURDAY" {
		 prevSaturday := draw.DrawDate.AddDate(0, 0, -7)
		 eligibilityStart = time.Date(prevSaturday.Year(), prevSaturday.Month(), prevSaturday.Day(), 18, 0, 1, 0, draw.DrawDate.Location())
	 } else {
		 // Daily draw eligibility starts at 00:00:00 on the draw day
		 eligibilityStart = time.Date(draw.DrawDate.Year(), draw.DrawDate.Month(), draw.DrawDate.Day(), 0, 0, 0, 0, draw.DrawDate.Location())
	 }
	 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Eligibility window: %s to %s", eligibilityStart.Format(time.RFC3339), eligibilityCutoff.Format(time.RFC3339)))

	 // 4. Fetch Participant Pools
	 // Pool A (Jackpot): All users with *any* recharge in the window
	 // Note: FindUsersByRechargeWindow currently returns all users (placeholder)
	 poolA, err := s.userRepo.FindUsersByRechargeWindow(ctx, eligibilityStart, eligibilityCutoff)
	 if err != nil {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Failed to fetch Pool A")
		 return draw, fmt.Errorf("failed to fetch jackpot participant pool: %w", err)
	 }
	 draw.TotalParticipants = len(poolA)
	 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Fetched Pool A (Jackpot Pool): %d users", len(poolA)))

	 // Pool B (Consolation): Opt-in users meeting all criteria
	 // Note: FindEligibleConsolationUsers has placeholder filters for recharge/blacklist
	 poolB, err := s.userRepo.FindEligibleConsolationUsers(ctx, draw.EligibleDigits, eligibilityCutoff, eligibilityStart, eligibilityCutoff)
	 if err != nil {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Failed to fetch Pool B")
		 return draw, fmt.Errorf("failed to fetch consolation participant pool: %w", err)
	 }
	 draw.EligibleOptedInParticipants = len(poolB)
	 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Fetched Pool B (Consolation Pool): %d users", len(poolB)))

	 // 5. Select Jackpot Winner (if Pool A not empty)
	 var potentialJackpotWinner *models.User
	 if len(poolA) > 0 {
		 // Use points-weighted selection for Jackpot Pool A as well (REQFUNC027)
		 weightedPoolA := createWeightedPool(poolA)
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Created weighted pool for jackpot winner (total weight: %d)", len(weightedPoolA)))
		 var selectionErr error
		 potentialJackpotWinner, _, selectionErr = selectWeightedWinner(weightedPoolA)
		 if selectionErr != nil {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR selecting weighted jackpot winner: %s", selectionErr.Error()))
			 return draw, fmt.Errorf("failed to select weighted jackpot winner: %w", selectionErr)
		 }
		 draw.JackpotWinnerMsisdn = potentialJackpotWinner.MSISDN
		 draw.JackpotWinnerValidationStatus = models.JackpotValidationPending
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Potential Jackpot Winner Selected (Weighted): %s (Points: %d)", maskMsisdn(potentialJackpotWinner.MSISDN), potentialJackpotWinner.Points))
	 } else {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Pool A is empty, cannot select Jackpot Winner.")
		 draw.JackpotWinnerValidationStatus = "NO_PARTICIPANTS" // Use a string constant or define in models
	 }

	 // 6. Validate Jackpot Winner & Handle Rollover
	 isJackpotWinnerValid := false
	 if potentialJackpotWinner != nil {
		 // Check Opt-in status and timing
		 if potentialJackpotWinner.OptInStatus && !potentialJackpotWinner.OptInDate.IsZero() && potentialJackpotWinner.OptInDate.Before(eligibilityCutoff) {
			 isJackpotWinnerValid = true
			 draw.JackpotWinnerValidationStatus = models.JackpotValidationValid
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Jackpot Winner Validated: %s (Opted-in)", maskMsisdn(potentialJackpotWinner.MSISDN)))
		 } else {
			 isJackpotWinnerValid = false
			 draw.JackpotWinnerValidationStatus = models.JackpotValidationInvalidNotOptIn
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Jackpot Winner Invalid: %s (Not Opted-in or Opted-in too late)", maskMsisdn(potentialJackpotWinner.MSISDN)))

			 // --- Trigger Rollover Logic ---
			 draw.RolloverExecuted = true
			 rolloverAmount := draw.CalculatedJackpotAmount
			 var destinationDate time.Time
			 // Find the next scheduled draw to determine destination date
			 nextDraw, findNextErr := s.drawRepo.FindNextScheduledDraw(ctx, draw.DrawDate)
			 if findNextErr != nil {
				 slog.Error("Failed to find next scheduled draw for rollover destination", "error", findNextErr, "sourceDrawId", draw.ID)
				 // Fallback: Use simple date logic if next draw isn't found (might be inaccurate if schedule changes)
				 if draw.DrawType == "SATURDAY" {
					 destinationDate = draw.DrawDate.AddDate(0, 0, 7)
				 } else {
					 daysUntilSaturday := time.Saturday - draw.DrawDate.Weekday()
					 if daysUntilSaturday <= 0 {
						 daysUntilSaturday += 7
					 }
					 destinationDate = draw.DrawDate.AddDate(0, 0, int(daysUntilSaturday))
				 }
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("WARN: Could not find next scheduled draw, using calculated date %s for rollover", destinationDate.Format("2006-01-02")))
			 } else {
				 destinationDate = nextDraw.DrawDate
			 }

			 rolloverRecord := &models.JackpotRollover{
				 SourceDrawID:        draw.ID,
				 SourceDrawDate:      draw.DrawDate,
				 RolloverAmount:      rolloverAmount,
				 DestinationDrawDate: destinationDate,
				 Reason:              string(models.JackpotValidationInvalidNotOptIn), // Convert status to string
				 CreatedAt:           time.Now(),
			 }
			 err = s.jackpotRolloverRepo.Create(ctx, rolloverRecord)
			 if err != nil {
				 slog.Error("Failed to create jackpot rollover record", "error", err, "sourceDrawId", draw.ID)
				 // Log error but don't fail the entire draw execution
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR creating rollover record: %s", err.Error()))
			 } else {
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Rollover Triggered: Amount %.2f to %s", rolloverAmount, destinationDate.Format("2006-01-02")))
			 }
		 }
	 }

	 // 7. Select Consolation Winners (Pool B, Points Weighted)
	 var consolationWinners []*models.Winner
	 if len(poolB) > 0 {
		 weightedPoolB := createWeightedPool(poolB)
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Created weighted pool for consolation winners (total weight: %d)", len(weightedPoolB)))

		 selectedConsolationMSISDNs := make(map[string]bool)
		 // Ensure Jackpot winner (even if invalid) isn't selected for consolation (REQFUNC039)
		 if potentialJackpotWinner != nil {
			 selectedConsolationMSISDNs[potentialJackpotWinner.MSISDN] = true
		 }

		 for _, prize := range draw.Prizes {
			 // Corrected: Check against models.JackpotCategory constant
			 if prize.Category == models.JackpotCategory { // Skip jackpot prize here
				 continue
			 }

			 for i := 0; i < prize.NumWinners; i++ {
				 if len(weightedPoolB) == 0 {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Weighted pool B exhausted while selecting for %s prize", prize.Category))
					 break // Stop selecting for this prize category if pool is empty
				 }

				 winnerUser, remainingPool, selectionErr := selectWeightedWinner(weightedPoolB)
				 weightedPoolB = remainingPool // Update pool for next selection

				 if selectionErr != nil { // Corrected syntax error here
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR selecting weighted consolation winner for %s: %s", prize.Category, selectionErr.Error()))
					 // Decide if this should halt the draw. For now, log and continue selecting for this category.
					 continue
				 } // Added missing closing brace here

				 // Check if user was already selected for another consolation prize in this draw
				 if selectedConsolationMSISDNs[winnerUser.MSISDN] {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Skipping user %s for %s prize (already won consolation)", maskMsisdn(winnerUser.MSISDN), prize.Category))
					 i-- // Decrement i to try selecting another winner for this slot
					 continue
				 }

				 // Check blacklist status
				 isBlacklisted, blacklistErr := s.blacklistRepo.IsBlacklisted(ctx, winnerUser.MSISDN)
				 if blacklistErr != nil {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR checking blacklist for %s: %s", maskMsisdn(winnerUser.MSISDN), blacklistErr.Error()))
					 // Decide if this should halt the draw. For now, log and skip this potential winner.
					 i-- // Decrement i to try selecting another winner for this slot
					 continue
				 }
				 if isBlacklisted {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Skipping user %s for %s prize (blacklisted)", maskMsisdn(winnerUser.MSISDN), prize.Category))
					 i-- // Decrement i to try selecting another winner for this slot
					 continue
				 }

				 // Winner selected and valid for this prize
				 selectedConsolationMSISDNs[winnerUser.MSISDN] = true
				 winner := &models.Winner{
					 DrawID:        draw.ID,
					 UserID:        winnerUser.ID,
					 MSISDN:        winnerUser.MSISDN,
					 PrizeCategory: prize.Category,
					 PrizeAmount:   prize.Amount,
					 WinDate:       draw.DrawDate,
					 ClaimStatus:   models.ClaimStatusPending, // Default status
					 CreatedAt:     time.Now(),
					 UpdatedAt:     time.Now(),
				 }
				 consolationWinners = append(consolationWinners, winner)
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Consolation Winner Selected (%s): %s (Points: %d)", prize.Category, maskMsisdn(winnerUser.MSISDN), winnerUser.Points))
			 }
		 }
	 } else {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Pool B is empty, cannot select Consolation Winners.")
	 }

	 // 8. Save Winners (Jackpot if valid, Consolation)
	 var winnersToSave []*models.Winner
	 if isJackpotWinnerValid {
		 jackpotPrize, found := findPrizeByCategory(draw.Prizes, models.JackpotCategory)
		 if !found {
			 draw.ExecutionLog = append(draw.ExecutionLog, "ERROR: Jackpot prize category not found in draw config!")
			 return draw, errors.New("jackpot prize category not found in draw configuration")
		 }
		 jackpotWinner := &models.Winner{
			 DrawID:        draw.ID,
			 UserID:        potentialJackpotWinner.ID,
			 MSISDN:        potentialJackpotWinner.MSISDN,
			 PrizeCategory: models.JackpotCategory,
			 PrizeAmount:   jackpotPrize.Amount, // Use amount from draw's prize structure
			 WinDate:       draw.DrawDate,
			 ClaimStatus:   models.ClaimStatusPending, // Default status
			 CreatedAt:     time.Now(),
			 UpdatedAt:     time.Now(),
		 }
		 winnersToSave = append(winnersToSave, jackpotWinner)
	 }
	 winnersToSave = append(winnersToSave, consolationWinners...)

	 if len(winnersToSave) > 0 {
		 err = s.winnerRepo.CreateMany(ctx, winnersToSave)
		 if err != nil {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR saving winners: %s", err.Error()))
			 return draw, fmt.Errorf("failed to save winners: %w", err)
		 }
		 draw.NumWinners = len(winnersToSave) // Corrected: Use NumWinners field
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Saved %d winners to database", len(winnersToSave)))
	 } else {
		 draw.NumWinners = 0 // Corrected: Use NumWinners field
		 draw.ExecutionLog = append(draw.ExecutionLog, "No winners to save.")
	 }

	 // 9. Final status update is handled by the deferred function
	 return draw, nil // Return nil error if execution reaches here
}

// --- Helper & Read Methods ---

// GetDrawByID retrieves a single draw by its ID
func (s *DrawServiceImpl) GetDrawByID(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error) {
	 return s.drawRepo.FindByID(ctx, drawID)
}

// GetWinnersByDrawID retrieves all winners for a specific draw ID
func (s *DrawServiceImpl) GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	 return s.winnerRepo.FindByDrawID(ctx, drawID)
}

// GetDraws retrieves draws within a specified date range
func (s *DrawServiceImpl) GetDraws(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error) {
	 return s.drawRepo.FindByDateRange(ctx, startDate, endDate)
}

// GetPrizeStructure retrieves the prize structure for a given draw type from system config
func (s *DrawServiceImpl) GetPrizeStructure(ctx context.Context, drawType string) ([]models.Prize, error) {
	 key := "prize_structure_" + strings.ToUpper(drawType)
	 config, err := s.systemConfigRepo.FindByKey(ctx, key)
	 if err != nil {
		 slog.Error("Failed to fetch prize structure config", "error", err, "key", key)
		 return nil, fmt.Errorf("failed to fetch prize structure config %s: %w", key, err)
	 }

	 // Assuming the structure is stored as a JSON string
	 structureJSON, ok := config.Value.(string)
	 if !ok {
		 slog.Error("Invalid prize structure format in config (expected JSON string)", "key", key, "valueType", fmt.Sprintf("%T", config.Value))
		 return nil, fmt.Errorf("invalid prize structure format in config %s", key)
	 }

	 var prizes []models.Prize
	 err = json.Unmarshal([]byte(structureJSON), &prizes)
	 if err != nil {
		 slog.Error("Failed to unmarshal prize structure JSON", "error", err, "key", key, "json", structureJSON)
		 return nil, fmt.Errorf("failed to parse prize structure config %s: %w", key, err)
	 }

	 return prizes, nil
}

// UpdatePrizeStructure updates the prize structure for a given draw type in system config
func (s *DrawServiceImpl) UpdatePrizeStructure(ctx context.Context, drawType string, structure []models.Prize) error {
	 key := "prize_structure_" + strings.ToUpper(drawType)

	 // Marshal the structure back to JSON string
	 structureJSON, err := json.Marshal(structure)
	 if err != nil {
		 slog.Error("Failed to marshal prize structure to JSON", "error", err, "key", key)
		 return fmt.Errorf("failed to serialize prize structure for %s: %w", key, err)
	 }

	 // Upsert the JSON string into system config
	 err = s.systemConfigRepo.UpsertByKey(ctx, key, string(structureJSON))
	 if err != nil {
		 slog.Error("Failed to upsert prize structure config", "error", err, "key", key)
		 return fmt.Errorf("failed to update prize structure config %s: %w", key, err)
	 }

	 slog.Info("Prize structure updated successfully", "key", key)
	 return nil
}

// GetDefaultDigitsForDay retrieves the default eligible digits for a given day of the week.
func (s *DrawServiceImpl) GetDefaultDigitsForDay(ctx context.Context, dayOfWeek time.Weekday) ([]int, error) {
	 // This currently uses a utility function. If it needed DB access, it would go here.
	 digits := utils.GetDefaultEligibleDigits(dayOfWeek)
	 return digits, nil // Return nil error as it's a simple calculation
}

// GetDrawByDate retrieves a draw for a specific date.
func (s *DrawServiceImpl) GetDrawByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	 // Normalize date to the start of the day to ensure consistent matching
	 startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	 return s.drawRepo.FindByDate(ctx, startOfDay)
}

// GetJackpotStatus retrieves the current jackpot status
func (s *DrawServiceImpl) GetJackpotStatus(ctx context.Context) (*models.JackpotStatus, error) {
	 status := &models.JackpotStatus{
		 CurrentAmount: 0.0, // Default
		 LastUpdatedAt: time.Time{}, // Default
	 }

	 // 1. Find the latest completed or scheduled Saturday draw
	 latestSaturdayDraw, err := s.drawRepo.FindLatestDrawByTypeAndStatus(
		 ctx,
		 "SATURDAY",
		 []string{string(models.DrawStatusCompleted), string(models.DrawStatusScheduled)}, // Corrected: Convert DrawStatus to string
	 )
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to find latest Saturday draw", "error", err)
		 return nil, fmt.Errorf("failed to get latest Saturday draw: %w", err)
	 }

	 // 2. If a Saturday draw exists, use its jackpot amount
	 if latestSaturdayDraw != nil {
		 status.CurrentAmount = latestSaturdayDraw.CalculatedJackpotAmount
		 status.LastUpdatedAt = latestSaturdayDraw.UpdatedAt // Use draw's update time
	 } else {
		 // 3. If no Saturday draw, check the base config (less ideal, but a fallback)
		 baseJackpotConfig, err := s.systemConfigRepo.FindByKey(ctx, "base_jackpot_SATURDAY")
		 if err != nil {
			 slog.Error("GetJackpotStatus: Failed to fetch base jackpot config as fallback", "error", err)
			 // Return default status or error?
			 return status, fmt.Errorf("failed to get base jackpot config: %w", err)
		 }
		 baseAmount, ok := baseJackpotConfig.Value.(float64)
		 if !ok {
			 slog.Error("GetJackpotStatus: Invalid base jackpot amount format in config")
			 return status, errors.New("invalid base jackpot amount format in config")
		 }
		 status.CurrentAmount = baseAmount
		 status.LastUpdatedAt = baseJackpotConfig.UpdatedAt // Use config update time
	 }

	 // 4. Add any pending rollovers targeting a future date (after the last draw's date)
	 var effectiveDate time.Time
	 if latestSaturdayDraw != nil {
		 effectiveDate = latestSaturdayDraw.DrawDate
	 }
	 // Find rollovers created after the last draw date OR targeting a future date
	 pendingRollovers, err := s.jackpotRolloverRepo.FindPendingRollovers(ctx, effectiveDate)
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to fetch pending rollovers", "error", err)
		 // Continue without adding pending rollovers
	 } else if err == nil {
		 for _, rollover := range pendingRollovers {
			 status.CurrentAmount += rollover.RolloverAmount
			 // Update LastUpdatedAt if rollover is more recent
			 if rollover.CreatedAt.After(status.LastUpdatedAt) {
				 status.LastUpdatedAt = rollover.CreatedAt
			 }
		 }
	 }

	 return status, nil
}

// AllocatePointsForTopup calculates and allocates points for a user based on a top-up amount.
// This is now part of DrawService as points are intrinsically linked to draw eligibility/weighting.
func (s *DrawServiceImpl) AllocatePointsForTopup(ctx context.Context, userID primitive.ObjectID, amount float64, transactionTime time.Time) (int, error) {
	 // 1. Calculate Points (using the logic previously in topup_service)
	 var pointsToAdd int
	 if amount >= 1000 {
		 pointsToAdd = 10 // 10 points for N1000 or more
	 } else {
		 // 1 point for every N100
		 pointsToAdd = int(amount / 100)
	 }

	 if pointsToAdd <= 0 {
		 slog.Info("No points to add for top-up", "userId", userID, "amount", amount)
		 return 0, nil // No points to add, not an error
	 }

	 // 2. Create Point Transaction Record
	 transaction := &models.PointTransaction{
		 UserID:          userID,
		 PointsAwarded:   pointsToAdd, // Corrected field name
		 Type:            models.TransactionTypeTopup, // Corrected field name
		 Source:          fmt.Sprintf("Points for top-up of %.2f", amount), // Corrected field name
		 TransactionTime: transactionTime, // Corrected field name
		 CreatedAt:       time.Now(),
	 }
	 err := s.pointTransactionRepo.Create(ctx, transaction)
	 if err != nil {
		 slog.Error("Failed to create point transaction record", "error", err, "userId", userID, "points", pointsToAdd)
		 return 0, fmt.Errorf("failed to record point transaction: %w", err)
	 }

	 // 3. Update User's Total Points
	 err = s.userRepo.IncrementPoints(ctx, userID, pointsToAdd)
	 if err != nil {
		 slog.Error("Failed to update user points balance", "error", err, "userId", userID, "pointsToAdd", pointsToAdd)
		 // Consider compensating transaction if user update fails?
		 return 0, fmt.Errorf("failed to update user points: %w", err)
	 }

	 slog.Info("Points allocated successfully", "userId", userID, "pointsAdded", pointsToAdd, "topupAmount", amount)
	 return pointsToAdd, nil
}

// --- Internal Helper Functions ---

// createWeightedPool creates a slice where each user appears once for every point they have.
func createWeightedPool(users []*models.User) []*models.User {
	 var weightedPool []*models.User
	 for _, user := range users {
		 // Ensure user has at least 1 entry even if points are 0?
		 // Requirement REQFUNC027 implies weighting by points.
		 // If points can be 0, should they be excluded or have a base weight?
		 // Assuming points > 0 for weighting.
		 weight := user.Points
		 if weight <= 0 {
			 weight = 1 // Give at least one chance if points are 0 or negative?
		 }
		 for i := 0; i < weight; i++ {
			 weightedPool = append(weightedPool, user)
		 }
	 }
	 return weightedPool
}

// selectWeightedWinner selects a random winner from a weighted pool and returns the winner and the pool without the winner.
func selectWeightedWinner(weightedPool []*models.User) (*models.User, []*models.User, error) {
	 if len(weightedPool) == 0 {
		 return nil, weightedPool, errors.New("weighted pool is empty")
	 }

	 winnerIndex := rand.Intn(len(weightedPool))
	 winner := weightedPool[winnerIndex]

	 // Create a new pool excluding all entries for the selected winner
	 var remainingPool []*models.User
	 for _, user := range weightedPool {
		 if user.ID != winner.ID {
			 remainingPool = append(remainingPool, user)
		 }
	 }

	 return winner, remainingPool, nil
}

// findPrizeByCategory searches for a prize category in a slice of prizes.
func findPrizeByCategory(prizes []models.Prize, category string) (models.Prize, bool) {
	 for _, p := range prizes {
		 if p.Category == category {
			 return p, true
		 }
	 }
	 return models.Prize{}, false
}

// maskMsisdn masks an MSISDN for logging (e.g., "234803***1234")
func maskMsisdn(msisdn string) string {
	 if len(msisdn) > 7 {
		 return msisdn[:6] + "***" + msisdn[len(msisdn)-4:]
	 }
	 return msisdn // Return original if too short to mask reasonably
}


}



