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
	 if winners == nil {
		 winners = []*models.Winner{}
	 }
	 return winners, nil
}

// FindByDrawID finds all winners for a specific draw ID (pagination removed as per interface update)
func (r *WinnerRepository) FindByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error) {
	 opts := options.Find().
		 SetSort(bson.M{"prizeAmount": -1, "winDate": 1}) // Sort by prize amount descending, then date ascending

	 cursor, err := r.collection.Find(ctx, bson.M{"drawId": drawID}, opts)
	 if err != nil {
		 return nil, fmt.Errorf("error finding winners by draw ID %s: %w", drawID.Hex(), err)
	}
	 defer cursor.Close(ctx)

	 var winners []*models.Winner
	 if err := cursor.All(ctx, &winners); err != nil {
		 return nil, fmt.Errorf("error decoding winners for draw ID %s: %w", drawID.Hex(), err)
	}
	 if winners == nil {
		 winners = []*models.Winner{}
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
	 if winners == nil {
		 winners = []*models.Winner{}
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

// CreateMany creates multiple winner records efficiently.
func (r *WinnerRepository) CreateMany(ctx context.Context, winners []*models.Winner) error {
	 if len(winners) == 0 {
		 return nil // Nothing to insert
	 }

	 // Convert []*models.Winner to []interface{} for InsertMany
	 docs := make([]interface{}, len(winners))
	 now := time.Now()
	 for i, w := range winners {
		 w.CreatedAt = now
		 w.UpdatedAt = now
		 // Ensure ID is not set if it's zero, allowing MongoDB to generate it
		 if w.ID.IsZero() {
			 // No need to explicitly set ID to nil, InsertMany handles it
		 } else {
			 // If ID is pre-set, ensure it's used (though usually not recommended for CreateMany)
		 }
		 docs[i] = w
	 }

	 _, err := r.collection.InsertMany(ctx, docs)
	 if err != nil {
		 return fmt.Errorf("failed to insert multiple winners: %w", err)
	 }
	 return nil
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



// FindByDrawIDAndCategory finds winners for a specific draw ID and prize category.
func (r *WinnerRepository) FindByDrawIDAndCategory(ctx context.Context, drawID primitive.ObjectID, category string) ([]*models.Winner, error) {
	filter := bson.M{
		"drawId":       drawID,
		"prizeCategory": category,
	}
	// No specific sorting needed here, but could add if required
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error finding winners by draw ID %s and category %s: %w", drawID.Hex(), category, err)
	}
	defer cursor.Close(ctx)

	var winners []*models.Winner
	 if err := cursor.All(ctx, &winners); err != nil {
		 return nil, fmt.Errorf("error decoding winners for draw ID %s and category %s: %w", drawID.Hex(), category, err)
	}
	 if winners == nil {
		 winners = []*models.Winner{}
	 }
	 return winners, nil
}

