package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
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
) *DrawServiceImpl {
	return &DrawServiceImpl{
		 drawRepo:             drawRepo,
		 userRepo:             userRepo,
		 winnerRepo:           winnerRepo,
		 blacklistRepo:        blacklistRepo,
		 systemConfigRepo:     systemConfigRepo,
		 pointTransactionRepo: pointTransactionRepo,
		 jackpotRolloverRepo:  jackpotRolloverRepo,
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
	 if err != nil && err != mongo.ErrNoDocuments {
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
	 baseJackpotKey := "base_jackpot_" + drawType
	 baseJackpotConfig, err := s.systemConfigRepo.FindByKey(ctx, baseJackpotKey)
	 if err != nil {
		 slog.Error("Failed to fetch base jackpot config", "error", err, "key", baseJackpotKey)
		 return nil, fmt.Errorf("failed to fetch base jackpot config ", baseJackpotKey, ": %w", err)
	 }
	 baseJackpotAmount, ok := baseJackpotConfig.Value.(float64) // Assuming stored as float64
	 if !ok {
		 slog.Error("Invalid base jackpot amount format in config", "key", baseJackpotKey, "valueType", fmt.Sprintf("%T", baseJackpotConfig.Value))
		 return nil, fmt.Errorf("invalid base jackpot amount format in config ", baseJackpotKey)
	 }

	 // 5. Calculate Rollover Amount *into* this draw
	 accumulatedRollover := 0.0
	 rollovers, err := s.jackpotRolloverRepo.FindRolloversByDestinationDate(ctx, drawDate)
	 if err != nil && err != mongo.ErrNoDocuments {
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
		 DrawType:                drawType,
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
		 return draw, fmt.Errorf("failed to mark draw as executing: %w", err)
	 }

	 // Defer status update on failure/completion
	 defer func() {
		 if r := recover(); r != nil {
			 draw.Status = models.DrawStatusFailed
			 draw.ErrorMessage = fmt.Sprintf("Panic during execution: %v", r)
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: PANIC: %v", time.Now().Format(time.RFC3339), r))
		 } else if err != nil {
			 draw.Status = models.DrawStatusFailed
			 draw.ErrorMessage = err.Error()
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: ERROR: %s", time.Now().Format(time.RFC3339), err.Error()))
		 } else {
			 draw.Status = models.DrawStatusCompleted
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("%s: Execution completed successfully", time.Now().Format(time.RFC3339)))
		 }
		 draw.ExecutionEndTime = time.Now()
		 updateErr := s.drawRepo.Update(ctx, draw)
		 if updateErr != nil {
			 slog.Error("ExecuteDraw: CRITICAL: Failed to update final draw status", "error", updateErr, "drawId", drawID, "finalStatusAttempt", draw.Status)
		 }
	 }()

	 // 3. Determine Eligibility Time Windows
	 eligibilityCutoff := time.Date(draw.DrawDate.Year(), draw.DrawDate.Month(), draw.DrawDate.Day(), 18, 0, 0, 0, draw.DrawDate.Location())
	 var eligibilityStart time.Time
	 if draw.DrawType == "SATURDAY" {
		 prevSaturday := draw.DrawDate.AddDate(0, 0, -7)
		 eligibilityStart = time.Date(prevSaturday.Year(), prevSaturday.Month(), prevSaturday.Day(), 18, 0, 1, 0, draw.DrawDate.Location())
	 } else {
		 // Daily draw eligibility starts at 00:00:00 on the draw day (Confirmed assumption)
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
		 winnerIndex := rand.Intn(len(poolA))
		 potentialJackpotWinner = poolA[winnerIndex]
		 draw.JackpotWinnerMsisdn = potentialJackpotWinner.MSISDN
		 draw.JackpotWinnerValidationStatus = models.JackpotValidationPending
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Potential Jackpot Winner Selected: %s", maskMsisdn(potentialJackpotWinner.MSISDN)))
	 } else {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Pool A is empty, cannot select Jackpot Winner.")
		 draw.JackpotWinnerValidationStatus = "NO_PARTICIPANTS"
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
			 if draw.DrawType == "SATURDAY" {
				 destinationDate = draw.DrawDate.AddDate(0, 0, 7)
			 } else {
				 daysUntilSaturday := time.Saturday - draw.DrawDate.Weekday()
				 if daysUntilSaturday <= 0 {
					 daysUntilSaturday += 7
				 }
				 destinationDate = draw.DrawDate.AddDate(0, 0, int(daysUntilSaturday))
			 }

			 rolloverRecord := &models.JackpotRollover{
				 SourceDrawID:        draw.ID,
				 SourceDrawDate:      draw.DrawDate,
				 RolloverAmount:      rolloverAmount,
				 DestinationDrawDate: destinationDate,
				 Reason:              models.JackpotValidationInvalidNotOptIn,
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
		 weightedPool := createWeightedPool(poolB)
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Created weighted pool for consolation winners (total weight: %d)", len(weightedPool)))

		 selectedConsolationMSISDNs := make(map[string]bool)
		 for _, prize := range draw.Prizes {
			 if prize.Category == models.JackpotCategory { // Skip jackpot prize here
				 continue
			 }

			 for i := 0; i < prize.NumWinners; i++ {
				 if len(weightedPool) == 0 {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Weighted pool exhausted while selecting for %s prize", prize.Category))
					 break // Stop selecting for this prize category if pool is empty
				 }

				 winnerUser, remainingPool, selectionErr := selectWeightedWinner(weightedPool)
				 weightedPool = remainingPool // Update pool for next selection

				 if selectionErr != nil {
					 slog.Error("Error selecting weighted winner", "error", selectionErr)
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("ERROR selecting weighted winner: %s", selectionErr.Error()))
					 continue // Skip this selection attempt
				 }

				 // Avoid selecting the same MSISDN multiple times in the same draw (REQFUNC039)
				 if selectedConsolationMSISDNs[winnerUser.MSISDN] {
					 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Skipping duplicate consolation winner selection: %s", maskMsisdn(winnerUser.MSISDN)))
					 i-- // Decrement i to retry selection for this slot
					 continue
				 }

				 selectedConsolationMSISDNs[winnerUser.MSISDN] = true
				 winner := &models.Winner{
					 DrawID:      draw.ID,
					 UserID:      winnerUser.ID,
					 MSISDN:      winnerUser.MSISDN,
					 PrizeCategory: prize.Category,
					 PrizeAmount: prize.Amount,
					 WinDate:     draw.DrawDate,
					 ClaimStatus: models.ClaimStatusPending, // Default status
				 }
				 consolationWinners = append(consolationWinners, winner)
				 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Selected %s Winner: %s (Points: %d)", prize.Category, maskMsisdn(winnerUser.MSISDN), winnerUser.Points))
			 }
		 }
	 } else {
		 draw.ExecutionLog = append(draw.ExecutionLog, "Pool B is empty, cannot select Consolation Winners.")
	 }

	 // 8. Create Winner Records (Jackpot if valid, and Consolation)
	 finalWinnersToCreate := consolationWinners
	 if isJackpotWinnerValid && potentialJackpotWinner != nil {
		 // Find the jackpot prize amount from the draw's prize list
		 var jackpotAmount float64
		 for _, p := range draw.Prizes {
			 if p.Category == models.JackpotCategory {
				 jackpotAmount = draw.CalculatedJackpotAmount // Use the final calculated amount
				 break
			 }
		 }
		 if jackpotAmount > 0 {
			 jackpotWinnerRecord := &models.Winner{
				 DrawID:      draw.ID,
				 UserID:      potentialJackpotWinner.ID,
				 MSISDN:      potentialJackpotWinner.MSISDN,
				 PrizeCategory: models.JackpotCategory,
				 PrizeAmount: jackpotAmount,
				 WinDate:     draw.DrawDate,
				 ClaimStatus: models.ClaimStatusPending,
			 }
			 finalWinnersToCreate = append(finalWinnersToCreate, jackpotWinnerRecord)
			 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Adding valid Jackpot Winner record for %s", maskMsisdn(potentialJackpotWinner.MSISDN)))
		 } else {
			 draw.ExecutionLog = append(draw.ExecutionLog, "Jackpot prize category not found or amount is zero, cannot create jackpot winner record.")
		 }
	 }

	 if len(finalWinnersToCreate) > 0 {
		 err = s.winnerRepo.CreateMany(ctx, finalWinnersToCreate)
		 if err != nil {
			 slog.Error("Failed to create winner records", "error", err, "drawId", drawID)
			 // This is a critical error, fail the draw execution
			 return draw, fmt.Errorf("failed to create winner records: %w", err)
		 }
		 draw.NumWinners = len(finalWinnersToCreate)
		 draw.ExecutionLog = append(draw.ExecutionLog, fmt.Sprintf("Successfully created %d winner records", len(finalWinnersToCreate)))
	 } else {
		 draw.NumWinners = 0
		 draw.ExecutionLog = append(draw.ExecutionLog, "No winners selected or created for this draw.")
	 }

	 // 9. Final status update is handled by the deferred function
	 return draw, nil // err will be nil here if execution reached this point
}

// --- Helper Functions for Draw Execution ---

// createWeightedPool creates a slice where each user appears N times based on their points.
func createWeightedPool(users []*models.User) []*models.User {
	 var weightedPool []*models.User
	 for _, user := range users {
		 // Ensure points are at least 1 for weighting (or handle 0 points case)
		 weight := user.Points
		 if weight <= 0 {
			 weight = 1 // Give at least one chance, or skip? Based on rules.
		 }
		 for i := 0; i < weight; i++ {
			 weightedPool = append(weightedPool, user)
		 }
	 }
	 return weightedPool
}

// selectWeightedWinner randomly selects a winner from the weighted pool and returns the winner
// and the pool with all instances of that winner removed.
func selectWeightedWinner(weightedPool []*models.User) (*models.User, []*models.User, error) {
	 if len(weightedPool) == 0 {
		 return nil, weightedPool, errors.New("weighted pool is empty")
	 }

	 winnerIndex := rand.Intn(len(weightedPool))
	 winner := weightedPool[winnerIndex]

	 // Remove all instances of the winner from the pool
	 var remainingPool []*models.User
	 for _, user := range weightedPool {
		 if user.ID != winner.ID { // Compare by unique ID
			 remainingPool = append(remainingPool, user)
		 }
	 }

	 return winner, remainingPool, nil
}

// maskMsisdn masks an MSISDN for logging (e.g., show first 3 and last 3 digits)
func maskMsisdn(msisdn string) string {
	 if len(msisdn) > 6 {
		 return msisdn[:3] + "******" + msisdn[len(msisdn)-3:]
	 }
	 return "******"
}

// --- Other Service Methods (Refined/Implemented) ---

// AllocatePoints allocates points based on top-up amount.
func (s *DrawServiceImpl) AllocatePoints(ctx context.Context, userID primitive.ObjectID, topupAmount float64) error {
	 // Rule: 1 point for every N100, 10 points for N1000+
	 pointsToAdd := 0
	 if topupAmount >= 1000 {
		 pointsToAdd = 10
	 } else if topupAmount >= 100 {
		 pointsToAdd = int(topupAmount / 100)
	 }

	 if pointsToAdd > 0 {
		 // Get user to log MSISDN
		 user, err := s.userRepo.FindByID(ctx, userID)
		 if err != nil {
			 slog.Error("AllocatePoints: Failed to find user", "error", err, "userId", userID)
			 return fmt.Errorf("user not found for point allocation: %w", err)
		 }

		 // Atomically increment points
		 err = s.userRepo.IncrementPoints(ctx, userID, pointsToAdd)
		 if err != nil {
			 slog.Error("AllocatePoints: Failed to increment points", "error", err, "userId", userID, "pointsToAdd", pointsToAdd)
			 return fmt.Errorf("failed to increment points: %w", err)
		 }

		 // Create a transaction record
		 transaction := &models.PointTransaction{
			 UserID:             userID,
			 MSISDN:             user.MSISDN, // Log MSISDN
			 Points:             pointsToAdd,
			 Source:             "TOPUP",
			 TransactionTimestamp: time.Now(),
			 RelatedTopupAmount: topupAmount, // Store the related topup amount
		 }
		 err = s.pointTransactionRepo.Create(ctx, transaction)
		 if err != nil {
			 slog.Error("AllocatePoints: Failed to create point transaction record", "error", err, "userId", userID)
			 // Log error but don't fail the allocation itself
		 }
		 slog.Info("Points allocated successfully", "userId", userID, "msisdn", maskMsisdn(user.MSISDN), "pointsAdded", pointsToAdd, "topupAmount", topupAmount)
	 } else {
		 slog.Info("No points allocated for topup amount", "userId", userID, "topupAmount", topupAmount)
	 }

	 return nil
}

// GetDrawByDate retrieves draw details for a specific date.
func (s *DrawServiceImpl) GetDrawByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	 draw, err := s.drawRepo.FindByDate(ctx, date)
	 if err != nil {
		 if err == mongo.ErrNoDocuments {
			 slog.Info("No draw found for date", "date", date.Format("2006-01-02"))
			 return nil, errors.New("no draw found for the specified date")
		 }
		 slog.Error("Failed to get draw by date", "error", err, "date", date.Format("2006-01-02"))
		 return nil, fmt.Errorf("failed to retrieve draw: %w", err)
	 }
	 return draw, nil
}

