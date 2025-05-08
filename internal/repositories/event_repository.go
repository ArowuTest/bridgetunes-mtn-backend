package repositories

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindAll(ctx context.Context, page, limit int) ([]*models.Event, error)
}

type eventRepository struct {
	collection *mongo.Collection
}

func NewEventRepository(db *mongo.Database) EventRepository {
	return &eventRepository{
		collection: db.Collection("events"),
	}
}

func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	event.ID = primitive.NewObjectID()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *eventRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error) {
	var event models.Event
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) Update(ctx context.Context, event *models.Event) error {
	event.UpdatedAt = time.Now()
	
	update := bson.M{
		"$set": event,
	}
	
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": event.ID},
		update,
	)
	return err
}

func (r *eventRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *eventRepository) FindAll(ctx context.Context, page, limit int) ([]*models.Event, error) {
	skip := (page - 1) * limit
	
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})
	
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var events []*models.Event
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	
	return events, nil
} 