package repositories

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Existing Interfaces (Updated based on redesign plan) ---

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	FindByMSISDN(ctx context.Context, msisdn string) (*models.User, error)
	FindAll(ctx context.Context) ([]*models.User, error) // Removed pagination for now, adjust if needed
	FindByOptInStatus(ctx context.Context, optInStatus bool) ([]*models.User, error) // Removed pagination
	// FindByEligibleDigits(ctx context.Context, digits []int, optInStatus bool) ([]*models.User, error) // Replaced by FindEligibleConsolationUsers
	// FindByRechargeTimeRange(ctx context.Context, start, end time.Time) ([]*models.User, error) // Replaced by FindUsersByRechargeWindow
	FindUsersByRechargeWindow(ctx context.Context, start, end time.Time) ([]*models.User, error) // Added for Pool A
	FindEligibleConsolationUsers(ctx context.Context, digits []int, optInCutoff time.Time, rechargeStart time.Time, rechargeEnd time.Time) ([]*models.User, error) // Added for Pool B
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	IncrementPoints(ctx context.Context, userID primitive.ObjectID, pointsToAdd int) error // Added for atomic update
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// TopupRepository defines the interface for topup data access
type TopupRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Topup, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Topup, error)
	FindByMSISDNAndRef(ctx context.Context, msisdn string, transactionRef string) ([]*models.Topup, error)
	FindByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Topup, error)
	// Add method to find topups within a window for a specific user?
	// FindByUserAndDateRange(ctx context.Context, userID primitive.ObjectID, start, end time.Time) ([]*models.Topup, error)
	Create(ctx context.Context, topup *models.Topup) error
	Update(ctx context.Context, topup *models.Topup) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// DrawRepository defines the interface for draw data access
type DrawRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error)
	FindByDate(ctx context.Context, date time.Time) (*models.Draw, error)
	FindByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Draw, error)
	FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Draw, error)
	// FindCompletedWithJackpot(ctx context.Context, limit int) ([]*models.Draw, error) // Keep or remove?
	// FindMostRecentCompletedBefore(ctx context.Context, date time.Time) (*models.Draw, error) // Keep or remove?
	FindNextScheduledDraw(ctx context.Context, afterDate time.Time, drawType string) (*models.Draw, error) // Added for rollover target
	Create(ctx context.Context, draw *models.Draw) error
	Update(ctx context.Context, draw *models.Draw) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// WinnerRepository defines the interface for winner data access
type WinnerRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Winner, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Winner, error)
	FindByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) // Removed pagination for simplicity
	FindByClaimStatus(ctx context.Context, status string, page, limit int) ([]*models.Winner, error)
	Create(ctx context.Context, winner *models.Winner) error
	CreateMany(ctx context.Context, winners []*models.Winner) error // Added for bulk creation
	Update(ctx context.Context, winner *models.Winner) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// NotificationRepository defines the interface for notification data access (Keep as is for now)
type NotificationRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Notification, error)
	FindByCampaignID(ctx context.Context, campaignID primitive.ObjectID, page, limit int) ([]*models.Notification, error)
	FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Notification, error)
	Create(ctx context.Context, notification *models.Notification) error
	Update(ctx context.Context, notification *models.Notification) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// TemplateRepository defines the interface for template data access (Keep as is for now)
type TemplateRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error)
	FindByName(ctx context.Context, name string) (*models.Template, error)
	FindByType(ctx context.Context, templateType string, page, limit int) ([]*models.Template, error)
	FindAll(ctx context.Context, page, limit int) ([]*models.Template, error)
	Create(ctx context.Context, template *models.Template) error
	Update(ctx context.Context, template *models.Template) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// CampaignRepository defines the interface for campaign data access (Keep as is for now)
type CampaignRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Campaign, error)
	FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Campaign, error)
	FindAll(ctx context.Context, page, limit int) ([]models.Campaign, error)
	Create(ctx context.Context, campaign *models.Campaign) error
	Update(ctx context.Context, campaign *models.Campaign) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// BlacklistRepository defines the interface for blacklist data access
type BlacklistRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Blacklist, error)
	FindByMSISDN(ctx context.Context, msisdn string) (*models.Blacklist, error)
	IsBlacklisted(ctx context.Context, msisdn string) (bool, error) // Added for simpler check
	FindAll(ctx context.Context, page, limit int) ([]*models.Blacklist, error)
	Create(ctx context.Context, blacklist *models.Blacklist) error
	Update(ctx context.Context, blacklist *models.Blacklist) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// AdminUserRepository defines the interface for admin user data access (Keep as is for now)
type AdminUserRepository interface {
	Create(ctx context.Context, adminUser *models.AdminUser) (*models.AdminUser, error)
	FindByEmail(ctx context.Context, email string) (*models.AdminUser, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.AdminUser, error)
}

// SystemConfigRepository defines the interface for system configuration data access
type SystemConfigRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.SystemConfig, error)
	FindByKey(ctx context.Context, key string) (*models.SystemConfig, error)
	FindAll(ctx context.Context, page, limit int) ([]*models.SystemConfig, error)
	Create(ctx context.Context, config *models.SystemConfig) error
	Update(ctx context.Context, config *models.SystemConfig) error
	UpsertByKey(ctx context.Context, key string, value interface{}, description string) error // Added for easier updates
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// --- New Interfaces from Redesign Plan ---

// PointTransactionRepository defines the interface for point transaction data access
type PointTransactionRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.PointTransaction, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]*models.PointTransaction, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.PointTransaction, error)
	Create(ctx context.Context, transaction *models.PointTransaction) error
	// Add other methods if needed (e.g., FindByDateRange)
}

// JackpotRolloverRepository defines the interface for jackpot rollover data access
type JackpotRolloverRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.JackpotRollover, error)
	FindBySourceDrawID(ctx context.Context, sourceDrawID primitive.ObjectID) (*models.JackpotRollover, error)
	FindRolloversByDestinationDate(ctx context.Context, destinationDate time.Time) ([]*models.JackpotRollover, error)
	Create(ctx context.Context, rollover *models.JackpotRollover) error
	// Add other methods if needed (e.g., Update if destination ID needs setting later)
}


