package services

import (
	"context"
	"errors" // Added for placeholder errors
	"math/rand"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Compile-time check to ensure LegacyDrawService implements DrawService
var _ DrawService = (*LegacyDrawService)(nil)

// LegacyDrawService handles the original draw-related business logic
type LegacyDrawService struct {
	 drawRepo         repositories.DrawRepository
	 userRepo         repositories.UserRepository
	 winnerRepo       repositories.WinnerRepository
	 blacklistRepo    repositories.BlacklistRepository    // Added dependency
	 systemConfigRepo repositories.SystemConfigRepository // Added dependency
}

// NewLegacyDrawService creates a new LegacyDrawService
func NewLegacyDrawService(
	 drawRepo repositories.DrawRepository,
	 userRepo repositories.UserRepository,
	 winnerRepo repositories.WinnerRepository,
	 blacklistRepo repositories.BlacklistRepository,    // Added parameter
	 systemConfigRepo repositories.SystemConfigRepository, // Added parameter
) *LegacyDrawService {
	return &LegacyDrawService{
		 drawRepo:         drawRepo,
		 userRepo:         userRepo,
		 winnerRepo:       winnerRepo,
		 blacklistRepo:    blacklistRepo,    // Added assignment
		 systemConfigRepo: systemConfigRepo, // Added assignment
	}
}

// GetDrawByID retrieves a draw by ID
func (s *LegacyDrawService) GetDrawByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error) {
	return s.drawRepo.FindByID(ctx, id)
}

// GetDrawByDate retrieves a draw by date
func (s *LegacyDrawService) GetDrawByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	return s.drawRepo.FindByDate(ctx, date)
}

// GetDrawsByDateRange retrieves draws by date range with pagination
func (s *LegacyDrawService) GetDrawsByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Draw, error) {
	return s.drawRepo.FindByDateRange(ctx, start, end, page, limit)
}

// GetDrawsByStatus retrieves draws by status with pagination
func (s *LegacyDrawService) GetDrawsByStatus(ctx context.Context, status string, page, limit int) ([]*models.Draw, error) {
	return s.drawRepo.FindByStatus(ctx, status, page, limit)
}

// CreateDraw creates a new draw
func (s *LegacyDrawService) CreateDraw(ctx context.Context, draw *models.Draw) error {
	 draw.CreatedAt = time.Now()
	 draw.UpdatedAt = time.Now()
	 return s.drawRepo.Create(ctx, draw)
}

// UpdateDraw updates a draw
func (s *LegacyDrawService) UpdateDraw(ctx context.Context, draw *models.Draw) error {
	 draw.UpdatedAt = time.Now()
	 return s.drawRepo.Update(ctx, draw)
}

// DeleteDraw deletes a draw
func (s *LegacyDrawService) DeleteDraw(ctx context.Context, id primitive.ObjectID) error {
	 return s.drawRepo.Delete(ctx, id)
}

// ScheduleDraw schedules a new draw (implements DrawService interface method)
func (s *LegacyDrawService) ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int, useDefaultDigits bool) (*models.Draw, error) {
	 // Check if a draw already exists for this date
	 existingDraw, err := s.GetDrawByDate(ctx, drawDate)
	 if err == nil && existingDraw != nil {
		 return existingDraw, nil // Or return an error indicating it already exists?
	 }

	 // Determine eligible digits
	 finalEligibleDigits := eligibleDigits
	 if useDefaultDigits {
		 finalEligibleDigits = s.GetDefaultEligibleDigits(drawDate.Weekday())
	 }

	 // Create prizes based on draw type (Consider fetching from systemConfigRepo later)
	 var prizes []models.Prize
	 if drawType == "DAILY" {
		 prizes = []models.Prize{
			 {Category: "FIRST", Amount: 5000},
			 {Category: "SECOND", Amount: 3000},
			 {Category: "THIRD", Amount: 2000},
			 {Category: "FOURTH", Amount: 1000},
			 {Category: "FIFTH", Amount: 500},
		 }
	 } else if drawType == "WEEKLY" { // Assuming WEEKLY, adjust if needed
		 prizes = []models.Prize{
			 {Category: "FIRST", Amount: 100000},
			 {Category: "SECOND", Amount: 50000},
			 {Category: "THIRD", Amount: 30000},
			 {Category: "FOURTH", Amount: 20000},
			 {Category: "FIFTH", Amount: 10000},
			 {Category: "SIXTH", Amount: 5000},
			 {Category: "SEVENTH", Amount: 3000},
			 {Category: "EIGHTH", Amount: 2000},
			 {Category: "NINTH", Amount: 1000},
			 {Category: "TENTH", Amount: 500},
		 }
	 }

	 // Create the draw
	 draw := &models.Draw{
		 DrawDate:       drawDate,
		 DrawType:       drawType,
		 EligibleDigits: finalEligibleDigits,
		 Status:         "SCHEDULED",
		 Prizes:         prizes,
		 CreatedAt:      time.Now(),
		 UpdatedAt:      time.Now(),
	 }

	 err = s.drawRepo.Create(ctx, draw)
	 if err != nil {
		 return nil, err
	 }

	 return draw, nil
}

