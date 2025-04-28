package services

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Enhanced DrawService with jackpot rollover and two-pool selection
type DrawServiceEnhanced struct {
	drawRepo       repositories.DrawRepository
	userRepo       repositories.UserRepository
	winnerRepo     repositories.WinnerRepository
	configRepo     repositories.SystemConfigRepository
	topupRepo      repositories.TopupRepository
}

// NewDrawServiceEnhanced creates a new enhanced DrawService
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

// GetDrawConfig retrieves draw configuration for a specific date
func (s *DrawServiceEnhanced) GetDrawConfig(ctx context.Context, date time.Time) (map[string]interface{}, error) {
	// Determine if this is a weekday or Saturday draw
	isWeekly := date.Weekday() == time.Saturday
	drawType := "DAILY"
	if isWeekly {
		drawType = "WEEKLY"
	}

	// Get prize structure
	prizeStructure, err := s.GetPrizeStructure(ctx, drawType)
	if err != nil {
		return nil, err
	}

	// Get recommended digits for this day
	recommendedDigits := s.GetDefaultEligibleDigits(date.Weekday())

	// Calculate current jackpot amount (base + any rollovers)
	jackpotAmount, err := s.calculateJackpotAmount(ctx, date, drawType)
	if err != nil {
		return nil, err
	}

	// Return configuration
	return map[string]interface{}{
		"draw_date":          date.Format("2006-01-02"),
		"draw_type":          drawType,
		"prize_structure":    prizeStructure,
		"recommended_digits": recommendedDigits,
		"jackpot_amount":     jackpotAmount,
		"day_of_week":        date.Weekday().String(),
	}, nil
}

// GetPrizeStructure retrieves the prize structure for a draw type
func (s *DrawServiceEnhanced) GetPrizeStructure(ctx context.Context, drawType string) ([]models.PrizeStructure, error) {
	// Determine config key based on draw type
	configKey := "prizeStructureDaily"
	if drawType == "WEEKLY" {
		configKey = "prizeStructureWeekly"
	}

	// Get prize structure from config
	config, err := s.configRepo.FindByKey(ctx, configKey)
	if err != nil {
		// If not found, return default prize structure
		return s.getDefaultPrizeStructure(drawType), nil
	}

	// Convert to prize structure
	prizeStructure, ok := config.Value.([]models.PrizeStructure)
	if !ok {
		return s.getDefaultPrizeStructure(drawType), nil
	}

	return prizeStructure, nil
}

// UpdatePrizeStructure updates the prize structure for a draw type
func (s *DrawServiceEnhanced) UpdatePrizeStructure(ctx context.Context, drawType string, prizeStructure []models.PrizeStructure) error {
	// Validate prize structure
	if len(prizeStructure) == 0 {
		return errors.New("prize structure cannot be empty")
	}

	// Determine config key based on draw type
	configKey := "prizeStructureDaily"
	if drawType == "WEEKLY" {
		configKey = "prizeStructureWeekly"
	}

	// Update or create config
	config, err := s.configRepo.FindByKey(ctx, configKey)
	if err != nil {
		// Create new config
		config = &models.SystemConfig{
			Key:         configKey,
			Value:       prizeStructure,
			Description: drawType + " prize structure",
			UpdatedAt:   time.Now(),
		}
		return s.configRepo.Create(ctx, config)
	}

	// Update existing config
	config.Value = prizeStructure
	config.UpdatedAt = time.Now()
	return s.configRepo.Update(ctx, config)
}

