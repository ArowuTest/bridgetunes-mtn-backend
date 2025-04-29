package services

import (
	"context"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawService defines the interface for draw-related operations
type DrawService interface {
	// GetDrawConfig retrieves the configuration for a draw based on the date
	GetDrawConfig(ctx context.Context, date time.Time) (*models.DrawConfigResponse, error)

	// GetPrizeStructure retrieves the prize structure for a given draw type (DAILY or WEEKLY)
	GetPrizeStructure(ctx context.Context, drawType string) ([]models.PrizeStructure, error)

	// UpdatePrizeStructure updates the prize structure for a given draw type
	UpdatePrizeStructure(ctx context.Context, drawType string, prizes []models.PrizeStructure) error

	// ScheduleDraw schedules a new draw
	ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int, useDefault bool) (*models.Draw, error)

	// ExecuteDraw executes a scheduled draw
	ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error)

	// GetDrawByID retrieves a draw by its ID
	GetDrawByID(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error)

	// GetWinnersByDrawID retrieves the winners for a specific draw
	GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error)

	// GetDraws retrieves a list of draws based on optional filters
	GetDraws(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error)

	// GetJackpotHistory retrieves the jackpot history within a date range
	GetJackpotHistory(ctx context.Context, startDate, endDate time.Time) ([]*models.JackpotHistoryEntry, error)
}

// --- Other Service Interfaces (Add as needed) ---

// UserService defines the interface for user-related operations
// type UserService interface { ... }

// NotificationService defines the interface for notification-related operations
// type NotificationService interface { ... }

