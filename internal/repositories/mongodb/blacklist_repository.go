package mongodb

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BlacklistRepository implements the repositories.BlacklistRepository interface
type BlacklistRepository struct {
	collection *mongo.Collection
}

// NewBlacklistRepository creates a new BlacklistRepository
func NewBlacklistRepository(db *mongo.Database) repositories.BlacklistRepository {
	return &BlacklistRepository{
		collection: db.Collection("blacklist"),
	}
}

// IsBlacklisted checks if an MSISDN exists in the blacklist collection.
func (r *BlacklistRepository) IsBlacklisted(ctx context.Context, msisdn string) (bool, error) {
	filter := bson.M{"msisdn": msisdn}
	count, err := r.collection.CountDocuments(ctx, filter)
	 if err != nil {
		 return false, err // Return error if query fails
	}
	 return count > 0, nil // Return true if count > 0, false otherwise
}

// Add adds an MSISDN to the blacklist
func (r *BlacklistRepository) Add(ctx context.Context, msisdn string, reason string) error {
	entry := models.BlacklistEntry{
		MSISDN:    msisdn,
		Reason:    reason,
		CreatedAt: time.Now(),
	}
	_, err := r.collection.InsertOne(ctx, entry)
	 return err
}

// Remove removes an MSISDN from the blacklist
func (r *BlacklistRepository) Remove(ctx context.Context, msisdn string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"msisdn": msisdn})
	 return err
}

// FindAll finds all blacklist entries (consider pagination for large lists)
func (r *BlacklistRepository) FindAll(ctx context.Context) ([]*models.BlacklistEntry, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var entries []*models.BlacklistEntry
	 if err := cursor.All(ctx, &entries); err != nil {
		 return nil, err
	}
	 if entries == nil {
		 entries = []*models.BlacklistEntry{}
	 }
	 return entries, nil
}


