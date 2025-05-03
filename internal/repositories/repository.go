package repositories

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	FindByMSISDN(ctx context.Context, msisdn string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context) ([]*models.User, error)
	FindByEligibleDigitsAndOptIn(ctx context.Context, digits []int, optInStatus bool, optInCutoff time.Time) ([]*models.User, error)
	FindUsersByRechargeWindow(ctx context.Context, startTime, endTime time.Time) ([]*models.User, error)
	FindEligibleConsolationUsers(ctx context.Context, digits []int, optInCutoff time.Time, rechargeStart, rechargeEnd time.Time) ([]*models.User, error)
	IncrementPoints(ctx context.Context, userID primitive.ObjectID, pointsToAdd int) error
}

// DrawRepository defines the interface for draw data operations
type DrawRepository interface {
	Create(ctx context.Context, draw *models.Draw) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error)
	FindByDate(ctx context.Context, date time.Time) (*models.Draw, error)
	Update(ctx context.Context, draw *models.Draw) error
	FindAll(ctx context.Context) ([]*models.Draw, error)
	FindNextScheduledDraw(ctx context.Context, afterTime time.Time) (*models.Draw, error)
	// Add other query methods as needed (e.g., FindByStatus)
}

// WinnerRepository defines the interface for winner data operations
type WinnerRepository interface {
	Create(ctx context.Context, winner *models.Winner) error
	CreateMany(ctx context.Context, winners []*models.Winner) error
	FindByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error)
	FindAll(ctx context.Context) ([]*models.Winner, error) // Added missing method
	// Add other query methods as needed (e.g., FindByUserID, FindByMSISDN)
}

// BlacklistRepository defines the interface for blacklist operations
type BlacklistRepository interface {
	Add(ctx context.Context, entry *models.BlacklistEntry) error
	Remove(ctx context.Context, msisdn string) error
	IsBlacklisted(ctx context.Context, msisdn string) (bool, error)
	FindAll(ctx context.Context) ([]*models.BlacklistEntry, error)
}

// SystemConfigRepository defines the interface for system configuration operations
type SystemConfigRepository interface {
	FindByKey(ctx context.Context, key string) (*models.SystemConfig, error)
	Create(ctx context.Context, config *models.SystemConfig) error
	Update(ctx context.Context, config *models.SystemConfig) error
	Delete(ctx context.Context, key string) error // Takes key (string)
	FindAll(ctx context.Context) ([]*models.SystemConfig, error)
	UpsertByKey(ctx context.Context, key string, value interface{}, description string) error
}

// PointTransactionRepository defines the interface for point transaction operations
type PointTransactionRepository interface {
	Create(ctx context.Context, transaction *models.PointTransaction) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.PointTransaction, error)
}

// JackpotRolloverRepository defines the interface for jackpot rollover operations
type JackpotRolloverRepository interface {
	Create(ctx context.Context, rollover *models.JackpotRollover) error
	FindBySourceDrawID(ctx context.Context, drawID primitive.ObjectID) (*models.JackpotRollover, error)
	FindRolloversByDestinationDate(ctx context.Context, date time.Time) ([]*models.JackpotRollover, error)
}

// --- Placeholder Interfaces for Undefined Repositories --- 
// These are needed to resolve build errors but need actual implementation later

// AdminUserRepository (Placeholder)
type AdminUserRepository interface {
	// Define methods used by admin_user_repository.go implementation
	// Example: FindByUsername(ctx context.Context, username string) (*models.AdminUser, error)
}

// CampaignRepository (Placeholder)
type CampaignRepository interface {
	// Define methods used by campaign_repository.go implementation
}

// NotificationRepository (Placeholder)
type NotificationRepository interface {
	// Define methods used by notification_repository.go implementation
}

// TemplateRepository (Placeholder)
type TemplateRepository interface {
	// Define methods used by template_repository.go implementation
}

// TopupRepository (Placeholder)
type TopupRepository interface {
	// Define methods used by topup_repository.go implementation
	// Example: Create(ctx context.Context, topup *models.Topup) error
}



