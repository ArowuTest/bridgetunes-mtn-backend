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
	// New methods for redesign
	FindUsersByRechargeWindow(ctx context.Context, startTime, endTime time.Time) ([]*models.User, error)
	FindEligibleConsolationUsers(ctx context.Context, digits []int, optInCutoff, rechargeStart, rechargeEnd time.Time) ([]*models.User, error)
	IncrementPoints(ctx context.Context, userID primitive.ObjectID, points int) error
}

// DrawRepository defines the interface for draw data operations
type DrawRepository interface {
	Create(ctx context.Context, draw *models.Draw) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error)
	FindByDate(ctx context.Context, date time.Time) (*models.Draw, error)
	Update(ctx context.Context, draw *models.Draw) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context) ([]*models.Draw, error)
	// New/Updated methods for redesign
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error)
	FindByStatus(ctx context.Context, status string) ([]*models.Draw, error)
	FindNextScheduledDraw(ctx context.Context, currentDate time.Time) (*models.Draw, error)
	FindLatestDrawByTypeAndStatus(ctx context.Context, drawType string, statuses []string) (*models.Draw, error) // Added missing method used in GetJackpotStatus
}

// WinnerRepository defines the interface for winner data operations
type WinnerRepository interface {
	Create(ctx context.Context, winner *models.Winner) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Winner, error)
	FindByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Winner, error)
	Update(ctx context.Context, winner *models.Winner) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context) ([]*models.Winner, error)
	// New methods for redesign
	CreateMany(ctx context.Context, winners []*models.Winner) error
	FindByDrawIDAndCategory(ctx context.Context, drawID primitive.ObjectID, category string) ([]*models.Winner, error) // Added missing method used in GetJackpotStatus
}

// BlacklistRepository defines the interface for blacklist operations
type BlacklistRepository interface {
	IsBlacklisted(ctx context.Context, msisdn string) (bool, error)
	Add(ctx context.Context, msisdn string, reason string) error
	Remove(ctx context.Context, msisdn string) error
	FindAll(ctx context.Context) ([]*models.BlacklistEntry, error)
}

// SystemConfigRepository defines the interface for system configuration operations
type SystemConfigRepository interface {
	FindByKey(ctx context.Context, key string) (*models.SystemConfig, error)
	UpsertByKey(ctx context.Context, key string, value interface{}) error
	FindAll(ctx context.Context) ([]*models.SystemConfig, error)
}

// PointTransactionRepository defines the interface for point transaction operations
type PointTransactionRepository interface {
	Create(ctx context.Context, transaction *models.PointTransaction) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.PointTransaction, error)
}

// JackpotRolloverRepository defines the interface for jackpot rollover operations
type JackpotRolloverRepository interface {
	Create(ctx context.Context, rollover *models.JackpotRollover) error
	FindRolloversByDestinationDate(ctx context.Context, destinationDate time.Time) ([]*models.JackpotRollover, error)
	FindPendingRollovers(ctx context.Context, effectiveDate time.Time) ([]*models.JackpotRollover, error) // Added missing method used in GetJackpotStatus
}

// AdminUserRepository defines the interface for admin user data operations (Added)
type AdminUserRepository interface {
	Create(ctx context.Context, adminUser *models.AdminUser) error
	FindByEmail(ctx context.Context, email string) (*models.AdminUser, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.AdminUser, error)
	Update(ctx context.Context, adminUser *models.AdminUser) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context) ([]*models.AdminUser, error)
}