// ScheduleDraw schedules a new draw with proper jackpot calculation
func (s *DrawServiceEnhanced) ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int) (*models.Draw, error) {
	// Check if a draw already exists for this date
	existingDraw, err := s.drawRepo.FindByDate(ctx, drawDate)
	if err == nil && existingDraw != nil {
		return existingDraw, nil
	}

	// Calculate jackpot amount
	jackpotAmount, err := s.calculateJackpotAmount(ctx, drawDate, drawType)
	if err != nil {
		return nil, err
	}

	// Get prize structure
	prizeStructure, err := s.GetPrizeStructure(ctx, drawType)
	if err != nil {
		return nil, err
	}

	// Create prizes array from prize structure
	var prizes []models.Prize
	for _, ps := range prizeStructure {
		// For categories with multiple prizes (e.g., consolation)
		count := ps.Count
		if count <= 0 {
			count = 1 // Default to 1 if count not specified
		}

		for i := 0; i < count; i++ {
			category := ps.Category
			if count > 1 {
				// Add number for multiple prizes of same category (e.g., CONSOLATION_1, CONSOLATION_2)
				category = category + "_" + string(rune('1'+i))
			}
			prizes = append(prizes, models.Prize{
				Category: category,
				Amount:   ps.Amount,
			})
		}
	}

	// Override jackpot amount with calculated value
	for i, prize := range prizes {
		if prize.Category == "JACKPOT" || prize.Category == "FIRST" {
			prizes[i].Amount = jackpotAmount
			break
		}
	}

	// Create the draw
	draw := &models.Draw{
		DrawDate:       drawDate,
		DrawType:       drawType,
		EligibleDigits: eligibleDigits,
		Status:         "SCHEDULED",
		JackpotAmount:  jackpotAmount,
		Prizes:         prizes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// If there are rollovers contributing to this jackpot, record them
	if drawType == "WEEKLY" {
		// Find weekday draws with invalid jackpot winners since last Saturday
		lastSaturday := getPreviousSaturday(drawDate)
		weekdayDraws, err := s.findWeekdayDrawsWithInvalidJackpotWinners(ctx, lastSaturday, drawDate)
		if err != nil {
			return nil, err
		}

		// Record rollovers
		for _, weekdayDraw := range weekdayDraws {
			draw.RolloverSource = append(draw.RolloverSource, models.RolloverInfo{
				SourceDrawID: weekdayDraw.ID,
				Amount:       1000000, // ₦1M from each weekday
				Reason:       "INVALID_WINNER",
			})
		}

		// Check if previous Saturday had an invalid jackpot winner
		prevSaturdayDraw, err := s.findPreviousSaturdayDrawWithInvalidJackpotWinner(ctx, lastSaturday)
		if err == nil && prevSaturdayDraw != nil {
			draw.RolloverSource = append(draw.RolloverSource, models.RolloverInfo{
				SourceDrawID: prevSaturdayDraw.ID,
				Amount:       prevSaturdayDraw.JackpotAmount,
				Reason:       "INVALID_WINNER",
			})
		}
	}

	err = s.drawRepo.Create(ctx, draw)
	if err != nil {
		return nil, err
	}

	return draw, nil
}

// ExecuteDraw executes a scheduled draw with two-pool selection and winner validation
func (s *DrawServiceEnhanced) ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) error {
	// Get the draw
	draw, err := s.drawRepo.FindByID(ctx, drawID)
	if err != nil {
		return err
	}

	// Check if draw is already completed
	if draw.Status == "COMPLETED" {
		return nil
	}

	// Update draw status to in progress
	draw.Status = "IN_PROGRESS"
	err = s.drawRepo.Update(ctx, draw)
	if err != nil {
		return err
	}

	// Get eligible users for jackpot (ALL users who recharged, regardless of opt-in)
	jackpotEligibleUsers, err := s.getEligibleUsersForJackpot(ctx, draw)
	if err != nil {
		return err
	}

	// Get eligible users for consolation prizes (ONLY opted-in users)
	consolationEligibleUsers, err := s.getEligibleUsersForConsolation(ctx, draw)
	if err != nil {
		return err
	}

	// Update participant counts
	draw.TotalParticipants = len(jackpotEligibleUsers)
	draw.OptedInParticipants = len(consolationEligibleUsers)

	// If no eligible users, mark draw as completed
	if len(jackpotEligibleUsers) == 0 {
		draw.Status = "COMPLETED"
		return s.drawRepo.Update(ctx, draw)
	}

	// Select winners using two-pool approach
	winners, err := s.selectWinnersWithTwoPools(ctx, draw, jackpotEligibleUsers, consolationEligibleUsers)
	if err != nil {
		return err
	}

	// Check if jackpot winner is valid (opted in)
	var jackpotRolloverNeeded bool
	var jackpotWinner *models.Winner

	for _, winner := range winners {
		if winner.PrizeCategory == "JACKPOT" || winner.PrizeCategory == "FIRST" {
			jackpotWinner = winner
			// Jackpot winner is valid only if opted in
			if !winner.IsOptedIn {
				winner.IsValid = false
				jackpotRolloverNeeded = true
			} else {
				winner.IsValid = true
			}
			break
		}
	}

	// Handle jackpot rollover if needed
	if jackpotRolloverNeeded {
		err = s.handleJackpotRollover(ctx, draw)
		if err != nil {
			return err
		}
	}

	// Update draw status
	draw.Status = "COMPLETED"
	err = s.drawRepo.Update(ctx, draw)
	if err != nil {
		return err
	}

	// Create winner records
	for _, winner := range winners {
		err = s.winnerRepo.Create(ctx, winner)
		if err != nil {
			return err
		}

		// Update prize with winner ID and validity
		for i, prize := range draw.Prizes {
			if prize.Category == winner.PrizeCategory {
				draw.Prizes[i].WinnerID = winner.ID
				isValid := winner.IsValid
				draw.Prizes[i].IsValid = &isValid
				break
			}
		}
	}

	// Final update to draw with winner IDs and validity
	return s.drawRepo.Update(ctx, draw)
}

