package mongodb

import (
	"context"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// FindByKey finds a system configuration by key
func (r *SystemConfigRepository) FindByKey(ctx context.Context, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Create creates a new system configuration
func (r *SystemConfigRepository) Create(ctx context.Context, config *models.SystemConfig) error {
	_, err := r.collection.InsertOne(ctx, config)
	return err
}

// Update updates a system configuration
func (r *SystemConfigRepository) Update(ctx context.Context, config *models.SystemConfig) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"key": config.Key}, config)
	return err
}

// Delete deletes a system configuration
func (r *SystemConfigRepository) Delete(ctx context.Context, key string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"key": key})
	return err
}

// FindAll finds all system configurations
func (r *SystemConfigRepository) FindAll(ctx context.Context) ([]*models.SystemConfig, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []*models.SystemConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}
