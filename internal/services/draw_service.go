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
					 // Corrected line:
					 slog.Error("Error selecting weighted consolation winner", "error", selectionErr)
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR selecting weighted consolation winner: %s", selectionErr.Error()))
					 continue // Skip this selection attempt
				 }

				 // Avoid selecting the same MSISDN multiple times in the same draw (REQFUNC039)
				 if selectedConsolationMSISDNs[winnerUser.MSISDN] {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Skipping duplicate consolation winner selection: %s", maskMsisdn(winnerUser.MSISDN)))
					 i-- // Decrement i to retry selection for this prize slot
					 continue
				 }
				 selectedConsolationMSISDNs[winnerUser.MSISDN] = true

				 // Create Winner record
				 winner := &models.Winner{
					 DrawID:       draw.ID,
					 UserID:       winnerUser.ID,
					 MSISDN:       winnerUser.MSISDN,
					 PrizeCategory: prize.Category,
					 PrizeAmount:  prize.Amount,
					 DrawDate:     draw.DrawDate,
					 ClaimStatus:  models.ClaimStatusPending,
					 CreatedAt:    time.Now(),
					 UpdatedAt:    time.Now(),
				 }
				 consolationWinners = append(consolationWinners, winner)
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Consolation Winner Selected (%s): %s (Points: %d)", prize.Category, maskMsisdn(winnerUser.MSISDN), winnerUser.Points))
			 }
		 }
	 } else {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Pool B is empty, cannot select Consolation Winners.")
	 }

	 // 8. Create Winner Records in DB
	 allWinners := consolationWinners
	 if isJackpotWinnerValid && potentialJackpotWinner != nil {
		 // Find the jackpot prize details
		 var jackpotPrize models.Prize
		 for _, p := range draw.Prizes {
			 if p.Category == models.JackpotCategory {
				 jackpotPrize = p
				 break
			 }
		 }
		 if jackpotPrize.Category == "" {
			 // This should ideally not happen if prize structure is fetched correctly
			 slog.Error("Jackpot prize category not found in draw prize structure", "drawId", draw.ID)
			 err = errors.New("jackpot prize category not found")
			 return draw, err
		 }

		 jackpotWinnerRecord := &models.Winner{
			 DrawID:       draw.ID,
			 UserID:       potentialJackpotWinner.ID,
			 MSISDN:       potentialJackpotWinner.MSISDN,
			 PrizeCategory: models.JackpotCategory,
			 PrizeAmount:  draw.CalculatedJackpotAmount, // Use the final calculated amount
			 DrawDate:     draw.DrawDate,
			 ClaimStatus:  models.ClaimStatusPending,
			 CreatedAt:    time.Now(),
			 UpdatedAt:    time.Now(),
		 }
		 allWinners = append(allWinners, jackpotWinnerRecord)
	 }

	 if len(allWinners) > 0 {
		 err = s.winnerRepo.CreateMany(ctx, allWinners)
		 if err != nil {
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR saving winners to DB: %s", err.Error()))
			 return draw, fmt.Errorf("failed to save winners: %w", err)
		 }
		 draw.NumWinners = len(allWinners)
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Successfully saved %d winners to DB", len(allWinners)))
	 } else {
		 draw.NumWinners = 0
		 draw.ExecutionLog = append(draw.ExecutionLog, "No winners selected or saved.")
	 }

	 // 9. Final Update (Handled by defer)
	 return draw, nil // err will be nil here if execution reached the end
}

// --- Helper & Other Methods ---