// GetJackpotHistory retrieves jackpot history
func (s *DrawServiceEnhanced) GetJackpotHistory(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Get draws in date range
	draws, err := s.drawRepo.FindByDateRange(ctx, startDate, endDate, 1, 1000) // Large limit to get all
	if err != nil {
		return nil, err
	}

	var history []map[string]interface{}
	for _, draw := range draws {
		// Find jackpot prize
		var jackpotAmount float64
		var jackpotWinnerID primitive.ObjectID
		var jackpotValid *bool

		for _, prize := range draw.Prizes {
			if prize.Category == "JACKPOT" || prize.Category == "FIRST" {
				jackpotAmount = prize.Amount
				jackpotWinnerID = prize.WinnerID
				jackpotValid = prize.IsValid
				break
			}
		}

		// Get winner details if available
		var winnerMSISDN, maskedMSISDN string
		var isOptedIn bool
		if !jackpotWinnerID.IsZero() {
			winner, err := s.winnerRepo.FindByID(ctx, jackpotWinnerID)
			if err == nil && winner != nil {
				winnerMSISDN = winner.MSISDN
				maskedMSISDN = winner.MaskedMSISDN
				isOptedIn = winner.IsOptedIn
			}
		}

		// Add to history
		history = append(history, map[string]interface{}{
			"draw_date":      draw.DrawDate.Format("2006-01-02"),
			"draw_type":      draw.DrawType,
			"jackpot_amount": jackpotAmount,
			"winner_msisdn":  maskedMSISDN,
			"is_opted_in":    isOptedIn,
			"is_valid":       jackpotValid,
			"rollovers":      draw.RolloverSource,
			"rolled_over_to": draw.RolloverTarget,
		})
	}

	return history, nil
}

// Helper methods

// getDefaultPrizeStructure returns the default prize structure for a draw type
func (s *DrawServiceEnhanced) getDefaultPrizeStructure(drawType string) []models.PrizeStructure {
	if drawType == "WEEKLY" {
		return []models.PrizeStructure{
			{Category: "JACKPOT", Amount: 3000000, Count: 1},
			{Category: "SECOND", Amount: 1000000, Count: 1},
			{Category: "THIRD", Amount: 500000, Count: 1},
			{Category: "CONSOLATION", Amount: 100000, Count: 7},
		}
	}
	return []models.PrizeStructure{
		{Category: "JACKPOT", Amount: 1000000, Count: 1},
		{Category: "SECOND", Amount: 350000, Count: 1},
		{Category: "THIRD", Amount: 150000, Count: 1},
		{Category: "CONSOLATION", Amount: 75000, Count: 7},
	}
}

