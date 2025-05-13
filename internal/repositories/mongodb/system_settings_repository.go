package mongodb

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// SystemSettingsRepository implements repositories.SystemSettingsRepository
type SystemSettingsRepository struct {
	collection *mongo.Collection
}

// NewSystemSettingsRepository creates a new SystemSettingsRepository
func NewSystemSettingsRepository(db *mongo.Database) repositories.SystemSettingsRepository {
	return &SystemSettingsRepository{
		collection: db.Collection("system_settings"),
	}
}

// GetSettings retrieves the current system settings
func (r *SystemSettingsRepository) GetSettings(ctx context.Context) (*models.SystemSettings, error) {
	var settings models.SystemSettings
	err := r.collection.FindOne(ctx, bson.M{}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// If no settings exist, create default settings
		settings = models.SystemSettings{
			SMSGateway: "UDUX", // Default gateway
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		_, err = r.collection.InsertOne(ctx, settings)
		if err != nil {
			return nil, err
		}
		return &settings, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateSettings updates all system settings
func (r *SystemSettingsRepository) UpdateSettings(ctx context.Context, settings *models.SystemSettings) error {
	settings.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{}, settings)
	return err
}

// UpdateSMSGateway updates only the SMS gateway setting
func (r *SystemSettingsRepository) UpdateSMSGateway(ctx context.Context, gateway string, updatedBy string) error {
	update := bson.M{
		"$set": bson.M{
			"smsGateway": gateway,
			"updatedAt":  time.Now(),
			"updatedBy":  updatedBy,
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{}, update)
	return err
} 