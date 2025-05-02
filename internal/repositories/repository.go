package repositories

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- NOTE: This file assumes a generic Repository struct implements the interfaces below. ---
// --- You may need to adapt the GetAllCampaigns implementation to your specific structure. ---

// Repository struct (assuming a generic implementation)
type Repository struct {
	// Assuming db is your *mongo.Database instance
	// You might have separate repository structs for each entity
	// e.g., userRepo, topupRepo, etc.
	// Adjust the GetAllCampaigns method below accordingly.
	 db *mongo.Database
}

// NewRepository creates a new generic repository (example constructor)
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

// --- ADDED CODE START ---
// GetAllCampaigns retrieves all campaigns with pagination
// This assumes the generic Repository struct implements CampaignRepository
func (r *Repository) GetAllCampaigns(ctx context.Context, page, limit int) ([]models.Campaign, error) {
	var campaigns []models.Campaign
	collection := r.db.Collection("campaigns") // Assuming collection name is "campaigns"

	findOptions := options.Find()
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Sort by creation date descending

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	 if err != nil {
	 	return nil, err // Return error if Find fails
	 }
	defer cursor.Close(ctx)

	// Decode documents
	 err = cursor.All(ctx, &campaigns)
	 if err != nil {
	 	return nil, err // Return error if decoding fails
	 }

	// Ensure an empty slice is returned instead of nil if no campaigns found
	 if campaigns == nil {
	 	campaigns = []models.Campaign{}
	 }

	return campaigns, nil
}
// --- ADDED CODE END ---

// --- Existing Interfaces (copied from original for context) ---

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	FindByMSISDN(ctx context.Context, msisdn string) (*models.User, error)
	FindAll(ctx context.Context, page, limit int) ([]*models.User, error)
	FindByOptInStatus(ctx context.Context, optInStatus bool, page, limit int) ([]*models.User, error)
	FindByEligibleDigits(ctx context.Context, digits []int, optInStatus bool) ([]*models.User, error)
	FindByRechargeTimeRange(ctx context.Context, start, end time.Time) ([]*models.User, error) // Added this method
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// TopupRepository defines the interface for topup data access
type TopupRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Topup, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Topup, error)
	FindByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Topup, error)
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
	FindCompletedWithJackpot(ctx context.Context, limit int) ([]*models.Draw, error) // Added based on service usage
	FindMostRecentCompletedBefore(ctx context.Context, date time.Time) (*models.Draw, error) // Added based on service usage
	Create(ctx context.Context, draw *models.Draw) error
	Update(ctx context.Context, draw *models.Draw) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// WinnerRepository defines the interface for winner data access
type WinnerRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Winner, error)
	FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Winner, error)
	FindByDrawID(ctx context.Context, drawID primitive.ObjectID, page, limit int) ([]*models.Winner, error)
	FindByClaimStatus(ctx context.Context, status string, page, limit int) ([]*models.Winner, error)
	Create(ctx context.Context, winner *models.Winner) error
	Update(ctx context.Context, winner *models.Winner) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// NotificationRepository defines the interface for notification data access
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

// TemplateRepository defines the interface for template data access
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

// CampaignRepository defines the interface for campaign data access
type CampaignRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Campaign, error)
	FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Campaign, error)
	FindAll(ctx context.Context, page, limit int) ([]models.Campaign, error) // Changed return type to match implementation
	Create(ctx context.Context, campaign *models.Campaign) error
	Update(ctx context.Context, campaign *models.Campaign) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// BlacklistRepository defines the interface for blacklist data access
type BlacklistRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Blacklist, error)
	FindByMSISDN(ctx context.Context, msisdn string) (*models.Blacklist, error)
	FindAll(ctx context.Context, page, limit int) ([]*models.Blacklist, error)
	Create(ctx context.Context, blacklist *models.Blacklist) error
	Update(ctx context.Context, blacklist *models.Blacklist) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}

// --- NOTE: You might need to add implementations for other methods in the generic Repository ---
// --- or ensure your specific repository implementations match these interfaces. ---




// AdminUserRepository defines the interface for admin user data access
type AdminUserRepository interface {
	Create(ctx context.Context, adminUser *models.AdminUser) (*models.AdminUser, error)
	FindByEmail(ctx context.Context, email string) (*models.AdminUser, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.AdminUser, error)
	// Add other methods as needed (e.g., Update, Delete, FindAll)
}


