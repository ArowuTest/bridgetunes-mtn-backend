package services

import (
	"context"
	"encoding/json" // Added for prize structure parsing
	"errors"
	"fmt"
	"math/rand"
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
					 break // Stop selecting for this prize category
				 }

				 var consolationWinner *models.User
				 var selectionErr error
				 attempts := 0
				 maxAttempts := len(weightedPoolB) * 2 // Safety break

				 for attempts < maxAttempts {
					 attempts++
					 consolationWinner, weightedPoolB, selectionErr = selectWeightedWinner(weightedPoolB)
					 if selectionErr != nil {
						 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR selecting weighted consolation winner for %s: %s", prize.Category, selectionErr.Error()))
						 // Decide if this is fatal for the prize category or the draw
						 // For now, log and break inner loop
						 consolationWinner = nil // Ensure we don't process a nil winner
						 break
					 }

					 // Check if already selected or is the jackpot winner
					 if !selectedConsolationMSISDNs[consolationWinner.MSISDN] {
						 selectedConsolationMSISDNs[consolationWinner.MSISDN] = true
						 break // Found a unique winner for this slot
					 } else {
						 // Winner already selected, try again (pool was already modified by selectWeightedWinner)
						 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Re-selected winner %s for %s, trying again...", maskMsisdn(consolationWinner.MSISDN), prize.Category))
						 consolationWinner = nil // Reset winner for next loop iteration check
						 if len(weightedPoolB) == 0 {
							 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Pool B exhausted during re-selection for %s prize", prize.Category))
							 break
						 }
					 }
				 }

				 if consolationWinner == nil { // If loop finished without finding a unique winner
					 if selectionErr != nil {
						 // Error already logged, break outer loop for this prize
						 break
					 } else if attempts >= maxAttempts {
						 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Max attempts reached trying to find unique winner for %s prize", prize.Category))
						 break // Stop selecting for this prize category
					 } else {
						 // Pool likely exhausted
						 break
					 }
				 }

				 // Create Winner record
				 winnerRecord := &models.Winner{
					 DrawID:       draw.ID,
					 UserID:       consolationWinner.ID,
					 MSISDN:       consolationWinner.MSISDN,
					 PrizeCategory: prize.Category,
					 PrizeAmount:  prize.Amount,
					 ClaimStatus:  models.ClaimStatusPending, // Use ClaimStatus
					 WinDate:      draw.DrawDate,
					 CreatedAt:    time.Now(),
					 UpdatedAt:    time.Now(),
				 }
				 consolationWinners = append(consolationWinners, winnerRecord)
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Consolation Winner Selected (%s): %s (Points: %d)", prize.Category, maskMsisdn(consolationWinner.MSISDN), consolationWinner.Points))
			 }
		 }
	 } else {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Pool B is empty, cannot select Consolation Winners.")
	 }

	 // 8. Save Winners (including valid Jackpot winner if applicable)
	 allWinnersToSave := consolationWinners
	 if isJackpotWinnerValid && potentialJackpotWinner != nil {
		 // Find the jackpot prize amount from the draw's prize list
		 jackpotPrizeAmount := 0.0
		 for _, p := range draw.Prizes {
			 if p.Category == models.JackpotCategory {
				 jackpotPrizeAmount = draw.CalculatedJackpotAmount // Use the calculated amount for the winner
				 break
			 }
		 }
		 if jackpotPrizeAmount == 0.0 {
			 // This shouldn't happen if scheduling is correct, but handle defensively
			 slog.Error("Jackpot prize category not found in draw prizes", "drawId", draw.ID)
			 // Decide how to handle: fail draw or just log? For now, log and continue without saving jackpot winner.
			 draw.ExecutionLog = append(draw.ExecutionLog, "ERROR: Jackpot prize category missing, cannot save jackpot winner record.")
		 } else {
			 jackpotWinnerRecord := &models.Winner{
				 DrawID:       draw.ID,
				 UserID:       potentialJackpotWinner.ID,
				 MSISDN:       potentialJackpotWinner.MSISDN,
				 PrizeCategory: models.JackpotCategory,
				 PrizeAmount:  jackpotPrizeAmount,
				 ClaimStatus:  models.ClaimStatusPending,
				 WinDate:      draw.DrawDate,
				 CreatedAt:    time.Now(),
				 UpdatedAt:    time.Now(),
			 }
			 allWinnersToSave = append(allWinnersToSave, jackpotWinnerRecord)
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Jackpot Winner Record Prepared: %s, Amount: %.2f", maskMsisdn(potentialJackpotWinner.MSISDN), jackpotPrizeAmount))
		 }
	 }

	 if len(allWinnersToSave) > 0 {
		 // Use CreateMany for efficiency
		 createWinnersErr := s.winnerRepo.CreateMany(ctx, allWinnersToSave)
		 if createWinnersErr != nil {
			 // This is a significant error, fail the draw
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR saving winner records: %s", createWinnersErr.Error()))
			 return draw, fmt.Errorf("failed to save winner records: %w", createWinnersErr)
		 }
		 draw.NumWinners = len(allWinnersToSave)
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Successfully saved %d winner records", len(allWinnersToSave)))
	 } else {
		 draw.ExecutionLog = append(draw.ExecutionLog, "No winners selected or eligible to be saved.")
	 }

	 // 9. Final status update is handled by the deferred function
	 return draw, nil // err will be nil here if execution reached the end without errors
}


