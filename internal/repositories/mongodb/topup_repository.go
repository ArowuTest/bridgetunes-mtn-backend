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

// TopupRepository implements the repositories.TopupRepository interface
type TopupRepository struct {
	collection *mongo.Collection
}

// NewTopupRepository creates a new TopupRepository
func NewTopupRepository(db *mongo.Database) repositories.TopupRepository {
	return &TopupRepository{
		collection: db.Collection("topups"),
	}
}

// FindByID finds a topup by ID
func (r *TopupRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Topup, error) {
	var topup models.Topup
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&topup)
	if err != nil {
		return nil, err
	}
	return &topup, nil
}

// FindByMSISDN finds topups by MSISDN with pagination
func (r *TopupRepository) FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Topup, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"date": -1}) // Sort by date descending

	cursor, err := r.collection.Find(ctx, bson.M{"msisdn": msisdn}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topups []*models.Topup
	if err := cursor.All(ctx, &topups); err != nil {
		return nil, err
	}
	return topups, nil
}

// FindByDateRange finds topups by date range with pagination
func (r *TopupRepository) FindByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Topup, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"date": -1}) // Sort by date descending

	cursor, err := r.collection.Find(ctx, bson.M{
		"date": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topups []*models.Topup
	if err := cursor.All(ctx, &topups); err != nil {
		return nil, err
	}
	return topups, nil
}

// Create creates a new topup
func (r *TopupRepository) Create(ctx context.Context, topup *models.Topup) error {
	topup.CreatedAt = time.Now()
	topup.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, topup)
	return err
}

// Update updates a topup
func (r *TopupRepository) Update(ctx context.Context, topup *models.Topup) error {
	topup.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": topup.ID}, topup)
	return err
}

// Delete deletes a topup
func (r *TopupRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Count counts all topups
func (r *TopupRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}


// FindByMSISDNAndRef finds topups by MSISDN and Transaction Reference
func (r *TopupRepository) FindByMSISDNAndRef(ctx context.Context, msisdn string, transactionRef string) ([]*models.Topup, error) {
	filter := bson.M{
		"msisdn":         msisdn,
		"transactionRef": transactionRef,
	}

	cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, err
	 }
	defer cursor.Close(ctx)

	var topups []*models.Topup
	 if err := cursor.All(ctx, &topups); err != nil {
		 return nil, err
	 }
	 return topups, nil
}


