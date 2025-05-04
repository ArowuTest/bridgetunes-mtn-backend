package mongodb

import (
	"context"
	"errors"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Ensure adminUserRepository implements repositories.AdminUserRepository
var _ repositories.AdminUserRepository = (*adminUserRepository)(nil)

type adminUserRepository struct {
	collection *mongo.Collection
}

// NewAdminUserRepository creates a new repository for admin users
func NewAdminUserRepository(db *mongo.Database) repositories.AdminUserRepository {
	return &adminUserRepository{
		collection: db.Collection("admin_users"), // Use a dedicated collection for admins
	}
}

// Create inserts a new admin user into the database
// Corrected signature to match the interface: returns only error
func (r *adminUserRepository) Create(ctx context.Context, adminUser *models.AdminUser) error {
	adminUser.ID = primitive.NewObjectID() // Generate a new ID
	_, err := r.collection.InsertOne(ctx, adminUser)
	return err // Return only the error
}

// FindByEmail finds an admin user by their email address
func (r *adminUserRepository) FindByEmail(ctx context.Context, email string) (*models.AdminUser, error) {
	var adminUser models.AdminUser
	filter := bson.M{"email": email}
	 err := r.collection.FindOne(ctx, filter).Decode(&adminUser)
	 if err != nil {
	 	 if err == mongo.ErrNoDocuments {
	 	 	// Return the specific error so the service layer can distinguish 'not found' from other errors
	 	 	return nil, err
	 	 }
	 	return nil, err // Return other errors
	 }
	return &adminUser, nil
}

// FindByID finds an admin user by their ID
func (r *adminUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.AdminUser, error) {
	var adminUser models.AdminUser
	filter := bson.M{"_id": id}
	 err := r.collection.FindOne(ctx, filter).Decode(&adminUser)
	 if err != nil {
	 	 if err == mongo.ErrNoDocuments {
	 	 	return nil, errors.New("admin user not found") // Or return mongo.ErrNoDocuments
	 	 }
	 	return nil, err
	 }
	return &adminUser, nil
}

// Update updates an existing admin user
func (r *adminUserRepository) Update(ctx context.Context, adminUser *models.AdminUser) error {
	filter := bson.M{"_id": adminUser.ID}
	update := bson.M{"$set": adminUser}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes an admin user by ID
func (r *adminUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindAll retrieves all admin users
func (r *adminUserRepository) FindAll(ctx context.Context) ([]*models.AdminUser, error) {
	var adminUsers []*models.AdminUser
	cursor, err := r.collection.Find(ctx, bson.M{})
	 if err != nil {
	 	return nil, err
	 }
	defer cursor.Close(ctx)
	 if err = cursor.All(ctx, &adminUsers); err != nil {
	 	return nil, err
	 }
	return adminUsers, nil
}


