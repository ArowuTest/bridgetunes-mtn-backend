package services

import (
	"context"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"github.com/bridgetunes/mtn-backend/pkg/mtnapi"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TopupService handles topup-related business logic
type TopupService struct {
	topupRepo repositories.TopupRepository
	userService *UserService
	mtnClient *mtnapi.Client
}

// NewTopupService creates a new TopupService
func NewTopupService(topupRepo repositories.TopupRepository, userService *UserService, mtnClient *mtnapi.Client) *TopupService {
	return &TopupService{
		topupRepo: topupRepo,
		userService: userService,
		mtnClient: mtnClient,
	}
}

// GetTopupByID retrieves a topup by ID
func (s *TopupService) GetTopupByID(ctx context.Context, id primitive.ObjectID) (*models.Topup, error) {
	return s.topupRepo.FindByID(ctx, id)
}

// GetTopupsByMSISDN retrieves topups by MSISDN with pagination
func (s *TopupService) GetTopupsByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Topup, error) {
	return s.topupRepo.FindByMSISDN(ctx, msisdn, page, limit)
}

// GetTopupsByDateRange retrieves topups by date range with pagination
func (s *TopupService) GetTopupsByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Topup, error) {
	return s.topupRepo.FindByDateRange(ctx, start, end, page, limit)
}

// CreateTopup creates a new topup
func (s *TopupService) CreateTopup(ctx context.Context, topup *models.Topup) error {
	// Calculate points based on topup amount
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
func (s *TopupService) ProcessTopups(ctx context.Context, startDate, endDate time.Time) (int, error) {
	// Get topups from MTN API
	topups, err := s.mtnClient.GetTopups(startDate, endDate)
	if err != nil {
		return 0, err
	}
	
	// Process each topup
	processed := 0
	for _, t := range topups {
		// Check if topup already exists
		existingTopups, err := s.topupRepo.FindByMSISDN(ctx, t.MSISDN, 1, 100)
		if err != nil {
			continue
		}
		
		// Check if this transaction reference already exists
		exists := false
		for _, existing := range existingTopups {
			if existing.TransactionRef == t.TransactionRef {
				exists = true
				break
			}
		}
		
		if !exists {
			// Create new topup
			topup := &models.Topup{
				MSISDN:         t.MSISDN,
				Amount:         t.Amount,
				Channel:        "MTN",
				Date:           t.Date,
				TransactionRef: t.TransactionRef,
				Processed:      false,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			
			err = s.CreateTopup(ctx, topup)
			if err == nil {
				processed++
			}
		}
	}
	
	return processed, nil
}

// calculatePoints calculates points based on topup amount
func calculatePoints(amount float64) int {
	switch {
	case amount >= 100 && amount < 200:
		return 1
	case amount >= 200 && amount < 300:
		return 2
	case amount >= 300 && amount < 400:
		return 3
	case amount >= 400 && amount < 500:
		return 4
	case amount >= 500 && amount < 1000:
		return 5
	case amount >= 1000:
		return 10
	default:
		return 0
	}
}

// GetTopupCount gets the total number of topups
func (s *TopupService) GetTopupCount(ctx context.Context) (int64, error) {
	return s.topupRepo.Count(ctx)
}
