package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	// Add blacklist collection if needed for direct checks, or rely on service layer
	// blacklistCollection *mongo.Collection
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *mongo.Database) repositories.UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
		// blacklistCollection: db.Collection("blacklist"), // Example if needed
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

// FindAll finds all users (consider adding pagination back if needed)
func (r *UserRepository) FindAll(ctx context.Context) ([]*models.User, error) {
	 cursor, err := r.collection.Find(ctx, bson.M{})
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, err
	}
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// FindByOptInStatus finds users by opt-in status (consider adding pagination back if needed)
func (r *UserRepository) FindByOptInStatus(ctx context.Context, optInStatus bool) ([]*models.User, error) {
	 cursor, err := r.collection.Find(ctx, bson.M{"optInStatus": optInStatus})
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, err
	}
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// --- New/Updated Methods from Redesign ---

// FindUsersByRechargeWindow finds users who recharged within a specific time range
// Assumes User model has LastTopupTimestamp field updated upon recharge.
func (r *UserRepository) FindUsersByRechargeWindow(ctx context.Context, start, end time.Time) ([]*models.User, error) {
	 filter := bson.M{
		// Assuming the user model has a field like lastTopupTimestamp
		"lastTopupTimestamp": bson.M{
			"$gte": start,
			"$lt":  end,
		},
		// Optionally add other base criteria like not being blacklisted if the flag exists
		// "isBlacklisted": bson.M{"ne": true},
	}

	 cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, fmt.Errorf("error finding users by recharge window: %w", err)
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, fmt.Errorf("error decoding users by recharge window: %w", err)
	}
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// FindEligibleConsolationUsers finds users eligible for consolation prizes based on multiple criteria.
// Assumes User model has OptInStatus, OptInDate, LastTopupTimestamp, and IsBlacklisted fields.
func (r *UserRepository) FindEligibleConsolationUsers(ctx context.Context, digits []int, optInCutoff time.Time, rechargeStart time.Time, rechargeEnd time.Time) ([]*models.User, error) {
	// Build the core filter
	 filter := bson.M{
		"optInStatus": true,
		"optInDate":   bson.M{"lt": optInCutoff}, // Opted in before the cutoff
		// Assuming lastTopupTimestamp exists and is relevant for the draw window
		"lastTopupTimestamp": bson.M{
			"$gte": rechargeStart,
			"$lt":  rechargeEnd,
		},
		"isBlacklisted": bson.M{"ne": true}, // Exclude blacklisted users
	}

	// Add digit matching if digits are provided
	 if len(digits) > 0 {
		 var digitStrings []string
		 for _, digit := range digits {
			 if digit < 0 || digit > 9 {
				 return nil, errors.New("invalid digit provided")
			}
			 digitStrings = append(digitStrings, strconv.Itoa(digit))
		}
		 regexPattern := "(" + strings.Join(digitStrings, "|") + ")$"
		 filter["msisdn"] = bson.M{"regex": regexPattern}
	}

	 cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, fmt.Errorf("error finding eligible consolation users: %w", err)
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, fmt.Errorf("error decoding eligible consolation users: %w", err)
	}
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// IncrementPoints atomically increments the points for a given user.
func (r *UserRepository) IncrementPoints(ctx context.Context, userID primitive.ObjectID, pointsToAdd int) error {
	 if pointsToAdd <= 0 {
		 return errors.New("pointsToAdd must be positive")
	 }

	 filter := bson.M{"_id": userID}
	 update := bson.M{
		"$inc": bson.M{"points": pointsToAdd},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	 result, err := r.collection.UpdateOne(ctx, filter, update)
	 if err != nil {
		 return fmt.Errorf("failed to increment points for user %s: %w", userID.Hex(), err)
	 }

	 if result.MatchedCount == 0 {
		 return fmt.Errorf("user %s not found for point increment", userID.Hex())
	 }
	 // result.ModifiedCount could also be checked if needed

	 return nil
}

// --- Standard CRUD Methods (Ensure compatibility with updated User model) ---

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	 user.CreatedAt = time.Now()
	 user.UpdatedAt = time.Now()
	 // Ensure default values are set if needed (e.g., Points = 0)
	 if user.Points < 0 { user.Points = 0 }
	 _, err := r.collection.InsertOne(ctx, user)
	 return err
}

// Update updates a user (using ReplaceOne, consider UpdateOne with $set for partial updates)
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

// --- Deprecated/Replaced Methods (Keep or remove?) ---

/*
// FindByEligibleDigits finds users by eligible digits (last digits of MSISDN)
// Deprecated: Replaced by FindEligibleConsolationUsers which includes more criteria
func (r *UserRepository) FindByEligibleDigits(ctx context.Context, digits []int, optInStatus bool) ([]*models.User, error) {
	 // ... (implementation kept for reference if needed) ...
}
*/

/*
// FindByRechargeTimeRange finds users who recharged within a specific time range
// Deprecated: Replaced by FindUsersByRechargeWindow
func (r *UserRepository) FindByRechargeTimeRange(ctx context.Context, start, end time.Time) ([]*models.User, error) {
	 // ... (implementation kept for reference if needed) ...
}
*/

