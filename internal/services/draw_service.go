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
	 updateErr := s.drawRepo.Update(ctx, draw) // Use a different variable name here
	 if updateErr != nil {
		 slog.Error("ExecuteDraw: Failed to update draw status to EXECUTING", "error", updateErr, "drawId", drawID)
		 // Attempt to return the original draw object on failure
		 originalDraw, findErr := s.drawRepo.FindByID(ctx, drawID)
		 if findErr != nil || originalDraw == nil {
			 originalDraw = draw // Fallback to the modified object if find fails
		 }
		 return originalDraw, fmt.Errorf("failed to mark draw as executing: %w", updateErr)
	 }

	 // Defer status update on failure/completion
	 defer func() {
		 finalStatus := models.DrawStatusCompleted
		 if r := recover(); r != nil {
			 finalStatus = models.DrawStatusFailed
			 draw.ErrorMessage = fmt.Sprintf("Panic during execution: %v", r)
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: PANIC: %v", time.Now().Format(time.RFC3339), r))
			 slog.Error("ExecuteDraw: Panic recovered", "panic", r, "drawId", drawID)
			 // Ensure 'err' is set if it wasn't already (to prevent overwriting by defer update error)
			 if err == nil {
				 err = fmt.Errorf("panic during execution: %v", r)
			 }
		 } else if err != nil {
			 finalStatus = models.DrawStatusFailed
			 draw.ErrorMessage = err.Error()
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: ERROR: %s", time.Now().Format(time.RFC3339), err.Error()))
			 slog.Error("ExecuteDraw: Execution failed", "error", err, "drawId", drawID)
		 } else {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: Execution completed successfully", time.Now().Format(time.RFC3339)))
			 slog.Info("ExecuteDraw: Execution completed", "drawId", drawID)
		 }

		 draw.Status = finalStatus
		 draw.ExecutionEndTime = time.Now()
		 finalUpdateErr := s.drawRepo.Update(ctx, draw)
		 if finalUpdateErr != nil {
			 slog.Error("ExecuteDraw: CRITICAL: Failed to update final draw status", "error", finalUpdateErr, "drawId", drawID, "finalStatusAttempt", draw.Status)
			 // If the final update fails, the original error (err) takes precedence
			 if err == nil {
				 err = fmt.Errorf("failed to update final draw status: %w", finalUpdateErr)
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
	 poolA, err := s.userRepo.FindUsersByRechargeWindow(ctx, eligibilityStart, eligibilityCutoff)
	 if err != nil {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Failed to fetch Pool A")
		 return draw, fmt.Errorf("failed to fetch jackpot participant pool: %w", err)
	 }
	 draw.TotalParticipants = len(poolA)
	 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Fetched Pool A (Jackpot Pool): %d users", len(poolA)))

	 // Pool B (Consolation): Opt-in users meeting all criteria
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
		 draw.JackpotWinnerValidationStatus = models.JackpotValidationNoParticipants // Use defined constant
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
			 // Use 'err' for the rollover creation error, shadowing the outer 'err'
			 rolloverErr := s.jackpotRolloverRepo.Create(ctx, rolloverRecord)
			 if rolloverErr != nil {
				 slog.Error("Failed to create jackpot rollover record", "error", rolloverErr, "sourceDrawId", draw.ID)
				 // Log error but don't fail the entire draw execution
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR creating rollover record: %s", rolloverErr.Error()))
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

				 if selectionErr != nil {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR selecting weighted consolation winner for %s: %s", prize.Category, selectionErr.Error()))
					 // Decide if this should halt the draw. For now, log and continue selecting for this category.
					 continue
				 }

				 // Check if user was already selected for another consolation prize in this draw
				 if selectedConsolationMSISDNs[winnerUser.MSISDN] {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Skipping already selected user %s for %s prize", maskMsisdn(winnerUser.MSISDN), prize.Category))
					 i-- // Decrement i to retry selecting for this specific slot
					 continue // Continue to the next iteration of the inner loop
				 }

				 // If user is not already selected, record them as a winner
				 selectedConsolationMSISDNs[winnerUser.MSISDN] = true
				 winner := &models.Winner{
					 DrawID:       draw.ID,
					 UserID:       winnerUser.ID,
					 MSISDN:       winnerUser.MSISDN,
					 PrizeCategory: prize.Category,
					 PrizeAmount:  prize.Amount,
					 WinDate:      draw.DrawDate, // Use DrawDate as WinDate
					 Status:       models.WinnerStatusPending, // Or determine appropriate initial status
					 CreatedAt:    time.Now(),
					 UpdatedAt:    time.Now(),
				 }
				 consolationWinners = append(consolationWinners, winner)
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Consolation Winner Selected (%s): %s (Points: %d)", prize.Category, maskMsisdn(winnerUser.MSISDN), winnerUser.Points))

			 } // Close inner loop: for i := 0; i < prize.NumWinners; i++
		 } // Close outer loop: for _, prize := range draw.Prizes
	 } // Close if len(poolB) > 0

	 // 8. Save Winners (if any)
	 var saveErr error // Declare error variable for saving winners
	 if len(consolationWinners) > 0 {
		 saveErr = s.winnerRepo.CreateMany(ctx, consolationWinners)
		 if saveErr != nil {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR saving consolation winners: %s", saveErr.Error()))
			 // Decide if this is fatal. For now, log and continue.
			 // Assign saveErr to the main 'err' if it's currently nil, so defer catches it
			 if err == nil {
				 err = fmt.Errorf("failed to save consolation winners: %w", saveErr)
			 }
		 } else {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Saved %d consolation winners", len(consolationWinners)))
		 }
		 draw.NumWinners = len(consolationWinners) // Update count based on actual winners

		 if isJackpotWinnerValid {
			 // If jackpot winner is valid, also save them
			 jackpotWinnerRecord := &models.Winner{
				 DrawID:       draw.ID,
				 UserID:       potentialJackpotWinner.ID,
				 MSISDN:       potentialJackpotWinner.MSISDN,
				 PrizeCategory: models.JackpotCategory,
				 PrizeAmount:  draw.CalculatedJackpotAmount, // Use calculated amount
				 WinDate:      draw.DrawDate,
				 Status:       models.WinnerStatusPending, // Or Validated? Needs clarification
				 CreatedAt:    time.Now(),
				 UpdatedAt:    time.Now(),
			 }
			 jackpotSaveErr := s.winnerRepo.Create(ctx, jackpotWinnerRecord)
			 if jackpotSaveErr != nil {
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR saving jackpot winner: %s", jackpotSaveErr.Error()))
				 // Log error but don't fail the draw; assign to main 'err' if nil
				 if err == nil {
					 err = fmt.Errorf("failed to save jackpot winner: %w", jackpotSaveErr)
				 }
			 } else {
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Saved valid jackpot winner %s", maskMsisdn(potentialJackpotWinner.MSISDN)))
				 draw.NumWinners++ // Increment total winner count
			 }
		 }
	 } else if isJackpotWinnerValid {
		 // Only jackpot winner, save them
		 jackpotWinnerRecord := &models.Winner{
			 DrawID:       draw.ID,
			 UserID:       potentialJackpotWinner.ID,
			 MSISDN:       potentialJackpotWinner.MSISDN,
			 PrizeCategory: models.JackpotCategory,
			 PrizeAmount:  draw.CalculatedJackpotAmount,
			 WinDate:      draw.DrawDate,
			 Status:       models.WinnerStatusPending,
			 CreatedAt:    time.Now(),
			 UpdatedAt:    time.Now(),
		 }
		 jackpotSaveErr := s.winnerRepo.Create(ctx, jackpotWinnerRecord)
		 if jackpotSaveErr != nil {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR saving jackpot winner: %s", jackpotSaveErr.Error()))
			 // Assign to main 'err' if nil
			 if err == nil {
				 err = fmt.Errorf("failed to save jackpot winner: %w", jackpotSaveErr)
			 }
		 } else {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Saved valid jackpot winner %s", maskMsisdn(potentialJackpotWinner.MSISDN)))
			 draw.NumWinners = 1 // Set total winner count
		 }
	 }

	 // 9. Final logging and return (handled by defer)
	 return draw, err // Return the potentially modified draw object and the final error status

} // Close func ExecuteDraw