// calculateJackpotAmount calculates the jackpot amount for a draw
func (s *DrawServiceEnhanced) calculateJackpotAmount(ctx context.Context, drawDate time.Time, drawType string) (float64, error) {
	// Get base jackpot amount from prize structure
	prizeStructure, err := s.GetPrizeStructure(ctx, drawType)
	if err != nil {
		return 0, err
	}

	var baseJackpot float64
	for _, prize := range prizeStructure {
		if prize.Category == "JACKPOT" || prize.Category == "FIRST" {
			baseJackpot = prize.Amount
			break
		}
	}

	// For daily draws, just return the base amount
	if drawType == "DAILY" {
		return baseJackpot, nil
	}

	// For weekly draws, add rollovers
	totalJackpot := baseJackpot

	// Add weekday rollovers since last Saturday
	lastSaturday := getPreviousSaturday(drawDate)
	weekdayDraws, err := s.findWeekdayDrawsWithInvalidJackpotWinners(ctx, lastSaturday, drawDate)
	if err != nil {
		return baseJackpot, nil // Return base amount on error
	}

	// Add ₦1M for each weekday with invalid winner
	totalJackpot += float64(len(weekdayDraws)) * 1000000

	// Check if previous Saturday had an invalid jackpot winner
	prevSaturdayDraw, err := s.findPreviousSaturdayDrawWithInvalidJackpotWinner(ctx, lastSaturday)
	if err == nil && prevSaturdayDraw != nil {
		// Add entire previous Saturday jackpot
		totalJackpot += prevSaturdayDraw.JackpotAmount
	}

	return totalJackpot, nil
}

// getPreviousSaturday returns the previous Saturday date
func getPreviousSaturday(from time.Time) time.Time {
	daysToSaturday := (int(from.Weekday()) + 1) % 7
	if daysToSaturday == 0 {
		daysToSaturday = 7 // If today is Saturday, go back 7 days
	}
	return from.AddDate(0, 0, -daysToSaturday)
}

// findWeekdayDrawsWithInvalidJackpotWinners finds weekday draws with invalid jackpot winners
func (s *DrawServiceEnhanced) findWeekdayDrawsWithInvalidJackpotWinners(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error) {
	// Get completed draws in date range
	draws, err := s.drawRepo.FindByDateRange(ctx, startDate, endDate, 1, 100)
	if err != nil {
		return nil, err
	}

	var invalidJackpotDraws []*models.Draw
	for _, draw := range draws {
		// Skip Saturday draws
		if draw.DrawDate.Weekday() == time.Saturday {
			continue
		}

		// Skip draws that aren't completed
		if draw.Status != "COMPLETED" {
			continue
		}

		// Check if jackpot winner is invalid
		for _, prize := range draw.Prizes {
			if (prize.Category == "JACKPOT" || prize.Category == "FIRST") && prize.IsValid != nil && !*prize.IsValid {
				invalidJackpotDraws = append(invalidJackpotDraws, draw)
				break
			}
		}
	}

	return invalidJackpotDraws, nil
}

// findPreviousSaturdayDrawWithInvalidJackpotWinner finds the previous Saturday draw with invalid jackpot winner
func (s *DrawServiceEnhanced) findPreviousSaturdayDrawWithInvalidJackpotWinner(ctx context.Context, saturdayDate time.Time) (*models.Draw, error) {
	// Get draw for the specified Saturday
	draw, err := s.drawRepo.FindByDate(ctx, saturdayDate)
	if err != nil {
		return nil, err
	}

	// Check if jackpot winner is invalid
	for _, prize := range draw.Prizes {
		if (prize.Category == "JACKPOT" || prize.Category == "FIRST") && prize.IsValid != nil && !*prize.IsValid {
			return draw, nil
		}
	}

	return nil, errors.New("no invalid jackpot winner found")
}

// getEligibleUsersForJackpot gets eligible users for jackpot (ALL users who recharged)
func (s *DrawServiceEnhanced) getEligibleUsersForJackpot(ctx context.Context, draw *models.Draw) ([]*models.User, error) {
	// Get cutoff time (6 PM on draw date)
	cutoffTime := time.Date(
		draw.DrawDate.Year(),
		draw.DrawDate.Month(),
		draw.DrawDate.Day(),
		18, 0, 0, 0, // 6 PM
		draw.DrawDate.Location(),
	)

	// For Saturday draws, start from previous Saturday 6:01 PM
	var startTime time.Time
	if draw.DrawType == "WEEKLY" {
		prevSaturday := getPreviousSaturday(draw.DrawDate)
		startTime = time.Date(
			prevSaturday.Year(),
			prevSaturday.Month(),
			prevSaturday.Day(),
			18, 1, 0, 0, // 6:01 PM
			prevSaturday.Location(),
		)
	} else {
		// For daily draws, no specific start time, just check the last digit
		startTime = time.Time{}
	}

	// Get users who recharged before cutoff time
	users, err := s.userRepo.FindByRechargeTimeRange(ctx, startTime, cutoffTime)
	if err != nil {
		return nil, err
	}

	// Filter by eligible last digits for daily draws
	if draw.DrawType == "DAILY" && len(draw.EligibleDigits) > 0 {
		var eligibleUsers []*models.User
		for _, user := range users {
			// Check if last digit of MSISDN is eligible
			if len(user.MSISDN) > 0 {
				lastDigit := int(user.MSISDN[len(user.MSISDN)-1] - '0')
				for _, digit := range draw.EligibleDigits {
					if lastDigit == digit {
						eligibleUsers = append(eligibleUsers, user)
						break
					}
				}
			}
		}
		return eligibleUsers, nil
	}

	return users, nil
}

