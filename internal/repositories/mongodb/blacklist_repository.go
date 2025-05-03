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

// Compile-time check to ensure BlacklistRepository implements the interface
var _ repositories.BlacklistRepository = (*BlacklistRepository)(nil)

// BlacklistRepository handles MongoDB operations for BlacklistEntry
type BlacklistRepository struct {
	collection *mongo.Collection
}

// NewBlacklistRepository creates a new BlacklistRepository
func NewBlacklistRepository(db *mongo.Database) *BlacklistRepository {
	return &BlacklistRepository{
		collection: db.Collection("blacklist"),
	}
}

// Add inserts a new entry into the blacklist
// Corrected signature: takes *models.BlacklistEntry
func (r *BlacklistRepository) Add(ctx context.Context, entry *models.BlacklistEntry) error {
	entry.ID = primitive.NewObjectID()
	entry.CreatedAt = time.Now()
	// Use Upsert to prevent duplicate MSISDN entries, update if exists
	filter := bson.M{"msisdn": entry.MSISDN}
	update := bson.M{"$set": entry}
	 opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// Remove deletes an entry from the blacklist by MSISDN
func (r *BlacklistRepository) Remove(ctx context.Context, msisdn string) error {
	filter := bson.M{"msisdn": msisdn}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// IsBlacklisted checks if an MSISDN exists in the blacklist
func (r *BlacklistRepository) IsBlacklisted(ctx context.Context, msisdn string) (bool, error) {
	filter := bson.M{"msisdn": msisdn}
	count, err := r.collection.CountDocuments(ctx, filter)
	 if err != nil {
		 return false, err
	 }
	 return count > 0, nil
}

// FindAll retrieves all blacklist entries (consider pagination for production)
func (r *BlacklistRepository) FindAll(ctx context.Context) ([]*models.BlacklistEntry, error) {
	var entries []*models.BlacklistEntry
	cursor, err := r.collection.Find(ctx, bson.M{})
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &entries); err != nil {
		 return nil, err
	 }
	 if entries == nil {
		 entries = []*models.BlacklistEntry{}
	 }
	 return entries, nil
}
