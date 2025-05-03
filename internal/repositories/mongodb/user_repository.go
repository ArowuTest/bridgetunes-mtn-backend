package mongodb

import (
	"context"
	"errors"
	"fmt"     // Added missing import
	"strings"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options" // Removed unused import
)

// Compile-time check to ensure UserRepository implements the interface
var _ repositories.UserRepository = (*UserRepository)(nil)

// UserRepository handles MongoDB operations for User
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	filter := bson.M{"email": email}
	 err := r.collection.FindOne(ctx, filter).Decode(&user)
	 if err != nil {
		 return nil, err // Includes mongo.ErrNoDocuments
	 }
	 return &user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	filter := bson.M{"_id": id}
	 err := r.collection.FindOne(ctx, filter).Decode(&user)
	 if err != nil {
		 return nil, err // Includes mongo.ErrNoDocuments
	 }
	 return &user, nil
}

// FindByMSISDN finds a user by MSISDN
func (r *UserRepository) FindByMSISDN(ctx context.Context, msisdn string) (*models.User, error) {
	var user models.User
	filter := bson.M{"msisdn": msisdn}
	 err := r.collection.FindOne(ctx, filter).Decode(&user)
	 if err != nil {
		 return nil, err // Includes mongo.ErrNoDocuments
	 }
	 return &user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindAll retrieves all users (consider pagination for production)
func (r *UserRepository) FindAll(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	cursor, err := r.collection.Find(ctx, bson.M{})
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &users); err != nil {
		 return nil, err
	 }
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// FindByEligibleDigitsAndOptIn finds users who opted in and whose MSISDN ends with specific digits
func (r *UserRepository) FindByEligibleDigitsAndOptIn(ctx context.Context, digits []int, optInStatus bool, optInCutoff time.Time) ([]*models.User, error) {
	var users []*models.User
	// Build the regex pattern for ending digits
	var regexPatterns []string
	for _, digit := range digits {
		regexPatterns = append(regexPatterns, fmt.Sprintf("%d$", digit)) // Use fmt
	}
	regex := strings.Join(regexPatterns, "|")

	filter := bson.M{
		"optInStatus": optInStatus,
		"optInDate":   bson.M{"$lt": optInCutoff}, // Opted in before the cutoff
		"msisdn":      bson.M{"$regex": regex},
		// "isBlacklisted": false, // Add blacklist check later
	}

	cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &users); err != nil {
		 return nil, err
	 }
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// FindUsersByRechargeWindow finds users who had any recharge within the specified window
// Placeholder: This currently returns ALL users. Needs integration with Topup data.
func (r *UserRepository) FindUsersByRechargeWindow(ctx context.Context, startTime, endTime time.Time) ([]*models.User, error) {
	// TODO: Implement actual logic to query users based on topup records within the time window.
	// This might involve joining/looking up in the topups collection or having relevant flags/timestamps on the user model.
	// For now, returning all users as a placeholder for Pool A.
	return r.FindAll(ctx)
}

// FindEligibleConsolationUsers finds users eligible for consolation prizes
func (r *UserRepository) FindEligibleConsolationUsers(ctx context.Context, digits []int, optInCutoff time.Time, rechargeStart, rechargeEnd time.Time) ([]*models.User, error) {
	var users []*models.User
	// Build the regex pattern for ending digits
	var regexPatterns []string
	for _, digit := range digits {
		regexPatterns = append(regexPatterns, fmt.Sprintf("%d$", digit)) // Use fmt
	}
	regex := strings.Join(regexPatterns, "|")

	filter := bson.M{
		"optInStatus": true,
		"optInDate":   bson.M{"$lt": optInCutoff}, // Opted in before the cutoff
		"msisdn":      bson.M{"$regex": regex},
		"isBlacklisted": bson.M{"$ne": true}, // Exclude blacklisted users
		// TODO: Add filter based on recharge within [rechargeStart, rechargeEnd)
		// This requires joining/lookup with topup data or flags on the user model.
		// Example placeholder: "lastRechargeDate": bson.M{"gte": rechargeStart, "lt": rechargeEnd}
	}

	cursor, err := r.collection.Find(ctx, filter)
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &users); err != nil {
		 return nil, err
	 }
	 if users == nil {
		 users = []*models.User{}
	 }
	 return users, nil
}

// IncrementPoints atomically increments the points for a user
func (r *UserRepository) IncrementPoints(ctx context.Context, userID primitive.ObjectID, pointsToAdd int) error {
	 if pointsToAdd <= 0 {
		 return errors.New("points to add must be positive")
	 }
	 filter := bson.M{"_id": userID}
	 update := bson.M{"$inc": bson.M{"points": pointsToAdd}}
	 result, err := r.collection.UpdateOne(ctx, filter, update)
	 if err != nil {
		 return err
	 }
	 if result.MatchedCount == 0 {
		 return mongo.ErrNoDocuments // Or a more specific error
	 }
	 return nil
}


