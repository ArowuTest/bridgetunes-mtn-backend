package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SystemConfigRepository implements the repositories.SystemConfigRepository interface
type SystemConfigRepository struct {
	collection *mongo.Collection
}

// NewSystemConfigRepository creates a new SystemConfigRepository
func NewSystemConfigRepository(db *mongo.Database) repositories.SystemConfigRepository {
	return &SystemConfigRepository{
		collection: db.Collection("system_config"),
	}
}

// FindByID finds a system configuration by ID
func (r *SystemConfigRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.SystemConfig, error) {
	var config models.SystemConfig
	 err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&config)
	 if err != nil {
		 return nil, err
	}
	 return &config, nil
}

// FindByKey finds a system configuration by key.
// Note: The Value field is interface{}, so the caller needs to perform type assertion.
func (r *SystemConfigRepository) FindByKey(ctx context.Context, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	 err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&config)
	 if err != nil {
		 return nil, fmt.Errorf("failed to find system config by key ", key, ": %w", err)
	}
	 return &config, nil
}

// FindAll finds all system configurations with pagination
func (r *SystemConfigRepository) FindAll(ctx context.Context, page, limit int) ([]*models.SystemConfig, error) {
	 opts := options.Find()
	 if page > 0 && limit > 0 {
		 opts.SetSkip(int64((page - 1) * limit))
		 opts.SetLimit(int64(limit))
	 }
	 opts.SetSort(bson.M{"key": 1}) // Sort by key ascending

	 cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var configs []*models.SystemConfig
	 if err := cursor.All(ctx, &configs); err != nil {
		 return nil, err
	}
	 if configs == nil {
		 configs = []*models.SystemConfig{}
	 }
	 return configs, nil
}

// Create creates a new system configuration
func (r *SystemConfigRepository) Create(ctx context.Context, config *models.SystemConfig) error {
	 config.CreatedAt = time.Now()
	 config.UpdatedAt = time.Now()
	 _, err := r.collection.InsertOne(ctx, config)
	 return err
}

// Update updates a system configuration (using ReplaceOne)
func (r *SystemConfigRepository) Update(ctx context.Context, config *models.SystemConfig) error {
	 config.UpdatedAt = time.Now()
	 _, err := r.collection.ReplaceOne(ctx, bson.M{"_id": config.ID}, config)
	 return err
}

// UpsertByKey updates a system configuration by key, or creates it if it doesn't exist.
func (r *SystemConfigRepository) UpsertByKey(ctx context.Context, key string, value interface{}, description string) error {
	 filter := bson.M{"key": key}
	 update := bson.M{
		 "$set": bson.M{
			 "value":       value,
			 "description": description,
			 "updatedAt":   time.Now(),
		 },
		 "$setOnInsert": bson.M{
			 "key":       key,
			 "createdAt": time.Now(),
		 },
	 }
	 opts := options.Update().SetUpsert(true)

	 _, err := r.collection.UpdateOne(ctx, filter, update, opts)
	 if err != nil {
		 return fmt.Errorf("failed to upsert system config for key ", key, ": %w", err)
	 }
	 return nil
}

// Delete deletes a system configuration by ID (Changed from key to ID as per interface)
func (r *SystemConfigRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	 _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	 return err
}

// Count counts all system configurations
func (r *SystemConfigRepository) Count(ctx context.Context) (int64, error) {
	 return r.collection.CountDocuments(ctx, bson.M{})
}

