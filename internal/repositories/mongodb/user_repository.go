package mongodb

import (
	"context"
	"errors" // Added for error handling
	"strconv" // Added for Itoa
	"strings" // Added for Join
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository implements the repositories.UserRepository interface
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *mongo.Database) repositories.UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User

	 err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	 if err != nil {
		 return nil, err
	}

	 return &user, nil
}

// FindByMSISDN finds a user by MSISDN
func (r *UserRepository) FindByMSISDN(ctx context.Context, msisdn string) (*models.User, error) {
	var user models.User

	 err := r.collection.FindOne(ctx, bson.M{"msisdn": msisdn}).Decode(&user)
	 if err != nil {
		 return nil, err
	}

	 return &user, nil
}

// FindAll finds all users with pagination
func (r *UserRepository) FindAll(ctx context.Context, page, limit int) ([]*models.User, error) {
	 opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))

	 cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, err
	}

	 return users, nil
}

// FindByOptInStatus finds users by opt-in status with pagination
func (r *UserRepository) FindByOptInStatus(ctx context.Context, optInStatus bool, page, limit int) ([]*models.User, error) {
	 opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))

	 cursor, err := r.collection.Find(ctx, bson.M{"optInStatus": optInStatus}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, err
	}

	 return users, nil
}

// FindByEligibleDigits finds users by eligible digits (last digits of MSISDN)
func (r *UserRepository) FindByEligibleDigits(ctx context.Context, digits []int, optInStatus bool) ([]*models.User, error) {
	// Handle empty digits slice
	 if len(digits) == 0 {
		 return []*models.User{}, nil // Return empty slice, no users match
	}

	// Convert digits to strings for regex matching
	 var digitStrings []string
	 for _, digit := range digits {
		 if digit < 0 || digit > 9 {
			 return nil, errors.New("invalid digit provided") // Basic validation
		}
		 // Correctly convert int digit to string using strconv.Itoa
		 digitStrings = append(digitStrings, strconv.Itoa(digit))
	}

	// Create regex pattern for matching last digit (e.g., "(0|1|5)$" )
	 regexPattern := "(" + strings.Join(digitStrings, "|") + ")$"

	// Find users with matching last digit and opt-in status
	 filter := bson.M{
		"msisdn":      bson.M{"regex": regexPattern},
		"optInStatus": optInStatus,
	}

	 cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, err
	}

	 return users, nil
}

// FindByRechargeTimeRange finds users who recharged within a specific time range
func (r *UserRepository) FindByRechargeTimeRange(ctx context.Context, start, end time.Time) ([]*models.User, error) {
	// This assumes the User model has a field like LastRechargeDate or similar.
	// If recharge info is in a separate Topup collection, this logic needs adjustment.
	// For now, assuming User has LastRechargeDate.
	 filter := bson.M{
		"lastRechargeDate": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}

	 cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, err
	}

	 return users, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	 user.CreatedAt = time.Now()
	 user.UpdatedAt = time.Now()
	 _, err := r.collection.InsertOne(ctx, user)
	 return err
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	 user.UpdatedAt = time.Now()
	 _, err := r.collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	 return err
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	 _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	 return err
}

// Count counts all users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	 return r.collection.CountDocuments(ctx, bson.M{})
}

