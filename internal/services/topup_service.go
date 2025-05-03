package services

import (
	"context"
	"math" // Import math package for Floor
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"github.com/ArowuTest/bridgetunes-mtn-backend/pkg/mtnapi"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Compile-time check to ensure LegacyTopupService implements TopupService
var _ TopupService = (*LegacyTopupService)(nil)

// LegacyTopupService handles topup-related business logic
type LegacyTopupService struct {
	 topupRepo repositories.TopupRepository
	 userService UserService // Use interface type for dependency
	 mtnClient *mtnapi.Client
}

// NewLegacyTopupService creates a new LegacyTopupService
func NewLegacyTopupService(topupRepo repositories.TopupRepository, userService UserService, mtnClient *mtnapi.Client) *LegacyTopupService {
	return &LegacyTopupService{
		 topupRepo: topupRepo,
		 userService: userService,
		 mtnClient: mtnClient,
	}
}

// GetTopupByID retrieves a topup by ID
func (s *LegacyTopupService) GetTopupByID(ctx context.Context, id primitive.ObjectID) (*models.Topup, error) {
	return s.topupRepo.FindByID(ctx, id)
}

// GetTopupsByMSISDN retrieves topups by MSISDN with pagination
func (s *LegacyTopupService) GetTopupsByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Topup, error) {
	return s.topupRepo.FindByMSISDN(ctx, msisdn, page, limit)
}

// GetTopupsByDateRange retrieves topups by date range with pagination
func (s *LegacyTopupService) GetTopupsByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Topup, error) {
	return s.topupRepo.FindByDateRange(ctx, start, end, page, limit)
}

// CreateTopup creates a new topup
func (s *LegacyTopupService) CreateTopup(ctx context.Context, topup *models.Topup) error {
	// Calculate points based on topup amount (proportional logic)
	points := calculatePoints(topup.Amount)
	 topup.PointsEarned = points

	// Create the topup record
	 err := s.topupRepo.Create(ctx, topup)
	 if err != nil {
		 return err
	 }

	// Add points to the user
	 return s.userService.AddPoints(ctx, topup.MSISDN, points)
}

// ProcessTopups processes topups from the MTN API
func (s *LegacyTopupService) ProcessTopups(ctx context.Context, startDate, endDate time.Time) (int, error) {
	// Get topups from MTN API
	 topups, err := s.mtnClient.GetTopups(startDate, endDate)
	 if err != nil {
		 return 0, err
	 }

	// Process each topup
	 processed := 0
	 for _, t := range topups {
		 // Check if topup already exists
		 // Consider optimizing this check if performance becomes an issue
		 existingTopups, err := s.topupRepo.FindByMSISDNAndRef(ctx, t.MSISDN, t.TransactionRef) // Assuming FindByMSISDNAndRef exists
		 if err != nil {
			 // Log error but continue processing other topups
			 continue
		 }
		 if len(existingTopups) > 0 {
			 continue // Skip if already exists
		 }

		 // Create new topup
		 topup := &models.Topup{
			 MSISDN:         t.MSISDN,
			 Amount:         t.Amount,
			 Channel:        "MTN",
			 Date:           t.Date,
			 TransactionRef: t.TransactionRef,
			 Processed:      true, // Mark as processed since we are creating it here
			 CreatedAt:      time.Now(),
			 UpdatedAt:      time.Now(),
		 }

		 err = s.CreateTopup(ctx, topup)
		 if err == nil {
			 processed++
		 } else {
			 // Log error creating topup
		 }
	 }

	 return processed, nil
}

// calculatePoints calculates points proportionally (1 point per 100 Naira)
func calculatePoints(amount float64) int {
	 if amount < 100 {
		 return 0
	 }
	 return int(math.Floor(amount / 100.0))
}

// GetTopupCount gets the total number of topups
func (s *LegacyTopupService) GetTopupCount(ctx context.Context) (int64, error) {
	return s.topupRepo.Count(ctx)
}