// --- Helper & Getter Methods ---

// GetPrizeStructure retrieves the prize structure for a given draw type from system config
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
		 slog.Error("Invalid prize structure format in config (expected JSON string)", "key", configKey, "valueType", fmt.Sprintf("%T", config.Value))
		 return nil, fmt.Errorf("invalid prize structure format in config %s (expected JSON string)", configKey)
	 }

	 var prizes []models.Prize
	 err = json.Unmarshal([]byte(jsonString), &prizes)
	 if err != nil {
		 slog.Error("Failed to unmarshal prize structure JSON", "error", err, "key", configKey, "jsonString", jsonString)
		 return nil, fmt.Errorf("failed to parse prize structure JSON for %s: %w", configKey, err)
	 }

	 // Validate structure (optional but recommended)
	 if len(prizes) == 0 {
		 slog.Warn("Prize structure is empty", "key", configKey)
		 // Decide if this is an error or acceptable
	 }
	 // Add more validation if needed (e.g., check for Jackpot category)

	 return prizes, nil
}

// UpdatePrizeStructure updates the prize structure in system config
func (s *DrawServiceImpl) UpdatePrizeStructure(ctx context.Context, drawType string, structure []models.Prize) error {
	 configKey := "prize_structure_" + strings.ToUpper(drawType)

	 // Marshal the structure back to JSON string
	 jsonBytes, err := json.Marshal(structure)
	 if err != nil {
		 slog.Error("Failed to marshal prize structure to JSON", "error", err, "key", configKey)
		 return fmt.Errorf("failed to marshal prize structure for %s: %w", configKey, err)
	 }

	 // Upsert the JSON string into the config
	 err = s.systemConfigRepo.UpsertByKey(ctx, configKey, string(jsonBytes))
	 if err != nil {
		 slog.Error("Failed to upsert prize structure config", "error", err, "key", configKey)
		 return fmt.Errorf("failed to save prize structure config %s: %w", configKey, err)
	 }

	 slog.Info("Prize structure updated successfully", "key", configKey)
	 return nil
}

// GetWinnersByDrawID retrieves all winners for a specific draw
func (s *DrawServiceImpl) GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	 winners, err := s.winnerRepo.FindByDrawID(ctx, drawID)
	 if err != nil {
		 slog.Error("Failed to get winners by draw ID", "error", err, "drawId", drawID)
		 return nil, fmt.Errorf("failed to retrieve winners for draw %s: %w", drawID.Hex(), err)
	 }
	 return winners, nil
}

// GetDraws retrieves draws within a date range
func (s *DrawServiceImpl) GetDraws(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error) {
	 draws, err := s.drawRepo.FindByDateRange(ctx, startDate, endDate)
	 if err != nil {
		 slog.Error("Failed to get draws by date range", "error", err, "startDate", startDate, "endDate", endDate)
		 return nil, fmt.Errorf("failed to retrieve draws: %w", err)
	 }
	 return draws, nil
}