// GetWinnersByDrawID retrieves all winners for a specific draw.
func (s *DrawServiceImpl) GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	 winners, err := s.winnerRepo.FindByDrawID(ctx, drawID)
	 if err != nil {
		 slog.Error("Failed to get winners by draw ID", "error", err, "drawId", drawID)
		 return nil, fmt.Errorf("failed to retrieve winners: %w", err)
	 }
	 return winners, nil
}

// GetPrizeStructure retrieves the prize structure for a given draw type from config.
func (s *DrawServiceImpl) GetPrizeStructure(ctx context.Context, drawType string) ([]models.Prize, error) {
	 prizeConfigKey := "prize_structure_" + drawType
	 prizeConfig, err := s.systemConfigRepo.FindByKey(ctx, prizeConfigKey)
	 if err != nil {
		 slog.Error("Failed to fetch prize structure config", "error", err, "key", prizeConfigKey)
		 return nil, fmt.Errorf("failed to fetch prize structure config ", prizeConfigKey, ": %w", err)
	 }

	 // --- Configuration Parsing Logic --- 
	 // Assume Value is stored as []primitive.M (slice of BSON documents/maps)
	 prizeDataRaw, ok := prizeConfig.Value.(primitive.A) // BSON Array
	 if !ok {
		 slog.Error("Invalid prize structure format in config (expected BSON array)", "key", prizeConfigKey, "valueType", fmt.Sprintf("%T", prizeConfig.Value))
		 return nil, fmt.Errorf("invalid prize structure format in config ", prizeConfigKey)
	 }

	 var prizes []models.Prize
	 for _, item := range prizeDataRaw {
		 prizeMap, mapOk := item.(primitive.M) // BSON Document (Map)
		 if !mapOk {
			 slog.Warn("Skipping non-document item in prize structure array", "item", item)
			 continue
		 }

		 // Safely extract and convert fields
		 category, catOk := prizeMap["category"].(string)
		 amountRaw, amountOk := prizeMap["amount"]
		 countRaw, countOk := prizeMap["count"]

		 if !catOk || !amountOk || !countOk {
			 slog.Warn("Skipping prize item with missing or invalid fields", "item", prizeMap)
			 continue
		 }

		 // Convert amount (could be int32, int64, float64)
		 var amount float64
		 switch v := amountRaw.(type) {
		 case int32:
			 amount = float64(v)
		 case int64:
			 amount = float64(v)
		 case float64:
			 amount = v
		 default:
			 slog.Warn("Skipping prize item with invalid amount type", "item", prizeMap, "amountType", fmt.Sprintf("%T", amountRaw))
			 continue
		 }

		 // Convert count (could be int32, int64)
		 var count int
		 switch v := countRaw.(type) {
		 case int32:
			 count = int(v)
		 case int64:
			 count = int(v) // Potential overflow if count is huge, but unlikely for num winners
		 default:
			 slog.Warn("Skipping prize item with invalid count type", "item", prizeMap, "countType", fmt.Sprintf("%T", countRaw))
			 continue
		 }

		 prizes = append(prizes, models.Prize{
			 Category:   category,
			 Amount:     amount,
			 NumWinners: count,
		 })
	 }

	 if len(prizes) == 0 {
		 slog.Error("No valid prizes found after parsing config", "key", prizeConfigKey)
		 return nil, fmt.Errorf("no valid prizes configured for draw type ", drawType)
	 }

	 return prizes, nil
}