// --- Helper & Other Service Methods ---

// GetPrizeStructure retrieves the prize structure for a given draw type
func (s *DrawServiceImpl) GetPrizeStructure(ctx context.Context, drawType string) ([]models.Prize, error) {
	 configKey := "prize_structure_" + strings.ToUpper(drawType)
	 config, err := s.systemConfigRepo.FindByKey(ctx, configKey)
	 if err != nil {
		 slog.Error("Failed to fetch prize structure config", "error", err, "key", configKey)
		 return nil, fmt.Errorf("failed to fetch prize structure config %s: %w", configKey, err)
	 }

	 // Assuming the prize structure is stored as a JSON string in the config value
	 jsonString, ok := config.Value.(string)
	 if !ok {
		 slog.Error("Invalid prize structure format in config: expected JSON string", "key", configKey, "valueType", fmt.Sprintf("%T", config.Value))
		 return nil, fmt.Errorf("invalid prize structure format in config %s: expected JSON string", configKey)
	 }

	 var prizes []models.Prize
	 err = json.Unmarshal([]byte(jsonString), &prizes)
	 if err != nil {
		 slog.Error("Failed to unmarshal prize structure JSON", "error", err, "key", configKey, "jsonString", jsonString)
		 return nil, fmt.Errorf("failed to parse prize structure JSON for %s: %w", configKey, err)
	 }

	 return prizes, nil
}

