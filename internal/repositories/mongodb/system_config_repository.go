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

// Compile-time check to ensure SystemConfigRepository implements the interface
var _ repositories.SystemConfigRepository = (*SystemConfigRepository)(nil)

// SystemConfigRepository handles MongoDB operations for SystemConfig
type SystemConfigRepository struct {
	collection *mongo.Collection
}

// NewSystemConfigRepository creates a new SystemConfigRepository
func NewSystemConfigRepository(db *mongo.Database) *SystemConfigRepository {
	return &SystemConfigRepository{
		collection: db.Collection("system_configs"),
	}
}

// FindByKey finds a system config by its key
func (r *SystemConfigRepository) FindByKey(ctx context.Context, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	filter := bson.M{"key": key}
	 err := r.collection.FindOne(ctx, filter).Decode(&config)
	 if err != nil {
		 return nil, err // Includes mongo.ErrNoDocuments if not found
	 }
	 return &config, nil
}

// Create inserts a new system config
func (r *SystemConfigRepository) Create(ctx context.Context, config *models.SystemConfig) error {
	config.ID = primitive.NewObjectID()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, config)
	return err
}

// Update updates an existing system config
func (r *SystemConfigRepository) Update(ctx context.Context, config *models.SystemConfig) error {
	config.UpdatedAt = time.Now()
	filter := bson.M{"_id": config.ID}
	update := bson.M{"$set": config}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a system config by its key
// Corrected signature: takes key (string) not ID (primitive.ObjectID)
func (r *SystemConfigRepository) Delete(ctx context.Context, key string) error {
	filter := bson.M{"key": key}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindAll retrieves all system configs
func (r *SystemConfigRepository) FindAll(ctx context.Context) ([]*models.SystemConfig, error) {
	var configs []*models.SystemConfig
	cursor, err := r.collection.Find(ctx, bson.M{})
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &configs); err != nil {
		 return nil, err
	 }
	 // Return empty slice instead of nil
	 if configs == nil {
		 configs = []*models.SystemConfig{}
	 }
	 return configs, nil
}

// UpsertByKey creates or updates a system config by key
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
	return err
}

