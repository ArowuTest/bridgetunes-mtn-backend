package mongodb

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BlacklistRepository implements the repositories.BlacklistRepository interface
type BlacklistRepository struct {
	collection *mongo.Collection
}

// NewBlacklistRepository creates a new BlacklistRepository
func NewBlacklistRepository(db *mongo.Database) repositories.BlacklistRepository {
	return &BlacklistRepository{
		collection: db.Collection("blacklists"),
	}
}

// FindByID finds a blacklist entry by ID
func (r *BlacklistRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Blacklist, error) {
	var blacklist models.Blacklist
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&blacklist)
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

// FindByMSISDN finds a blacklist entry by MSISDN
func (r *BlacklistRepository) FindByMSISDN(ctx context.Context, msisdn string) (*models.Blacklist, error) {
	var blacklist models.Blacklist
	err := r.collection.FindOne(ctx, bson.M{"msisdn": msisdn}).Decode(&blacklist)
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

// FindAll finds all blacklist entries with pagination
func (r *BlacklistRepository) FindAll(ctx context.Context, page, limit int) ([]*models.Blacklist, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"blacklistedAt": -1}) // Sort by blacklisted date descending

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blacklists []*models.Blacklist
	if err := cursor.All(ctx, &blacklists); err != nil {
		return nil, err
	}
	return blacklists, nil
}

// Create creates a new blacklist entry
func (r *BlacklistRepository) Create(ctx context.Context, blacklist *models.Blacklist) error {
	blacklist.CreatedAt = time.Now()
	blacklist.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, blacklist)
	return err
}

// Update updates a blacklist entry
func (r *BlacklistRepository) Update(ctx context.Context, blacklist *models.Blacklist) error {
	blacklist.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": blacklist.ID}, blacklist)
	return err
}

// Delete deletes a blacklist entry
func (r *BlacklistRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Count counts all blacklist entries
func (r *BlacklistRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