// GetPrizeStructure fetches and parses the prize structure from config
func (s *DrawServiceImpl) GetPrizeStructure(ctx context.Context, drawType string) ([]models.Prize, error) {
	 configKey := "prize_structure_" + strings.ToUpper(drawType)
	 config, err := s.systemConfigRepo.FindByKey(ctx, configKey)
	 if err != nil {
		 slog.Error("Failed to fetch prize structure config", "error", err, "key", configKey)
		 return nil, fmt.Errorf("failed to fetch prize structure config %s: %w", configKey, err)
	 }

	 // Value is likely stored as a stringified JSON array or similar
	 // We need to parse it into []models.Prize
	 var prizes []models.Prize
	 switch v := config.Value.(type) {
	 case string:
		 // Attempt to unmarshal if it's a JSON string
		 err = json.Unmarshal([]byte(v), &prizes)
		 if err != nil {
			 slog.Error("Failed to unmarshal prize structure JSON string", "error", err, "key", configKey, "value", v)
			 return nil, fmt.Errorf("invalid prize structure format (string) in config %s: %w", configKey, err)
		 }
	 case primitive.A: // Handle if stored directly as MongoDB array
		 // Convert primitive.A to []models.Prize (requires careful handling of types)
		 tempBytes, err := json.Marshal(v) // Marshal to JSON bytes first
		 if err != nil {
			 slog.Error("Failed to marshal prize structure BSON array", "error", err, "key", configKey)
			 return nil, fmt.Errorf("failed to marshal prize structure BSON array %s: %w", configKey, err)
		 }
		 err = json.Unmarshal(tempBytes, &prizes) // Unmarshal JSON bytes into struct slice
		 if err != nil {
			 slog.Error("Failed to unmarshal prize structure from BSON array", "error", err, "key", configKey)
			 return nil, fmt.Errorf("failed to unmarshal prize structure from BSON array %s: %w", configKey, err)
		 }
	 default:
		 slog.Error("Unsupported prize structure format in config", "key", configKey, "valueType", fmt.Sprintf("%T", config.Value))
		 return nil, fmt.Errorf("unsupported prize structure format in config %s", configKey)
	 }

	 if len(prizes) == 0 {
		 slog.Error("Prize structure config is empty or failed to parse", "key", configKey)
		 return nil, fmt.Errorf("prize structure config %s is empty or invalid", configKey)
	 }

	 return prizes, nil
}

// UpdatePrizeStructure updates the prize structure in the system config
func (s *DrawServiceImpl) UpdatePrizeStructure(ctx context.Context, drawType string, prizes []models.Prize) error {
	 configKey := "prize_structure_" + strings.ToUpper(drawType)
	 // Store as JSON string or directly if DB supports complex types well
	 // Storing as JSON string is often safer for cross-language/driver compatibility
	 prizeBytes, err := json.Marshal(prizes)
	 if err != nil {
		 slog.Error("Failed to marshal prize structure to JSON", "error", err)
		 return fmt.Errorf("failed to marshal prize structure: %w", err)
	 }

	 err = s.systemConfigRepo.UpsertByKey(ctx, configKey, string(prizeBytes), fmt.Sprintf("%s prize structure", strings.Title(strings.ToLower(drawType))))
	 if err != nil {
		 slog.Error("Failed to upsert prize structure config", "error", err, "key", configKey)
		 return fmt.Errorf("failed to save prize structure config %s: %w", configKey, err)
	 }
	 slog.Info("Prize structure updated successfully", "key", configKey)
	 return nil
}

// GetDrawByDate retrieves a draw by its date
func (s *DrawServiceImpl) GetDrawByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	 return s.drawRepo.FindByDate(ctx, date)
}

// GetDrawByID retrieves a draw by its ID
func (s *DrawServiceImpl) GetDrawByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error) {
	 return s.drawRepo.FindByID(ctx, id)
}

// GetAllDraws retrieves all draws (consider pagination)
func (s *DrawServiceImpl) GetAllDraws(ctx context.Context) ([]*models.Draw, error) {
	 return s.drawRepo.FindAll(ctx)
}

// GetWinnersByDrawID retrieves winners for a specific draw
func (s *DrawServiceImpl) GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	 return s.winnerRepo.FindByDrawID(ctx, drawID)
}

// GetJackpotStatus retrieves the current jackpot status (e.g., current amount)
// This might involve finding the latest draw or a dedicated status record
func (s *DrawServiceImpl) GetJackpotStatus(ctx context.Context) (*models.JackpotStatus, error) {
	 // Option 1: Find the latest completed/scheduled draw and use its jackpot amount
	 // Option 2: Query a dedicated JackpotStatus collection/document (more robust)

	 // Using Option 1 for now (simpler, but less accurate if draws are out of order)
	 // Find the most recent draw (regardless of status? or only completed/scheduled?)
	 // Need a FindLatestDraw method in the repository

	 // Placeholder implementation - returns a static status
	 // TODO: Implement proper logic based on chosen approach (latest draw or dedicated status)
	 status := &models.JackpotStatus{
		 CurrentAmount: 5000000.00, // Example static value
		 LastDrawDate:  time.Now().AddDate(0, 0, -1),
		 NextDrawDate:  time.Now().AddDate(0, 0, 1),
		 UpdatedAt:     time.Now(),
	 }
	 slog.Warn("GetJackpotStatus using placeholder implementation")
	 return status, nil
}