// GetDefaultDigitsForDay retrieves the default eligible digits for a given day of the week
func (s *DrawServiceImpl) GetDefaultDigitsForDay(ctx context.Context, day time.Weekday) ([]int, error) {
	 // This might fetch from config in the future, but for now uses the utility
	 // Adding context to match interface, though not used here currently.
	 digits := utils.GetDefaultEligibleDigits(day)
	 if len(digits) == 0 {
		 slog.Warn("No default digits defined for weekday", "day", day)
		 // Return empty slice, not an error, unless specifically required
	 }
	 return digits, nil
}

// GetDrawByDate retrieves a draw by its specific date
func (s *DrawServiceImpl) GetDrawByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	 // Normalize date to midnight UTC or appropriate timezone if needed
	 // date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	 draw, err := s.drawRepo.FindByDate(ctx, date)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) {
			 slog.Info("No draw found for date", "date", date)
			 return nil, fmt.Errorf("no draw found for date %s", date.Format("2006-01-02"))
		 }
		 slog.Error("Failed to get draw by date", "error", err, "date", date)
		 return nil, fmt.Errorf("failed to retrieve draw for date %s: %w", date.Format("2006-01-02"), err)
	 }
	 return draw, nil
}

// GetDrawByID retrieves a draw by its ID
func (s *DrawServiceImpl) GetDrawByID(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error) {
	 slog.Info("Fetching draw by ID", "drawId", drawID)
	 draw, err := s.drawRepo.FindByID(ctx, drawID)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) {
			 slog.Warn("Draw not found by ID", "drawId", drawID)
			 return nil, fmt.Errorf("draw with ID %s not found", drawID.Hex())
		 }
		 slog.Error("Failed to fetch draw by ID", "error", err, "drawId", drawID)
		 return nil, fmt.Errorf("failed to fetch draw by ID: %w", err)
	 }
	 return draw, nil
}


// GetJackpotStatus retrieves the current jackpot status
func (s *DrawServiceImpl) GetJackpotStatus(ctx context.Context) (*models.JackpotStatus, error) {
	 slog.Info("Fetching current jackpot status")
	 now := time.Now()
	 status := &models.JackpotStatus{
		 LastUpdatedAt: now,
	 }

	 // 1. Find the latest completed Saturday draw
	 latestSaturdayDraw, err := s.drawRepo.FindLatestDrawByTypeAndStatus(ctx, "SATURDAY", []string{string(models.DrawStatusCompleted)})
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to find latest completed Saturday draw", "error", err)
		 return nil, fmt.Errorf("failed to find latest completed Saturday draw: %w", err)
	 }

	 // 2. If a draw is found, get its details and winner
	 if latestSaturdayDraw != nil {
		 status.LastDrawDate = latestSaturdayDraw.DrawDate
		 // Find the jackpot winner for that draw
		 winners, findWinnerErr := s.winnerRepo.FindByDrawIDAndCategory(ctx, latestSaturdayDraw.ID, models.JackpotCategory)
		 if findWinnerErr != nil && !errors.Is(findWinnerErr, mongo.ErrNoDocuments) {
			 slog.Error("GetJackpotStatus: Failed to find jackpot winner for last draw", "error", findWinnerErr, "drawId", latestSaturdayDraw.ID)
			 // Continue, but status will lack winner info
		 } else if len(winners) > 0 {
			 // Assuming only one jackpot winner per draw
			 status.LastWinnerMSISDN = winners[0].MSISDN
			 status.LastWinAmount = winners[0].PrizeAmount
		 }
	 }

	 // 3. Find pending rollovers effective today
	 pendingRollovers, err := s.jackpotRolloverRepo.FindPendingRollovers(ctx, now)
	 accumulatedRollover := 0.0
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to find pending rollovers", "error", err)
		 // Continue, but current amount might be inaccurate
	 } else if err == nil {
		 for _, rollover := range pendingRollovers {
			 accumulatedRollover += rollover.RolloverAmount
		 }
	 }

	 // 4. Get base jackpot amount from config (assuming Saturday for current status)
	 baseJackpotKey := "base_jackpot_SATURDAY"
	 baseJackpotConfig, err := s.systemConfigRepo.FindByKey(ctx, baseJackpotKey)
	 baseJackpotAmount := 0.0
	 if err != nil {
		 slog.Error("GetJackpotStatus: Failed to fetch base jackpot config", "error", err, "key", baseJackpotKey)
		 // Continue, but current amount might be inaccurate
	 } else {
		 amount, ok := baseJackpotConfig.Value.(float64)
		 if !ok {
			 slog.Error("GetJackpotStatus: Invalid base jackpot amount format in config", "key", baseJackpotKey, "valueType", fmt.Sprintf("%T", baseJackpotConfig.Value))
			 // Continue, but current amount might be inaccurate
		 } else {
			 baseJackpotAmount = amount
		 }
	 }

	 // 5. Calculate current jackpot amount
	 status.CurrentAmount = baseJackpotAmount + accumulatedRollover

	 // 6. Find the next scheduled draw date (any type)
	 nextDraw, err := s.drawRepo.FindNextScheduledDraw(ctx, now)
	 if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		 slog.Error("GetJackpotStatus: Failed to find next scheduled draw", "error", err)
		 // Continue, but status will lack next draw date
	 } else if nextDraw != nil {
		 status.NextDrawDate = nextDraw.DrawDate
	 }

	 slog.Info("Successfully fetched jackpot status", "currentAmount", status.CurrentAmount, "lastDrawDate", status.LastDrawDate, "nextDrawDate", status.NextDrawDate)
	 return status, nil
}

