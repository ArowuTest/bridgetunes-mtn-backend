package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"
)

// Compile-time check to ensure TopupServiceImpl implements TopupService
// Note: The TopupService interface itself is defined in service.go
var _ TopupService = (*TopupServiceImpl)(nil)

type TopupServiceImpl struct {
	userRepo             repositories.UserRepository
	pointTransactionRepo repositories.PointTransactionRepository
	drawService          DrawService // Inject DrawService for point allocation
}

func NewTopupService(userRepo repositories.UserRepository, pointTransactionRepo repositories.PointTransactionRepository, drawService DrawService) *TopupServiceImpl {
	return &TopupServiceImpl{
		userRepo:             userRepo,
		pointTransactionRepo: pointTransactionRepo,
		drawService:          drawService,
	}
}

// ProcessTopup processes a top-up event, finds the user, and allocates points.
func (s *TopupServiceImpl) ProcessTopup(ctx context.Context, msisdn string, amount float64, transactionTime time.Time) error {
	slog.Info("Processing top-up", "msisdn", msisdn, "amount", amount, "time", transactionTime)

	// 1. Find the user by MSISDN
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.Warn("User not found for top-up", "msisdn", msisdn)
			// Depending on requirements, we might create the user here or just ignore the top-up
			return fmt.Errorf("user with MSISDN %s not found", msisdn)
		}
		slog.Error("Failed to find user by MSISDN", "error", err, "msisdn", msisdn)
		return fmt.Errorf("failed to retrieve user: %w", err)
	}

	// 2. Allocate points using the DrawService's method
	// The DrawService now contains the canonical point calculation logic
	pointsToAdd, err := s.drawService.AllocatePointsForTopup(ctx, user.ID, amount, transactionTime)
	if err != nil {
		// Error is already logged within AllocatePointsForTopup
		return fmt.Errorf("failed to allocate points for top-up: %w", err)
	}

	slog.Info("Top-up processed successfully", "msisdn", msisdn, "amount", amount, "pointsAdded", pointsToAdd, "userId", user.ID)
	return nil
}

/*
// calculatePoints determines points based on top-up amount.
// THIS FUNCTION IS NOW DUPLICATED/REPLACED by the logic within DrawService.AllocatePointsForTopup
// It should be removed from here to avoid confusion and maintain a single source of truth.
func calculatePoints(amount float64) int {
    if amount >= 1000 {
        return 10 // 10 points for N1000 or more
    }
    // 1 point for every N100
    points := int(amount / 100)
    return points
}
*/