// ExecuteDraw executes a scheduled draw (implements DrawService interface method)
func (s *LegacyDrawService) ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error) {
	 // Get the draw
	 draw, err := s.drawRepo.FindByID(ctx, drawID)
	 if err != nil {
		 return nil, err
	 }

	 // Check if draw is already completed
	 if draw.Status == "COMPLETED" {
		 return draw, nil // Return the completed draw
	 }

	 // Get eligible users
	 eligibleUsers, err := s.userRepo.FindByEligibleDigits(ctx, draw.EligibleDigits, true)
	 if err != nil {
		 return draw, err // Return the draw and the error
	 }

	 // Filter out blacklisted users (Placeholder - implement actual check)
	 var filteredEligibleUsers []*models.User
	 for _, user := range eligibleUsers {
		 // Placeholder: Check if user.MSISDN is in blacklistRepo
		 // isBlacklisted, err := s.blacklistRepo.FindByMSISDN(ctx, user.MSISDN)
		 // if err != nil && err != mongo.ErrNoDocuments { /* handle error */ }
		 // if isBlacklisted == nil { // Not blacklisted
			 filteredEligibleUsers = append(filteredEligibleUsers, user)
		 // }
	 }

	 // Update total participants (using filtered list)
	 draw.TotalParticipants = len(filteredEligibleUsers)

	 // If no eligible users, mark draw as completed
	 if len(filteredEligibleUsers) == 0 {
		 draw.Status = "COMPLETED"
		 err = s.drawRepo.Update(ctx, draw)
		 return draw, err // Return updated draw and potential error
	 }

	 // Select winners from the filtered list
	 winners, err := s.selectWinners(ctx, draw, filteredEligibleUsers)
	 if err != nil {
		 return draw, err // Return the draw and the error
	 }

	 // Update draw status
	 draw.Status = "COMPLETED"
	 err = s.drawRepo.Update(ctx, draw)
	 if err != nil {
		 return draw, err // Return updated draw and potential error
	 }

	 // Create winner records
	 for _, winner := range winners {
		 err = s.winnerRepo.Create(ctx, winner)
		 if err != nil {
			 // Log or handle error, but maybe continue creating other winners?
			 // For now, return the draw and the first error encountered
			 return draw, err
		 }
	 }

	 return draw, nil // Return the completed draw
}

// selectWinners selects winners for a draw
func (s *LegacyDrawService) selectWinners(ctx context.Context, draw *models.Draw, eligibleUsers []*models.User) ([]*models.Winner, error) {
	 var winners []*models.Winner

	 // Shuffle eligible users
	 shuffleUsers(eligibleUsers)

	 // Select winners based on available prizes
	 for i, prize := range draw.Prizes {
		 if i < len(eligibleUsers) {
			 user := eligibleUsers[i]

			 // Create winner record
			 winner := &models.Winner{
				 MSISDN:        user.MSISDN,
				 DrawID:        draw.ID,
				 PrizeCategory: prize.Category,
				 PrizeAmount:   prize.Amount,
				 WinDate:       time.Now(),
				 ClaimStatus:   "PENDING",
				 CreatedAt:     time.Now(),
				 UpdatedAt:     time.Now(),
			 }

			 winners = append(winners, winner)

			 // Update prize with winner ID (This might need adjustment based on how Winner ID is generated)
			 // If winner.ID is only set after Create, this update needs to happen after winner creation loop
			 // draw.Prizes[i].WinnerID = winner.ID
		 }
	 }

	 return winners, nil
}

// shuffleUsers shuffles a slice of users
func shuffleUsers(users []*models.User) {
	 r := rand.New(rand.NewSource(time.Now().UnixNano()))
	 r.Shuffle(len(users), func(i, j int) {
		 users[i], users[j] = users[j], users[i]
	 })
}

// GetDrawCount gets the total number of draws
func (s *LegacyDrawService) GetDrawCount(ctx context.Context) (int64, error) {
	 return s.drawRepo.Count(ctx)
}

// GetDefaultEligibleDigits returns the default eligible digits for a given day of the week
func (s *LegacyDrawService) GetDefaultEligibleDigits(dayOfWeek time.Weekday) []int {
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
		 return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} // All digits for Saturday
	 default: // Sunday or other?
		 return []int{}
	 }
}

// --- Placeholder implementations for methods in DrawService interface but not in original LegacyDrawService ---

var errNotImplemented = errors.New("method not implemented")

func (s *LegacyDrawService) GetDrawConfig(ctx context.Context, date time.Time) (map[string]interface{}, error) {
	// Placeholder: Use systemConfigRepo to fetch config?
	// config, err := s.systemConfigRepo.FindByKey(ctx, "draw_config_" + date.Format("2006-01-02"))
	// if err != nil { return nil, err }
	// return config.Value.(map[string]interface{}), nil // Assuming Value is map[string]interface{}
	return make(map[string]interface{}), errNotImplemented
}

func (s *LegacyDrawService) GetPrizeStructure(ctx context.Context, drawType string) ([]models.PrizeStructure, error) {
	// Placeholder: Use systemConfigRepo to fetch prize structure?
	// config, err := s.systemConfigRepo.FindByKey(ctx, "prize_structure_" + drawType)
	// if err != nil { return nil, err }
	// return config.Value.([]models.PrizeStructure), nil // Assuming Value is []models.PrizeStructure
	return nil, errNotImplemented
}

func (s *LegacyDrawService) UpdatePrizeStructure(ctx context.Context, drawType string, structure []models.PrizeStructure) error {
	// Placeholder: Use systemConfigRepo to update prize structure?
	// config := &models.SystemConfig{ Key: "prize_structure_" + drawType, Value: structure }
	// return s.systemConfigRepo.Update(ctx, config)
	return errNotImplemented
}

func (s *LegacyDrawService) GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	// Placeholder: Use winnerRepo
	// return s.winnerRepo.FindByDrawID(ctx, drawID, 1, 1000) // Example with pagination
	return nil, errNotImplemented
}

func (s *LegacyDrawService) GetDraws(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error) {
	// Placeholder: Use existing GetDrawsByDateRange with default pagination?
	return s.GetDrawsByDateRange(ctx, startDate, endDate, 1, 100) // Example with pagination
}

func (s *LegacyDrawService) GetJackpotHistory(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Placeholder: This likely requires more complex logic involving multiple repos
	return nil, errNotImplemented
}

