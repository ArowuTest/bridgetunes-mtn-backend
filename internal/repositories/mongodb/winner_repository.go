package mongodb

import (
	"context"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Compile-time check to ensure WinnerRepository implements the interface
var _ repositories.WinnerRepository = (*WinnerRepository)(nil)

// WinnerRepository handles MongoDB operations for Winner
type WinnerRepository struct {
	collection *mongo.Collection
}

// NewWinnerRepository creates a new WinnerRepository
func NewWinnerRepository(db *mongo.Database) *WinnerRepository {
	return &WinnerRepository{
		collection: db.Collection("winners"),
	}
}

// Create inserts a new winner record
func (r *WinnerRepository) Create(ctx context.Context, winner *models.Winner) error {
	winner.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, winner)
	return err
}

// CreateMany inserts multiple winner records
func (r *WinnerRepository) CreateMany(ctx context.Context, winners []*models.Winner) error {
	 if len(winners) == 0 {
		 return nil
	 }
	 // Assign ObjectIDs if not already assigned
	 docs := make([]interface{}, len(winners))
	 for i, w := range winners {
		 if w.ID.IsZero() {
			 w.ID = primitive.NewObjectID()
		 }
		 docs[i] = w
	 }
	 _, err := r.collection.InsertMany(ctx, docs)
	 return err
}

// FindByDrawID finds all winners for a specific draw
func (r *WinnerRepository) FindByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	var winners []*models.Winner
	filter := bson.M{"drawId": drawID}
	// Optional: Sort by prize category or amount
	findOptions := options.Find() //.SetSort(bson.D{{Key: "prize.amount", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &winners); err != nil {
		 return nil, err
	 }

	 // Return empty slice instead of nil if no documents found
	 if winners == nil {
		 winners = []*models.Winner{}
	 }

	 return winners, nil
}

// FindAll retrieves all winners (consider pagination for production)
func (r *WinnerRepository) FindAll(ctx context.Context) ([]*models.Winner, error) {
	var winners []*models.Winner
	cursor, err := r.collection.Find(ctx, bson.M{})
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &winners); err != nil {
		 return nil, err
	 }
	 if winners == nil {
		 winners = []*models.Winner{}
	 }
	 return winners, nil
}
