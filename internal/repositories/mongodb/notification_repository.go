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

// NotificationRepository implements the repositories.NotificationRepository interface
type NotificationRepository struct {
	collection *mongo.Collection
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *mongo.Database) repositories.NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

// FindByID finds a notification by ID
func (r *NotificationRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error) {
	var notification models.Notification
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&notification)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// FindByMSISDN finds notifications by MSISDN with pagination
func (r *NotificationRepository) FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Notification, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"sentDate": -1}) // Sort by sent date descending

	cursor, err := r.collection.Find(ctx, bson.M{"msisdn": msisdn}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

// FindByCampaignID finds notifications by campaign ID with pagination
func (r *NotificationRepository) FindByCampaignID(ctx context.Context, campaignID primitive.ObjectID, page, limit int) ([]*models.Notification, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"sentDate": -1}) // Sort by sent date descending

	cursor, err := r.collection.Find(ctx, bson.M{"campaignId": campaignID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

// FindByStatus finds notifications by status with pagination
func (r *NotificationRepository) FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Notification, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"sentDate": -1}) // Sort by sent date descending

	cursor, err := r.collection.Find(ctx, bson.M{"status": status}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, notification)
	return err
}

// Update updates a notification
func (r *NotificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	notification.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": notification.ID}, notification)
	return err
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Count counts all notifications
func (r *NotificationRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