// UpdatePrizeStructure updates the prize structure for a given draw type
func (s *DrawServiceImpl) UpdatePrizeStructure(ctx context.Context, drawType string, prizes []models.Prize) error {
	 configKey := "prize_structure_" + strings.ToUpper(drawType)

	 // Marshal the prizes slice into a JSON string
	 jsonBytes, err := json.Marshal(prizes)
	 if err != nil {
		 slog.Error("Failed to marshal prize structure to JSON", "error", err, "drawType", drawType)
		 return fmt.Errorf("failed to encode prize structure for %s: %w", drawType, err)
	 }
	 jsonString := string(jsonBytes)

	 // Upsert the JSON string into the system config
	 err = s.systemConfigRepo.UpsertByKey(ctx, configKey, jsonString)
	 if err != nil {
		 slog.Error("Failed to upsert prize structure config", "error", err, "key", configKey)
		 return fmt.Errorf("failed to update prize structure config %s: %w", configKey, err)
	 }

	 slog.Info("Prize structure updated successfully", "drawType", drawType)
	 return nil
}

// GetJackpotStatus retrieves the current jackpot status
func (s *DrawServiceImpl) GetJackpotStatus(ctx context.Context) (*models.JackpotStatus, error) {
	 status := &models.JackpotStatus{}
	 now := time.Now()

	 // 1. Find the latest completed Saturday draw
	 latestSaturdayDraw, err := s.drawRepo.FindLatestDrawByTypeAndStatus(ctx, "SATURDAY", []string{string(models.DrawStatusCompleted)})
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to find latest completed Saturday draw", "error", err)
		 return nil, fmt.Errorf("failed to query latest Saturday draw: %w", err)
	 }

	 if latestSaturdayDraw != nil {
		 status.LastDrawDate = latestSaturdayDraw.DrawDate
		 // Find the jackpot winner for this draw
		 jackpotWinners, findWinnerErr := s.winnerRepo.FindByDrawIDAndCategory(ctx, latestSaturdayDraw.ID, models.JackpotCategory)
		 if findWinnerErr != nil && !errors.Is(findWinnerErr, mongo.ErrNoDocuments) {
			 slog.Error("GetJackpotStatus: Failed to find jackpot winner for last draw", "error", findWinnerErr, "drawId", latestSaturdayDraw.ID)
			 // Continue, but status might be incomplete
		 } else if len(jackpotWinners) > 0 {
			 // Assuming only one jackpot winner per draw
			 status.LastWinnerMSISDN = jackpotWinners[0].MSISDN // Corrected field name
			 status.LastWinAmount = jackpotWinners[0].PrizeAmount // Corrected field name
		 }
	 }

	 // 2. Find the next scheduled Saturday draw
	 nextSaturdayDraw, err := s.drawRepo.FindNextScheduledDrawByType(ctx, now, "SATURDAY")
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to find next scheduled Saturday draw", "error", err)
		 return nil, fmt.Errorf("failed to query next Saturday draw: %w", err)
	 }

	 if nextSaturdayDraw == nil {
		 // If no next draw is scheduled, maybe calculate based on current date?
		 // For now, leave it empty or return an error/specific status
		 slog.Warn("GetJackpotStatus: No next Saturday draw found in scheduled state")
		 // status.NextDrawDate = calculateNextSaturday(now) // Placeholder for calculation logic
		 status.CurrentJackpotAmount = 0 // Or fetch default base if no next draw?
	 } else {
		 status.NextDrawDate = nextSaturdayDraw.DrawDate         // Corrected field name
		 status.CurrentJackpotAmount = nextSaturdayDraw.CalculatedJackpotAmount
	 }

	 // 3. Add pending rollovers to the *current* jackpot amount (amount for the *next* draw)
	 // We need rollovers destined for *after* the last completed draw up to the next scheduled draw
	 effectiveDate := time.Time{} // Start from beginning if no last draw
	 if latestSaturdayDraw != nil {
		 effectiveDate = latestSaturdayDraw.DrawDate
	 }
	 pendingRollovers, err := s.jackpotRolloverRepo.FindPendingRollovers(ctx, effectiveDate)
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to fetch pending rollovers", "error", err)
		 // Continue, jackpot amount might not include all pending rollovers
	 } else {
		 for _, rollover := range pendingRollovers {
			 // Only add rollovers destined for the *next* scheduled draw we found
			 if nextSaturdayDraw != nil && rollover.DestinationDrawDate.Equal(nextSaturdayDraw.DrawDate) {
				 // This logic seems redundant as FindPendingRollovers should already give relevant ones
				 // And the next draw's CalculatedJackpotAmount should already include these.
				 // Let's rely on nextSaturdayDraw.CalculatedJackpotAmount from step 2.
				 // status.CurrentJackpotAmount += rollover.RolloverAmount
			 }
		 }
	 }

	 return status, nil
}