// AllocatePointsForTopup calculates and allocates points for a topup event
// This might be called by a separate process handling topup events
func (s *DrawServiceImpl) AllocatePointsForTopup(ctx context.Context, msisdn string, amount float64, topupTime time.Time) error {
	 user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) {
			 slog.Info("User not found for point allocation", "msisdn", maskMsisdn(msisdn))
			 return nil // Not an error if user doesn't exist
		 }
		 slog.Error("Failed to find user for point allocation", "error", err, "msisdn", maskMsisdn(msisdn))
		 return fmt.Errorf("failed to find user %s: %w", msisdn, err)
	 }

	 pointsToAdd := calculatePoints(amount)
	 if pointsToAdd <= 0 {
		 return nil // No points to add
	 }

	 // Atomically increment user points
	 err = s.userRepo.IncrementPoints(ctx, user.ID, pointsToAdd)
	 if err != nil {
		 slog.Error("Failed to increment user points", "error", err, "userId", user.ID, "pointsToAdd", pointsToAdd)
		 return fmt.Errorf("failed to increment points for user %s: %w", user.ID.Hex(), err)
	 }

	 // Create a point transaction record
	 transaction := &models.PointTransaction{
		 UserID:      user.ID,
		 MSISDN:      msisdn,
		 Points:      pointsToAdd,
		 Source:      "TOPUP",
		 Description: fmt.Sprintf("Points for N%.2f topup", amount),
		 Timestamp:   topupTime,
	 }
	 err = s.pointTransactionRepo.Create(ctx, transaction)
	 if err != nil {
		 slog.Error("Failed to create point transaction record", "error", err, "userId", user.ID)
		 // Log error but don't fail the overall operation
	 }

	 slog.Info("Points allocated successfully", "userId", user.ID, "msisdn", maskMsisdn(msisdn), "pointsAdded", pointsToAdd)
	 return nil
}

// --- Internal Helper Functions ---

// calculatePoints calculates points based on topup amount
func calculatePoints(amount float64) int {
	pointsToAdd := 0
	 if amount >= 1000 {
		 pointsToAdd = 10
	 } else {
		 pointsToAdd = int(amount / 100) // Integer division gives points per N100
	 }
	 return pointsToAdd
}

// createWeightedPool creates a slice where each user appears once per point they have
func createWeightedPool(users []*models.User) []*models.User {
	 totalWeight := 0
	 for _, u := range users {
		 // Ensure points are non-negative
		 if u.Points > 0 {
			 totalWeight += u.Points
		 }
	 }

	 if totalWeight == 0 {
		 // If no users have points, return the original pool (equal weighting)
		 slog.Warn("No users in the pool have points, falling back to equal weighting.")
		 return users
	 }

	 weightedPool := make([]*models.User, 0, totalWeight)
	 for _, u := range users {
		 if u.Points > 0 {
			 for i := 0; i < u.Points; i++ {
				 weightedPool = append(weightedPool, u)
			 }
		 }
	 }
	 return weightedPool
}

// selectWeightedWinner selects a winner randomly from the weighted pool
func selectWeightedWinner(weightedPool []*models.User) (winner *models.User, remainingPool []*models.User, err error) {
	 if len(weightedPool) == 0 {
		 return nil, weightedPool, errors.New("cannot select winner from empty pool")
	 }

	 // Seed the random number generator (important!)
	 // Using time.Now().UnixNano() is common but not perfectly random for rapid calls.
	 // Consider a shared rand.Source if performance is critical.
	 r := rand.New(rand.NewSource(time.Now().UnixNano()))

	 winnerIndex := r.Intn(len(weightedPool))
	 winner = weightedPool[winnerIndex]

	 // Create the remaining pool by removing *all* instances of the winner
	 remainingPool = make([]*models.User, 0, len(weightedPool)-winner.Points) // Approximate capacity
	 for _, user := range weightedPool {
		 if user.ID != winner.ID {
			 remainingPool = append(remainingPool, user)
		 }
	 }

	 return winner, remainingPool, nil
}

// maskMsisdn masks an MSISDN for logging
func maskMsisdn(msisdn string) string {
	 if len(msisdn) > 6 {
		 return msisdn[:3] + "..." + msisdn[len(msisdn)-3:]
	 }
	 return msisdn // Return as is if too short to mask meaningfully
}

// --- DrawService Interface Definition ---

// DrawService defines the interface for draw operations
type DrawService interface {
	ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int, useDefaultDigits bool) (*models.Draw, error)
	ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error)
	GetPrizeStructure(ctx context.Context, drawType string) ([]models.Prize, error)
	UpdatePrizeStructure(ctx context.Context, drawType string, prizes []models.Prize) error
	GetDrawByDate(ctx context.Context, date time.Time) (*models.Draw, error)
	GetDrawByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error)
	GetAllDraws(ctx context.Context) ([]*models.Draw, error)
	GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error)
	GetJackpotStatus(ctx context.Context) (*models.JackpotStatus, error)
	AllocatePointsForTopup(ctx context.Context, msisdn string, amount float64, topupTime time.Time) error
}
