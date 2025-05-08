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



type EventRepository struct {
	collection *mongo.Collection
}

func NewEventRepository(db *mongo.Database) repositories.EventRepository {
	return &EventRepository{
		collection: db.Collection("events"),
	}
}

func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *EventRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error) {
	var event models.Event
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) Update(ctx context.Context, event *models.Event) error {
	event.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": event.ID}, event)
	return err
}

func (r *EventRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *EventRepository) FindAll(ctx context.Context, page, limit int) ([]*models.Event, error) {
	opts := options.Find().
	SetSkip(int64((page - 1) * limit)).
	SetLimit(int64(limit))
	// .SetSort(bson.M{"scheduledAt": -1}) // Sort by scheduled date descending

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
   }
	 defer cursor.Close(ctx)

	 var events []*models.Event // Changed to pointer slice
	 if err := cursor.All(ctx, &events); err != nil {
		return nil, err
   }

   	 // Ensure an empty slice is returned instead of nil if no campaigns found
	 if events == nil {
		events = []*models.Event{}
	}
	return events, nil
} 