// AllocatePointsForTopup allocates points based on top-up amount
// TODO: Refine this - should it use UserService? How are users identified/created?
func (s *DrawServiceImpl) AllocatePointsForTopup(ctx context.Context, msisdn string, amount float64, source string) error {
	 // 1. Find or Create User (Simplified - assumes user exists for now)
	 user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) {
			 // TODO: User creation logic needed here or in UserService
			 slog.Warn("AllocatePointsForTopup: User not found, cannot allocate points", "msisdn", msisdn)
			 return fmt.Errorf("user %s not found", msisdn)
		 } else {
			 slog.Error("AllocatePointsForTopup: Failed to find user", "error", err, "msisdn", msisdn)
			 return fmt.Errorf("failed to find user %s: %w", msisdn, err)
		 }
	 }

	 // 2. Calculate Points (Using the standalone utility function)
	 points := utils.CalculatePoints(amount)
	 if points <= 0 {
		 slog.Info("AllocatePointsForTopup: No points awarded for amount", "msisdn", msisdn, "amount", amount)
		 return nil // No error, just no points
	 }

	 // 3. Create Point Transaction Record
	 transaction := &models.PointTransaction{
		 UserID:          user.ID,
		 MSISDN:          msisdn,
		 PointsAwarded:   points,                     // Corrected field name
		 TransactionType: models.TransactionTypeTopup, // Corrected field name & Use constant
		 Description:     fmt.Sprintf("Topup of %.2f via %s", amount, source), // Corrected field name
		 TransactionDate: time.Now(),                 // Corrected field name
		 CreatedAt:       time.Now(),
	 }
	 err = s.pointTransactionRepo.Create(ctx, transaction)
	 if err != nil {
		 slog.Error("AllocatePointsForTopup: Failed to create point transaction", "error", err, "msisdn", msisdn)
		 return fmt.Errorf("failed to record point transaction for %s: %w", msisdn, err)
	 }

	 // 4. Update User's Total Points
	 err = s.userRepo.IncrementPoints(ctx, user.ID, points)
	 if err != nil {
		 slog.Error("AllocatePointsForTopup: Failed to update user points", "error", err, "userId", user.ID, "points", points)
		 // Attempt to rollback or mark transaction as failed? For now, just return error.
		 return fmt.Errorf("failed to update points for user %s: %w", msisdn, err)
	 }

	 slog.Info("Points allocated successfully", "msisdn", msisdn, "amount", amount, "points", points)
	 return nil
}

