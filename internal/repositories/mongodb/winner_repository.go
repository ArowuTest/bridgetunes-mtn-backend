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

// WinnerRepository implements the repositories.WinnerRepository interface
type WinnerRepository struct {
	collection *mongo.Collection
}

// NewWinnerRepository creates a new WinnerRepository
func NewWinnerRepository(db *mongo.Database) repositories.WinnerRepository {
	return &WinnerRepository{
		collection: db.Collection("winners"),
	}
}

// FindByID finds a winner by ID
func (r *WinnerRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Winner, error) {
	var winner models.Winner
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&winner)
	if err != nil {
		return nil, err
	}
	return &winner, nil
}

// FindByMSISDN finds winners by MSISDN with pagination
func (r *WinnerRepository) FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Winner, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"winDate": -1}) // Sort by win date descending

	cursor, err := r.collection.Find(ctx, bson.M{"msisdn": msisdn}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var winners []*models.Winner
	if err := cursor.All(ctx, &winners); err != nil {
		return nil, err
	}
	return winners, nil
}

// FindByDrawID finds winners by draw ID with pagination
func (r *WinnerRepository) FindByDrawID(ctx context.Context, drawID primitive.ObjectID, page, limit int) ([]*models.Winner, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"winDate": -1}) // Sort by win date descending

	cursor, err := r.collection.Find(ctx, bson.M{"drawId": drawID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var winners []*models.Winner
	if err := cursor.All(ctx, &winners); err != nil {
		return nil, err
	}
	return winners, nil
}

// FindByClaimStatus finds winners by claim status with pagination
func (r *WinnerRepository) FindByClaimStatus(ctx context.Context, status string, page, limit int) ([]*models.Winner, error) {
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"winDate": -1}) // Sort by win date descending

	cursor, err := r.collection.Find(ctx, bson.M{"claimStatus": status}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var winners []*models.Winner
	if err := cursor.All(ctx, &winners); err != nil {
		return nil, err
	}
	return winners, nil
}

// Create creates a new winner
func (r *WinnerRepository) Create(ctx context.Context, winner *models.Winner) error {
	winner.CreatedAt = time.Now()
	winner.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, winner)
	return err
}

// Update updates a winner
func (r *WinnerRepository) Update(ctx context.Context, winner *models.Winner) error {
	winner.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": winner.ID}, winner)
	return err
}

// Delete deletes a winner
func (r *WinnerRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Count counts all winners
func (r *WinnerRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
