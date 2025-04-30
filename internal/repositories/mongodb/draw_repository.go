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

// FindByID finds a draw by ID
func (r *DrawRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Draw, error) {
	var draw models.Draw
	// Use constant DrawStatusCompleted from models or define locally if needed
	// Assuming models.DrawStatusCompleted exists and is "COMPLETED"
	filter := bson.M{"_id": id}

	 err := r.collection.FindOne(ctx, filter).Decode(&draw)
	 if err != nil {
		  return nil, err
	 }
	 return &draw, nil
}

// FindByDate finds a draw by date
func (r *DrawRepository) FindByDate(ctx context.Context, date time.Time) (*models.Draw, error) {
	// Create start and end of the day for the given date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	 endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	var draw models.Draw
	 err := r.collection.FindOne(ctx, bson.M{
		  "drawDate": bson.M{
			   "$gte": startOfDay,
			   "$lte": endOfDay,
		  },
	 }).Decode(&draw)
	 if err != nil {
		  return nil, err
	 }
	 return &draw, nil
}

// FindByDateRange finds draws by date range with pagination
func (r *DrawRepository) FindByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Draw, error) {
	 opts := options.Find()
	 if page > 0 && limit > 0 {
		  opts.SetSkip(int64((page - 1) * limit))
		  opts.SetLimit(int64(limit))
	 }
	 opts.SetSort(bson.M{"drawDate": -1}) // Sort by date descending

	 cursor, err := r.collection.Find(ctx, bson.M{
		  "drawDate": bson.M{
			   "$gte": start,
			   "$lte": end,
		  },
	 }, opts)
	 if err != nil {
		  return nil, err
	 }
	 defer cursor.Close(ctx)

	 var draws []*models.Draw
	 if err := cursor.All(ctx, &draws); err != nil {
		  return nil, err
	 }
	 return draws, nil
}

// FindByStatus finds draws by status with pagination
func (r *DrawRepository) FindByStatus(ctx context.Context, status string, page, limit int) ([]*models.Draw, error) {
	 opts := options.Find()
	 if page > 0 && limit > 0 {
		  opts.SetSkip(int64((page - 1) * limit))
		  opts.SetLimit(int64(limit))
	 }
	 opts.SetSort(bson.M{"drawDate": -1}) // Sort by date descending

	 cursor, err := r.collection.Find(ctx, bson.M{"status": status}, opts)
	 if err != nil {
		  return nil, err
	 }
	 defer cursor.Close(ctx)

	 var draws []*models.Draw
	 if err := cursor.All(ctx, &draws); err != nil {
		  return nil, err
	 }
	 return draws, nil
}

// FindCompletedWithJackpot finds completed draws containing a jackpot prize, limited and sorted by date descending
func (r *DrawRepository) FindCompletedWithJackpot(ctx context.Context, limit int) ([]*models.Draw, error) {
	 opts := options.Find().
		  SetSort(bson.M{"drawDate": -1}) // Sort by date descending
	 if limit > 0 {
		  opts.SetLimit(int64(limit))
	 }

	 // Assuming models.DrawStatusCompleted and models.JackpotCategory exist and are correct
	 filter := bson.M{
		  "status": models.DrawStatusCompleted, // Use constant if available
		  "prizes.category": models.JackpotCategory, // Use constant if available
	 }

	 cursor, err := r.collection.Find(ctx, filter, opts)
	 if err != nil {
		  return nil, err
	 }
	 defer cursor.Close(ctx)

	 var draws []*models.Draw
	 if err := cursor.All(ctx, &draws); err != nil {
		  return nil, err
	 }
	 return draws, nil
}

// FindMostRecentCompletedBefore finds the most recent completed draw before a given date
func (r *DrawRepository) FindMostRecentCompletedBefore(ctx context.Context, date time.Time) (*models.Draw, error) {
	 opts := options.FindOne().
		  SetSort(bson.M{"drawDate": -1}) // Find the latest one before the date

	 // Assuming models.DrawStatusCompleted exists and is correct
	 filter := bson.M{
		  "status": models.DrawStatusCompleted, // Use constant if available
		  "drawDate": bson.M{"$lt": date},
	 }

	 var draw models.Draw
	 err := r.collection.FindOne(ctx, filter, opts).Decode(&draw)
	 if err != nil {
		  // Includes mongo.ErrNoDocuments if none found
		  return nil, err
	 }
	 return &draw, nil
}

// Create creates a new draw
func (r *DrawRepository) Create(ctx context.Context, draw *models.Draw) error {
	 draw.CreatedAt = time.Now()
	 draw.UpdatedAt = time.Now()
	 _, err := r.collection.InsertOne(ctx, draw)
	 return err
}

// Update updates a draw
func (r *DrawRepository) Update(ctx context.Context, draw *models.Draw) error {
	 draw.UpdatedAt = time.Now()
	 _, err := r.collection.ReplaceOne(ctx, bson.M{"_id": draw.ID}, draw)
	 return err
}

// Delete deletes a draw
func (r *DrawRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	 _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	 return err
}

// Count counts all draws
func (r *DrawRepository) Count(ctx context.Context) (int64, error) {
	 return r.collection.CountDocuments(ctx, bson.M{})
}