// --- Utility functions specific to DrawService ---


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

// AllocatePointsForTopup calculates points for a top-up and updates user points and transaction log.
// This function now resides in DrawService as it's closely tied to draw eligibility/weighting.
func (s *DrawServiceImpl) AllocatePointsForTopup(ctx context.Context, userID primitive.ObjectID, amount float64, transactionTime time.Time) (int, error) {
	 // 1. Calculate points based on amount (centralized logic)
	 pointsToAdd := calculatePoints(amount)
	 if pointsToAdd <= 0 {
		 slog.Info("No points awarded for top-up amount", "userId", userID, "amount", amount)
		 return 0, nil // Not an error, just no points awarded
	 }

	 // 2. Fetch the user to get current points and MSISDN
	 user, err := s.userRepo.FindByID(ctx, userID)
	 if err != nil {
		 slog.Error("AllocatePointsForTopup: Failed to find user", "error", err, "userId", userID)
		 return 0, fmt.Errorf("failed to find user %s for point allocation: %w", userID.Hex(), err)
	 }

	 // 3. Create Point Transaction Record
	 transaction := &models.PointTransaction{
		 UserID:             user.ID,
		 MSISDN:             user.MSISDN, // Get MSISDN from user object
		 TopupAmount:        amount,
		 PointsAwarded:      pointsToAdd,
		 TransactionTimestamp: transactionTime,
		 CreatedAt:          time.Now(),
		 // Removed undefined fields: TransactionType, Description
	 }
	 err = s.pointTransactionRepo.Create(ctx, transaction)
	 if err != nil {
		 slog.Error("AllocatePointsForTopup: Failed to create point transaction record", "error", err, "userId", userID)
		 // Decide if this should prevent point update. For now, log and continue.
	 }

	 // 4. Update User's Points
	 err = s.userRepo.IncrementPoints(ctx, user.ID, pointsToAdd)
	 if err != nil {
		 slog.Error("AllocatePointsForTopup: Failed to increment user points", "error", err, "userId", userID, "pointsToAdd", pointsToAdd)
		 // Attempt to rollback transaction? Or just log? For now, log the inconsistency.
		 return 0, fmt.Errorf("failed to update user points for %s: %w", userID.Hex(), err)
	 }

	 slog.Info("Points allocated successfully", "userId", userID, "pointsAdded", pointsToAdd, "newTotalPoints", user.Points+pointsToAdd)
	 return pointsToAdd, nil
}

// calculatePoints determines points based on top-up amount.
// Centralized logic for point calculation.
func calculatePoints(amount float64) int {
	 if amount >= 1000 {
		 return 10 // 10 points for N1000 or more
	 }
	 // 1 point for every N100
	 points := int(amount / 100)
	 return points
}



