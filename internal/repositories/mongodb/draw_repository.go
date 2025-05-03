package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DrawRepository implements the repositories.DrawRepository interface
type DrawRepository struct {
	collection *mongo.Collection
}

// NewDrawRepository creates a new DrawRepository
func NewDrawRepository(db *mongo.Database) repositories.DrawRepository {
	return &DrawRepository{
		collection: db.Collection("draws"),
	}
}

// Create creates a new draw
func (r *DrawRepository) Create(ctx context.Context, draw *models.Draw) error {
	 draw.CreatedAt = time.Now()
	 draw.UpdatedAt = time.Now()
	 res, err := r.collection.InsertOne(ctx, draw)
	 if err != nil {
		 return err
	}
	 draw.ID = res.InsertedID.(primitive.ObjectID)
	 return nil
}

// FindByID finds a draw by ID
func (r *DrawRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error) {
	var draw models.Draw
	 err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&draw)
	 if err != nil {
		 return nil, err
	}
	 return &draw, nil
}

// FindByDate finds a draw by date (matching the start of the day)
func (r *DrawRepository) FindByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	var draw models.Draw
	 // Match draws where the DrawDate is on the specified day
	 startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	 endOfDay := startOfDay.AddDate(0, 0, 1)
	 filter := bson.M{
		 "drawDate": bson.M{
			 "$gte": startOfDay,
			 "$lt":  endOfDay,
		 },
	 }
	 err := r.collection.FindOne(ctx, filter).Decode(&draw)
	 if err != nil {
		 return nil, err // Returns mongo.ErrNoDocuments if not found
	}
	 return &draw, nil
}

// FindByDateRange finds draws within a date range
func (r *DrawRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error) {
	filter := bson.M{}
	 dateFilter := bson.M{}
	 if !startDate.IsZero() {
		 dateFilter["$gte"] = startDate
	 }
	 if !endDate.IsZero() {
		 dateFilter["$lt"] = endDate
	 }
	 if len(dateFilter) > 0 {
		 filter["drawDate"] = dateFilter
	 }

	 opts := options.Find().SetSort(bson.M{"drawDate": -1}) // Sort by date descending
	 cursor, err := r.collection.Find(ctx, filter, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var draws []*models.Draw
	 if err := cursor.All(ctx, &draws); err != nil {
		 return nil, err
	}
	 if draws == nil {
		 draws = []*models.Draw{}
	 }
	 return draws, nil
}

// FindByStatus finds draws by status
func (r *DrawRepository) FindByStatus(ctx context.Context, status string) ([]*models.Draw, error) {
	filter := bson.M{"status": status}
	 opts := options.Find().SetSort(bson.M{"drawDate": -1}) // Sort by date descending
	 cursor, err := r.collection.Find(ctx, filter, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var draws []*models.Draw
	 if err := cursor.All(ctx, &draws); err != nil {
		 return nil, err
	}
	 if draws == nil {
		 draws = []*models.Draw{}
	 }
	 return draws, nil
}

// FindNextScheduledDraw finds the next scheduled draw after a given date
func (r *DrawRepository) FindNextScheduledDraw(ctx context.Context, afterDate time.Time) (*models.Draw, error) {
	filter := bson.M{
		 "status": models.DrawStatusScheduled,
		 "drawDate": bson.M{"$gt": afterDate},
	 }
	 opts := options.FindOne().SetSort(bson.M{"drawDate": 1}) // Find the earliest one after the date

	 var draw models.Draw
	 err := r.collection.FindOne(ctx, filter, opts).Decode(&draw)
	 if err != nil {
		 if errors.Is(err, mongo.ErrNoDocuments) {
			 return nil, fmt.Errorf("no scheduled draw found after %s", afterDate.Format("2006-01-02"))
		 }
		 return nil, fmt.Errorf("failed to find next scheduled draw: %w", err)
	 }
	 return &draw, nil
}

// Update updates a draw
func (r *DrawRepository) Update(ctx context.Context, draw *models.Draw) error {
	 draw.UpdatedAt = time.Now()
	 _, err := r.collection.ReplaceOne(ctx, bson.M{"_id": draw.ID}, draw)
	 return err
}

// Delete deletes a draw by ID
func (r *DrawRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	 _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	 return err
}

// FindAll finds all draws (consider pagination for large datasets)
func (r *DrawRepository) FindAll(ctx context.Context) ([]*models.Draw, error) {
	 opts := options.Find().SetSort(bson.M{"drawDate": -1}) // Sort by date descending
	 cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var draws []*models.Draw
	 if err := cursor.All(ctx, &draws); err != nil {
		 return nil, err
	}
	 if draws == nil {
		 draws = []*models.Draw{}
	 }
	 return draws, nil
}


