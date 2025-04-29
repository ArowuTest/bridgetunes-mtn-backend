package services

import (
	"context"
	"math/rand" // Added import
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LegacyDrawService handles the original draw-related business logic
type LegacyDrawService struct { // Renamed from DrawService
	 drawRepo   repositories.DrawRepository
	 userRepo   repositories.UserRepository
	 winnerRepo repositories.WinnerRepository
}

// NewLegacyDrawService creates a new LegacyDrawService // Renamed from NewDrawService
func NewLegacyDrawService(drawRepo repositories.DrawRepository, userRepo repositories.UserRepository, winnerRepo repositories.WinnerRepository) *LegacyDrawService { // Renamed return type
	return &LegacyDrawService{ // Renamed struct type
		 drawRepo:   drawRepo,
		 userRepo:   userRepo,
		 winnerRepo: winnerRepo,
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

// ScheduleDraw schedules a new draw
func (s *LegacyDrawService) ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int) (*models.Draw, error) {
	 // Check if a draw already exists for this date
	 existingDraw, err := s.GetDrawByDate(ctx, drawDate)
	 if err == nil && existingDraw != nil {
		 return existingDraw, nil
	 }

	 // Create prizes based on draw type
	 var prizes []models.Prize
	 if drawType == "DAILY" {
		 prizes = []models.Prize{
			 {Category: "FIRST", Amount: 5000},
			 {Category: "SECOND", Amount: 3000},
			 {Category: "THIRD", Amount: 2000},
			 {Category: "FOURTH", Amount: 1000},
			 {Category: "FIFTH", Amount: 500},
		 }
	 } else if drawType == "WEEKLY" {
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
		 EligibleDigits: eligibleDigits,
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

// ExecuteDraw executes a scheduled draw
func (s *LegacyDrawService) ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) error {
	 // Get the draw
	 draw, err := s.drawRepo.FindByID(ctx, drawID)
	 if err != nil {
		 return err
	 }

	 // Check if draw is already completed
	 if draw.Status == "COMPLETED" {
		 return nil
	 }

	 // Get eligible users
	 eligibleUsers, err := s.userRepo.FindByEligibleDigits(ctx, draw.EligibleDigits, true)
	 if err != nil {
		 return err
	 }

	 // Update total participants
	 draw.TotalParticipants = len(eligibleUsers)

	 // If no eligible users, mark draw as completed
	 if len(eligibleUsers) == 0 {
		 draw.Status = "COMPLETED"
		 return s.drawRepo.Update(ctx, draw)
	 }

	 // Select winners
	 winners, err := s.selectWinners(ctx, draw, eligibleUsers)
	 if err != nil {
		 return err
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
	 }

	 return nil
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

			 // Update prize with winner ID
			 draw.Prizes[i].WinnerID = winner.ID
		 }
	 }

	 return winners, nil
}

// shuffleUsers shuffles a slice of users
func shuffleUsers(users []*models.User) {
	 // Use crypto/rand for better randomness if needed, but time-based is simpler for now
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

