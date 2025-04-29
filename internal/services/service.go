package services

import (
	"context"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawService defines the interface for draw-related operations
type DrawService interface {
	GetDrawConfig(ctx context.Context, date time.Time) (map[string]interface{}, error) // Use map[string]interface{}
	GetPrizeStructure(ctx context.Context, drawType string) ([]models.PrizeStructure, error)
	UpdatePrizeStructure(ctx context.Context, drawType string, structure []models.PrizeStructure) error
	ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int, useDefaultDigits bool) (*models.Draw, error)
	ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error)
	GetDrawByID(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error)
	GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error)
	GetDraws(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error)
	GetJackpotHistory(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) // Use []map[string]interface{}
}

// UserService defines the interface for user-related operations (Add other service interfaces as needed)
type UserService interface {
	// Define user service methods here
}

// TopupService defines the interface for topup-related operations
type TopupService interface {
	// Define topup service methods here
}

// NotificationService defines the interface for notification-related operations
type NotificationService interface {
	// Define notification service methods here
}
