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
	FindByOptInStatus(ctx context.Context, optInStatus bool) ([]*models.User, error) // Added missing method
	FindByEligibleDigits(ctx context.Context, digits []int) ([]*models.User, error) // Added missing method
	Count(ctx context.Context) (int64, error) // Added missing method
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
	FindByDateRangeAndStatus(ctx context.Context, startDate, endDate time.Time, statuses []string) ([]*models.Draw, error) // Added missing method used in GetJackpotHistory
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

// EventRepository defines the interface for event data operations
type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context, page, limit int, status models.EventStatus, filter string) ([]*models.Event, error)
}

// TopupRepository defines the interface for topup data operations
type TopupRepository interface {
	Create(ctx context.Context, topup *models.Topup) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Topup, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Topup, error)
	FindByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Topup, error)
	Count(ctx context.Context) (int64, error)
}

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Notification, error)
	FindByCampaignID(ctx context.Context, campaignID primitive.ObjectID, page, limit int) ([]*models.Notification, error)
	FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Notification, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string, statusMessage string) error
	Count(ctx context.Context) (int64, error)
}

// TemplateRepository defines the interface for template data operations
type TemplateRepository interface {
	Create(ctx context.Context, template *models.Template) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error)
	FindByName(ctx context.Context, name string) (*models.Template, error)
	FindByType(ctx context.Context, templateType string, page, limit int) ([]*models.Template, error)
	FindAll(ctx context.Context, page, limit int) ([]*models.Template, error)
	Update(ctx context.Context, template *models.Template) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// CampaignRepository defines the interface for campaign data operations
type CampaignRepository interface {
	Create(ctx context.Context, campaign *models.Campaign) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Campaign, error)
	FindAll(ctx context.Context, page, limit int) ([]*models.Campaign, error)
	Update(ctx context.Context, campaign *models.Campaign) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// SystemSettingsRepository defines the interface for system settings operations
type SystemSettingsRepository interface {
	GetSettings(ctx context.Context) (*models.SystemSettings, error)
	UpdateSettings(ctx context.Context, settings *models.SystemSettings) error
	UpdateSMSGateway(ctx context.Context, gateway string, updatedBy string) error
}