// getEligibleUsersForConsolation gets eligible users for consolation prizes (ONLY opted-in users)
func (s *DrawServiceEnhanced) getEligibleUsersForConsolation(ctx context.Context, draw *models.Draw) ([]*models.User, error) {
	// Get all eligible users for jackpot
	allEligibleUsers, err := s.getEligibleUsersForJackpot(ctx, draw)
	if err != nil {
		return nil, err
	}

	// Filter to only opted-in users
	var optedInUsers []*models.User
	for _, user := range allEligibleUsers {
		if user.OptInStatus {
			optedInUsers = append(optedInUsers, user)
		}
	}

	return optedInUsers, nil
}

// selectWinnersWithTwoPools selects winners using two-pool approach
func (s *DrawServiceEnhanced) selectWinnersWithTwoPools(
	ctx context.Context,
	draw *models.Draw,
	jackpotEligibleUsers []*models.User,
	consolationEligibleUsers []*models.User,
) ([]*models.Winner, error) {
	var winners []*models.Winner
	var selectedMSISDNs = make(map[string]bool) // Track selected MSISDNs to avoid duplicates

	// Sort prizes by category to ensure jackpot is first
	sort.Slice(draw.Prizes, func(i, j int) bool {
		// Order: JACKPOT/FIRST, SECOND, THIRD, CONSOLATION_*
		if draw.Prizes[i].Category == "JACKPOT" || draw.Prizes[i].Category == "FIRST" {
			return true
		}
		if draw.Prizes[j].Category == "JACKPOT" || draw.Prizes[j].Category == "FIRST" {
			return false
		}
		return draw.Prizes[i].Category < draw.Prizes[j].Category
	})

	// Select winners for each prize
	for _, prize := range draw.Prizes {
		var selectedUser *models.User
		var userPool []*models.User

		// Determine which pool to use
		if prize.Category == "JACKPOT" || prize.Category == "FIRST" {
			userPool = jackpotEligibleUsers
		} else {
			userPool = consolationEligibleUsers
		}

		// Skip if no eligible users
		if len(userPool) == 0 {
			continue
		}

		// Select winner using points-based weighting
		selectedUser = s.selectWinnerWithPointsWeighting(userPool, selectedMSISDNs)
		if selectedUser == nil {
			continue
		}

		// Mark this MSISDN as selected
		selectedMSISDNs[selectedUser.MSISDN] = true

		// Create winner record
		winner := &models.Winner{
			MSISDN:        selectedUser.MSISDN,
			MaskedMSISDN:  maskMSISDN(selectedUser.MSISDN), // Implement masking function
			DrawID:        draw.ID,
			PrizeCategory: prize.Category,
			PrizeAmount:   prize.Amount,
			IsOptedIn:     selectedUser.OptInStatus,
			IsValid:       selectedUser.OptInStatus || (prize.Category != "JACKPOT" && prize.Category != "FIRST"),
			Points:        selectedUser.Points,
			WinDate:       time.Now(),
			ClaimStatus:   "PENDING",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		winners = append(winners, winner)
	}

	return winners, nil
}

// selectWinnerWithPointsWeighting selects a winner using points-based weighting
func (s *DrawServiceEnhanced) selectWinnerWithPointsWeighting(users []*models.User, excludeMSISDNs map[string]bool) *models.User {
	if len(users) == 0 {
		return nil
	}

	// Create a pool of entries based on points
	var pool []string
	for _, user := range users {
		// Skip already selected MSISDNs
		if excludeMSISDNs[user.MSISDN] {
			continue
		}

		// Add user to pool multiple times based on points
		entries := user.Points
		if entries <= 0 {
			entries = 1 // Minimum 1 entry
		}

		for i := 0; i < entries; i++ {
			pool = append(pool, user.MSISDN)
		}
	}

	// If pool is empty after filtering, return nil
	if len(pool) == 0 {
		return nil
	}

	// Select random entry from pool
	rand.Seed(time.Now().UnixNano())
	selectedMSISDN := pool[rand.Intn(len(pool))]

	// Find and return the corresponding user
	for _, user := range users {
		if user.MSISDN == selectedMSISDN {
			return user
		}
	}

	return nil // Should never reach here
}

// handleJackpotRollover handles jackpot rollover when jackpot winner is invalid
func (s *DrawServiceEnhanced) handleJackpotRollover(ctx context.Context, draw *models.Draw) error {
	// Find target draw for rollover
	var targetDraw *models.Draw
	var err error

	if draw.DrawType == "DAILY" {
		// Weekday draws roll over to the next Saturday
		nextSaturday := getNextSaturday(draw.DrawDate)
		targetDraw, err = s.drawRepo.FindByDate(ctx, nextSaturday)
	} else {
		// Saturday draws roll over to the next Saturday
		nextSaturday := draw.DrawDate.AddDate(0, 0, 7)
		targetDraw, err = s.drawRepo.FindByDate(ctx, nextSaturday)
	}

	// If target draw doesn't exist, create it
	if err != nil || targetDraw == nil {
		// Determine target date
		var targetDate time.Time
		if draw.DrawType == "DAILY" {
			targetDate = getNextSaturday(draw.DrawDate)
		} else {
			targetDate = draw.DrawDate.AddDate(0, 0, 7)
		}

		// Schedule target draw
		targetDraw, err = s.ScheduleDraw(ctx, targetDate, "WEEKLY", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
		if err != nil {
			return err
		}
	}

	// Set rollover target
	draw.RolloverTarget = &targetDraw.ID

	// Add rollover source to target draw
	rolloverAmount := draw.JackpotAmount
	if draw.DrawType == "DAILY" {
		rolloverAmount = 1000000 // ₦1M for weekday draws
	}

	targetDraw.RolloverSource = append(targetDraw.RolloverSource, models.RolloverInfo{
		SourceDrawID: draw.ID,
		Amount:       rolloverAmount,
		Reason:       "INVALID_WINNER",
	})

	// Recalculate target draw jackpot amount
	targetDraw.JackpotAmount, err = s.calculateJackpotAmount(ctx, targetDraw.DrawDate, targetDraw.DrawType)
	if err != nil {
		return err
	}

	// Update jackpot amount in prizes
	for i, prize := range targetDraw.Prizes {
		if prize.Category == "JACKPOT" || prize.Category == "FIRST" {
			targetDraw.Prizes[i].Amount = targetDraw.JackpotAmount
			break
		}
	}

	// Update target draw
	return s.drawRepo.Update(ctx, targetDraw)
}

// getNextSaturday returns the next Saturday date
func getNextSaturday(from time.Time) time.Time {
	daysToSaturday := (7 - int(from.Weekday()) + int(time.Saturday)) % 7
	if daysToSaturday == 0 {
		daysToSaturday = 7 // If today is Saturday, go to next Saturday
	}
	return from.AddDate(0, 0, daysToSaturday)
}

// maskMSISDN masks an MSISDN showing only first 3 and last 2 digits
func maskMSISDN(msisdn string) string {
	if len(msisdn) <= 5 {
		return msisdn
	}

	first3 := msisdn[:3]
	last2 := msisdn[len(msisdn)-2:]
	masked := first3
	for i := 0; i < len(msisdn)-5; i++ {
		masked += "*"
	}
	masked += last2
	return masked
}

// GetDefaultEligibleDigits returns the default eligible digits for a given day of the week
func (s *DrawServiceEnhanced) GetDefaultEligibleDigits(dayOfWeek time.Weekday) []int {
	switch dayOfWeek {
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
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	default:
		return []int{}
	}
}
