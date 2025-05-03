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

// Compile-time check to ensure PointTransactionRepository implements the interface
var _ repositories.PointTransactionRepository = (*PointTransactionRepository)(nil)

// PointTransactionRepository handles MongoDB operations for PointTransaction
type PointTransactionRepository struct {
	collection *mongo.Collection
}

// NewPointTransactionRepository creates a new PointTransactionRepository
func NewPointTransactionRepository(db *mongo.Database) *PointTransactionRepository {
	return &PointTransactionRepository{
		collection: db.Collection("point_transactions"),
	}
}

// Create inserts a new point transaction record
func (r *PointTransactionRepository) Create(ctx context.Context, transaction *models.PointTransaction) error {
	transaction.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, transaction)
	return err
}

// FindByUserID finds all point transactions for a specific user
// Corrected signature to match interface: returns ([]*models.PointTransaction, error)
func (r *PointTransactionRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.PointTransaction, error) {
	var transactions []*models.PointTransaction
	filter := bson.M{"userId": userID}
	// Optional: Add sorting, e.g., by timestamp descending
	findOptions := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	 if err != nil {
		 return nil, err
	 }
	 defer cursor.Close(ctx)

	 if err = cursor.All(ctx, &transactions); err != nil {
		 return nil, err
	 }

	 // Return empty slice instead of nil if no documents found
	 if transactions == nil {
		 transactions = []*models.PointTransaction{}
	 }

	 return transactions, nil
}

