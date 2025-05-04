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

// JackpotRolloverRepository implements the repositories.JackpotRolloverRepository interface
type JackpotRolloverRepository struct {
	collection *mongo.Collection
}

// NewJackpotRolloverRepository creates a new JackpotRolloverRepository
func NewJackpotRolloverRepository(db *mongo.Database) repositories.JackpotRolloverRepository {
	return &JackpotRolloverRepository{
		collection: db.Collection("jackpot_rollovers"),
	}
}

// Create creates a new jackpot rollover record.
func (r *JackpotRolloverRepository) Create(ctx context.Context, rollover *models.JackpotRollover) error {
	 rollover.CreatedAt = time.Now()
	 _, err := r.collection.InsertOne(ctx, rollover)
	 if err != nil {
		 return fmt.Errorf("failed to create jackpot rollover record: %w", err)
	 }
	 return nil
}

// FindByID finds a jackpot rollover record by ID.
func (r *JackpotRolloverRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.JackpotRollover, error) {
	var rollover models.JackpotRollover
	 err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rollover)
	 if err != nil {
		 return nil, err
	}
	 return &rollover, nil
}

// FindBySourceDrawID finds a jackpot rollover record by the source draw ID.
func (r *JackpotRolloverRepository) FindBySourceDrawID(ctx context.Context, sourceDrawID primitive.ObjectID) (*models.JackpotRollover, error) {
	var rollover models.JackpotRollover
	 err := r.collection.FindOne(ctx, bson.M{"sourceDrawId": sourceDrawID}).Decode(&rollover)
	 if err != nil {
		 // Return error, including mongo.ErrNoDocuments if not found
		 return nil, err
	}
	 return &rollover, nil
}

// FindRolloversByDestinationDate finds all rollover records targeting a specific destination date.
func (r *JackpotRolloverRepository) FindRolloversByDestinationDate(ctx context.Context, destinationDate time.Time) ([]*models.JackpotRollover, error) {
	 // Match the specific date, ignoring time component if necessary, or use a range
	 startOfDay := time.Date(destinationDate.Year(), destinationDate.Month(), destinationDate.Day(), 0, 0, 0, 0, destinationDate.Location())
	 endOfDay := startOfDay.AddDate(0, 0, 1)

	 filter := bson.M{
		 "destinationDrawDate": bson.M{
			 "$gte": startOfDay,
			 "$lt":  endOfDay,
		 },
	 }
	 opts := options.Find().SetSort(bson.M{"createdAt": 1}) // Sort by creation time ascending

	 cursor, err := r.collection.Find(ctx, filter, opts)
	 if err != nil {
		 return nil, fmt.Errorf("error finding rollovers by destination date %s: %w", destinationDate.Format("2006-01-02"), err)
	}
	 defer cursor.Close(ctx)

	 var rollovers []*models.JackpotRollover
	 if err := cursor.All(ctx, &rollovers); err != nil {
		 return nil, fmt.Errorf("error decoding rollovers for destination date %s: %w", destinationDate.Format("2006-01-02"), err)
	}
	 if rollovers == nil {
		 rollovers = []*models.JackpotRollover{}
	 }
	 return rollovers, nil
}



// FindPendingRollovers finds rollover records created after a certain date or targeting a future date.
// Used by GetJackpotStatus to calculate the current effective jackpot amount.
func (r *JackpotRolloverRepository) FindPendingRollovers(ctx context.Context, effectiveDate time.Time) ([]*models.JackpotRollover, error) {
	// Find rollovers where the destination date is after the effective date
	filter := bson.M{
		"destinationDrawDate": bson.M{
			"$gt": effectiveDate,
		},
	}
	// Optionally, could also include rollovers created after effectiveDate, regardless of destination?
	// filter := bson.M{
	// 	"$or": []bson.M{
	// 		{"destinationDrawDate": bson.M{"$gt": effectiveDate}},
	// 		{"createdAt": bson.M{"$gt": effectiveDate}},
	// 	},
	// }

	 opts := options.Find().SetSort(bson.M{"createdAt": 1}) // Sort by creation time ascending

	 cursor, err := r.collection.Find(ctx, filter, opts)
	 if err != nil {
		 return nil, fmt.Errorf("error finding pending rollovers after %s: %w", effectiveDate.Format(time.RFC3339), err)
	 }
	 defer cursor.Close(ctx)

	 var rollovers []*models.JackpotRollover
	 if err := cursor.All(ctx, &rollovers); err != nil {
		 return nil, fmt.Errorf("error decoding pending rollovers after %s: %w", effectiveDate.Format(time.RFC3339), err)
	 }
	 if rollovers == nil {
		 rollovers = []*models.JackpotRollover{}
	 }
	 return rollovers, nil
}