// GetDrawByDate retrieves a draw by its date
func (s *DrawServiceImpl) GetDrawByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	 // Implementation using s.drawRepo.FindByDate
	 draw, err := s.drawRepo.FindByDate(ctx, date)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) {
			 return nil, errors.New("no draw found for the specified date")
		 } else {
			 slog.Error("GetDrawByDate: Failed to find draw by date", "error", err, "date", date)
			 return nil, fmt.Errorf("error retrieving draw: %w", err)
		 }
	 }
	 return draw, nil
}

// GetDefaultDigitsForDay returns default eligible digits for a given weekday
// Note: Moved implementation to utils package, this just calls it.
func (s *DrawServiceImpl) GetDefaultDigitsForDay(day time.Weekday) ([]int, error) {
	 // No context needed as it's a pure function
	 return utils.GetDefaultEligibleDigits(day), nil // Return nil error
}

// --- Helper Functions (Internal to Draw Service) ---

// createWeightedPool creates a slice where each user is repeated based on their points
func createWeightedPool(users []*models.User) []*models.User {
	 totalWeight := 0
	 for _, user := range users {
		 // Ensure minimum 1 entry even if points are 0 or negative
		 weight := user.Points
		 if weight <= 0 {
			 weight = 1
		 }
		 totalWeight += weight
	 }

	 if totalWeight == 0 {
		 return []*models.User{} // Return empty pool if no users or all have <= 0 points
	 }

	 weightedPool := make([]*models.User, 0, totalWeight)
	 for _, user := range users {
		 weight := user.Points
		 if weight <= 0 {
			 weight = 1
		 }
		 for i := 0; i < weight; i++ {
			 weightedPool = append(weightedPool, user)
		 }
	 }
	 return weightedPool
}

// selectWeightedWinner selects a winner randomly from a weighted pool and returns the winner
// and the pool with the winner removed.
func selectWeightedWinner(weightedPool []*models.User) (*models.User, []*models.User, error) {
	 if len(weightedPool) == 0 {
		 return nil, weightedPool, errors.New("cannot select winner from empty pool")
	 }

	 // Seed random number generator (should ideally be done once at application start)
	 // rand.Seed(time.Now().UnixNano()) // Deprecated since Go 1.20
	 // Use crypto/rand for better randomness if needed, or default rand is okay for non-security critical selection.

	 winnerIndex := rand.Intn(len(weightedPool))
	 winner := weightedPool[winnerIndex]

	 // Create a new pool excluding all entries for the selected winner
	 remainingPool := make([]*models.User, 0, len(weightedPool))
	 for _, user := range weightedPool {
		 if user.MSISDN != winner.MSISDN {
			 remainingPool = append(remainingPool, user)
		 }
	 }

	 return winner, remainingPool, nil
}

// maskMsisdn masks the middle digits of an MSISDN for logging
func maskMsisdn(msisdn string) string {
	 if len(msisdn) < 7 { // Need enough digits to mask
		 return msisdn
	 }
	 // Example: Mask all but first 3 and last 4 digits
	 prefix := msisdn[:3]
	 suffix := msisdn[len(msisdn)-4:]
	 maskedPart := strings.Repeat("*", len(msisdn)-7)
	 return prefix + maskedPart + suffix
}





