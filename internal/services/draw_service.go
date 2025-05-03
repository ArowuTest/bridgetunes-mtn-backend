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
					 slog.Error("Error selecting weighted consolation winne
(Content truncated due to size limit. Use line ranges to read in chunks)
