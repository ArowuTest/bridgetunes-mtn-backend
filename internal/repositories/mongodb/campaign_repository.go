package mongodb

import (
	"context"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CampaignRepository implements the repositories.CampaignRepository interface
type CampaignRepository struct {
	collection *mongo.Collection
}

// NewCampaignRepository creates a new CampaignRepository
func NewCampaignRepository(db *mongo.Database) repositories.CampaignRepository {
	return &CampaignRepository{
		collection: db.Collection("campaigns"),
	}
}

// FindByID finds a campaign by ID
func (r *CampaignRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Campaign, error) {
	var campaign models.Campaign

	 err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&campaign)
	 if err != nil {
		 return nil, err
	}

	 return &campaign, nil
}

// FindByStatus finds campaigns by status with pagination
// Note: Reverted return type back to []*models.Campaign to match interface
func (r *CampaignRepository) FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Campaign, error) {
	 opts := options.Find().
		 SetSkip(int64((page - 1) * limit)).
		 SetLimit(int64(limit)).
		 SetSort(bson.M{"scheduledAt": -1}) // Sort by scheduled date descending

	 cursor, err := r.collection.Find(ctx, bson.M{"status": status}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var campaigns []*models.Campaign // Reverted back to pointer slice
	 if err := cursor.All(ctx, &campaigns); err != nil {
		 return nil, err
	}

	 return campaigns, nil
}

// FindAll finds all campaigns with pagination
// Note: Reverted return type back to []*models.Campaign to match interface
func (r *CampaignRepository) FindAll(ctx context.Context, page, limit int) ([]*models.Campaign, error) {
	 opts := options.Find().
		 SetSkip(int64((page - 1) * limit)).
		 SetLimit(int64(limit)).
		 SetSort(bson.M{"scheduledAt": -1}) // Sort by scheduled date descending

	 cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var campaigns []*models.Campaign // Reverted back to pointer slice
	 if err := cursor.All(ctx, &campaigns); err != nil {
		 return nil, err
	}

	 return campaigns, nil
}

// Create creates a new campaign
func (r *CampaignRepository) Create(ctx context.Context, campaign *models.Campaign) error {
	 campaign.CreatedAt = time.Now()
	 campaign.UpdatedAt = time.Now()
	 _, err := r.collection.InsertOne(ctx, campaign)
	 return err
}

// Update updates a campaign
func (r *CampaignRepository) Update(ctx context.Context, campaign *models.Campaign) error {
	 campaign.UpdatedAt = time.Now()
	 _, err := r.collection.ReplaceOne(ctx, bson.M{"_id": campaign.ID}, campaign)
	 return err
}

// Delete deletes a campaign
func (r *CampaignRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	 _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	 return err
}

// Count counts all campaigns
func (r *CampaignRepository) Count(ctx context.Context) (int64, error) {
	 return r.collection.CountDocuments(ctx, bson.M{})
}

