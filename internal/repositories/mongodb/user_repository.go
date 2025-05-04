package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings" // Added import
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options" // Removed unused import
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

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	 user.CreatedAt = time.Now()
	 user.UpdatedAt = time.Now()
	 _, err := r.collection.InsertOne(ctx, user)
	 return err
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	 err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	 if err != nil {
		 return nil, err
	}
	 return &user, nil
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

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	 user.UpdatedAt = time.Now()
	 _, err := r.collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	 return err
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	 _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	 return err
}

// FindAll finds all users (consider pagination for large datasets)
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

// FindByEligibleDigitsAndOptIn finds users matching digits and opt-in status/time
func (r *UserRepository) FindByEligibleDigitsAndOptIn(ctx context.Context, digits []int, optInStatus bool, optInCutoff time.Time) ([]*models.User, error) {
	filter := bson.M{
		 "optInStatus": optInStatus,
		 "optInDate":   bson.M{"lt": optInCutoff}, // Opted in before the cutoff
	 }

	 if len(digits) > 0 {
		 // Build regex to match numbers ending in the specified digits
		 regexPatterns := []string{}
		 for _, digit := range digits {
			 regexPatterns = append(regexPatterns, fmt.Sprintf("%d$", digit))
		 }
		 filter["msisdn"] = bson.M{"regex": primitive.Regex{Pattern: strings.Join(regexPatterns, "|"), Options: ""}}
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
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// FindUsersByRechargeWindow finds users who recharged within a specific time window.
// This assumes recharge data is stored/linked within the User model or another collection.
// Placeholder implementation - needs actual recharge data query.
func (r *UserRepository) FindUsersByRechargeWindow(ctx context.Context, startTime, endTime time.Time) ([]*models.User, error) {
	 // TODO: Implement actual query based on how recharge data is stored.
	 // This might involve querying a separate 'recharges' collection and joining
	 // or checking a 'lastRechargeDate' field on the user model.
	 // For now, returning all users as a placeholder.
	 fmt.Println("WARN: FindUsersByRechargeWindow using placeholder logic - returning all users")
	 return r.FindAll(ctx)
}

// FindEligibleConsolationUsers finds users eligible for consolation prizes.
// Combines digit matching, opt-in status/time, recharge window, and blacklist check.
func (r *UserRepository) FindEligibleConsolationUsers(ctx context.Context, digits []int, optInCutoff, rechargeStart, rechargeEnd time.Time) ([]*models.User, error) {
	 // 1. Base filter: Opt-in status and time
	 filter := bson.M{
		 "optInStatus": true,
		 "optInDate":   bson.M{"lt": optInCutoff},
	 }

	 // 2. Add digit matching if applicable
	 if len(digits) > 0 {
		 regexPatterns := []string{}
		 for _, digit := range digits {
			 regexPatterns = append(regexPatterns, fmt.Sprintf("%d$", digit))
		 }
		 filter["msisdn"] = bson.M{"regex": primitive.Regex{Pattern: strings.Join(regexPatterns, "|"), Options: ""}}
	 }

	 // 3. Add recharge window filter (Placeholder - needs real implementation)
	 // filter["lastRechargeDate"] = bson.M{"gte": rechargeStart, "lt": rechargeEnd}
	 fmt.Println("WARN: FindEligibleConsolationUsers recharge window filter is a placeholder")

	 // 4. Add blacklist filter (Assuming IsBlacklisted is handled elsewhere or needs integration)
	 // filter["isBlacklisted"] = false // Or query blacklist collection separately
	 fmt.Println("WARN: FindEligibleConsolationUsers blacklist filter is not implemented")

	 cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, fmt.Errorf("failed to query eligible consolation users: %w", err)
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, fmt.Errorf("failed to decode eligible consolation users: %w", err)
	}
	 if users == nil {
		 users = []*models.User{}
	 }

	 // TODO: Post-query filtering might be needed if recharge/blacklist checks aren't in the main query

	 return users, nil
}

// IncrementPoints atomically increments the points for a user.
func (r *UserRepository) IncrementPoints(ctx context.Context, userID primitive.ObjectID, points int) error {
	 if points <= 0 {
		 return errors.New("points to add must be positive")
	 }
	 filter := bson.M{"_id": userID}
	 update := bson.M{
		 "$inc": bson.M{"points": points},
		 "$set": bson.M{"updatedAt": time.Now()},
	 }
	 result, err := r.collection.UpdateOne(ctx, filter, update)
	 if err != nil {
		 return fmt.Errorf("failed to increment points for user %s: %w", userID.Hex(), err)
	 }
	 if result.MatchedCount == 0 {
		 return fmt.Errorf("user %s not found for point increment", userID.Hex())
	 }
	 return nil
}


// Count returns the total number of users.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	 if err != nil {
		 return 0, fmt.Errorf("failed to count users: %w", err)
	}
	 return count, nil
}


// FindByEligibleDigits finds users matching the last digits of their MSISDN.
func (r *UserRepository) FindByEligibleDigits(ctx context.Context, digits []int) ([]*models.User, error) {
	filter := bson.M{}

	 if len(digits) > 0 {
		 // Build regex to match numbers ending in the specified digits
		 regexPatterns := []string{}
		 for _, digit := range digits {
			 regexPatterns = append(regexPatterns, fmt.Sprintf("%d$", digit))
		 }
		 filter["msisdn"] = bson.M{"regex": primitive.Regex{Pattern: strings.Join(regexPatterns, "|"), Options: ""}}
	 } else {
		 // If no digits are provided, maybe return all users or an empty list? Returning empty for now.
		 return []*models.User{}, nil
	 }

	 cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, fmt.Errorf("failed to query users by eligible digits: %w", err)
	}
	 defer cursor.Close(ctx)

	 var users []*models.User
	 if err := cursor.All(ctx, &users); err != nil {
		 return nil, fmt.Errorf("failed to decode users by eligible digits: %w", err)
	}
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}