// UpdatePrizeStructure updates the prize structure in the system config.
func (s *DrawServiceImpl) UpdatePrizeStructure(ctx context.Context, drawType string, prizes []models.Prize) error {
	 prizeConfigKey := "prize_structure_" + drawType
	 // Convert []models.Prize back to a BSON-compatible format, e.g., []primitive.M
	 var prizeDataBSON primitive.A
	 for _, p := range prizes {
		 prizeDataBSON = append(prizeDataBSON, primitive.M{
			 "category": p.Category,
			 "amount":   p.Amount,   // Store as float64
			 "count":    int32(p.NumWinners), // Store as int32
		 })
	 }

	 description := fmt.Sprintf("Prize structure for %s draws", drawType)
	 err := s.systemConfigRepo.UpsertByKey(ctx, prizeConfigKey, prizeDataBSON, description)
	 if err != nil {
		 slog.Error("Failed to update prize structure config", "error", err, "key", prizeConfigKey)
		 return fmt.Errorf("failed to update prize structure config ", prizeConfigKey, ": %w", err)
	 }
	 slog.Info("Prize structure updated successfully", "key", prizeConfigKey)
	 return nil
}

// GetJackpotStatus retrieves the current status of the jackpot (e.g., current amount).
// This might involve fetching the latest Saturday draw or a specific config value.
func (s *DrawServiceImpl) GetJackpotStatus(ctx context.Context) (*models.JackpotStatus, error) {
	 // Option 1: Get from a dedicated config key
	 configKey := "current_jackpot_amount"
	 config, err := s.systemConfigRepo.FindByKey(ctx, configKey)
	 if err == nil {
		 if amount, ok := config.Value.(float64); ok {
			 return &models.JackpotStatus{CurrentAmount: amount, LastUpdatedAt: config.UpdatedAt}, nil
		 }
		 slog.Warn("Invalid format for current_jackpot_amount config", "valueType", fmt.Sprintf("%T", config.Value))
	 }
	 if err != nil && err != mongo.ErrNoDocuments {
		 slog.Error("Failed to fetch current_jackpot_amount config", "error", err)
		 // Fall through to Option 2
	 }

	 // Option 2: Calculate based on the *next* scheduled Saturday draw (if config not found/invalid)
	 // Find the next Saturday
	 today := time.Now()
	 daysUntilSaturday := time.Saturday - today.Weekday()
	 if daysUntilSaturday <= 0 {
		 daysUntilSaturday += 7
	 }
	 nextSaturdayDate := today.AddDate(0, 0, int(daysUntilSaturday))
	 nextSaturdayStart := time.Date(nextSaturdayDate.Year(), nextSaturdayDate.Month(), nextSaturdayDate.Day(), 0, 0, 0, 0, nextSaturdayDate.Location())

	 // Find the scheduled draw for next Saturday
	 nextDraw, err := s.drawRepo.FindByDate(ctx, nextSaturdayStart) // Assuming FindByDate matches the start of the day
	 if err == nil && nextDraw.Status == models.DrawStatusScheduled {
		 return &models.JackpotStatus{CurrentAmount: nextDraw.CalculatedJackpotAmount, LastUpdatedAt: nextDraw.UpdatedAt}, nil
	 }
	 if err != nil && err != mongo.ErrNoDocuments {
		 slog.Error("Failed to find next scheduled Saturday draw for jackpot status", "error", err, "date", nextSaturdayStart)
	 }

	 // Option 3: Fallback - Get base amount if no config or next draw found
	 baseJackpotKey := "base_jackpot_SATURDAY" // Assuming Saturday is the main jackpot
	 baseConfig, err := s.systemConfigRepo.FindByKey(ctx, baseJackpotKey)
	 if err == nil {
		 if amount, ok := baseConfig.Value.(float64); ok {
			 return &models.JackpotStatus{CurrentAmount: amount, LastUpdatedAt: baseConfig.UpdatedAt}, nil
		 }
	 }

	 slog.Error("Failed to determine current jackpot status through all methods")
	 return nil, errors.New("failed to determine current jackpot status")
}

// GetJackpotHistory retrieves the history of jackpot rollovers.
func (s *DrawServiceImpl) GetJackpotHistory(ctx context.Context, page, limit int) ([]*models.JackpotRollover, error) {
	 // This might need a dedicated method in the JackpotRolloverRepository with pagination and sorting
	 // For now, let's assume a simple FindAll exists or implement it here if needed.
	 // Placeholder: Fetching all for now, needs proper repo method.
	 // rollovers, err := s.jackpotRolloverRepo.FindAll(ctx, page, limit)
	 // if err != nil { ... }
	 // return rollovers, nil
	 slog.Warn("GetJackpotHistory: Pagination/Sorting not fully implemented, returning empty list.")
	 return []*models.JackpotRollover{}, nil // Placeholder
}

// GetDrawConfig (Placeholder - might involve fetching multiple config keys)
func (s *DrawServiceImpl) GetDrawConfig(ctx context.Context) (map[string]interface{}, error) {
	 slog.Warn("GetDrawConfig: Not implemented")
	 return map[string]interface{}{}, errors.New("GetDrawConfig not implemented")
}

