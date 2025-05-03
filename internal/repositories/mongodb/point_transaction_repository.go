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

// PointTransactionRepository implements the repositories.PointTransactionRepository interface
type PointTransactionRepository struct {
	collection *mongo.Collection
}

// NewPointTransactionRepository creates a new PointTransactionRepository
func NewPointTransactionRepository(db *mongo.Database) repositories.PointTransactionRepository {
	return &PointTransactionRepository{
		collection: db.Collection("point_transactions"),
	}
}

// Create creates a new point transaction record.
func (r *PointTransactionRepository) Create(ctx context.Context, transaction *models.PointTransaction) error {
	 transaction.CreatedAt = time.Now()
	 _, err := r.collection.InsertOne(ctx, transaction)
	 if err != nil {
		 return fmt.Errorf("failed to create point transaction: %w", err)
	 }
	 return nil
}

// FindByID finds a point transaction by ID.
func (r *PointTransactionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.PointTransaction, error) {
	var transaction models.PointTransaction
	 err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&transaction)
	 if err != nil {
		 return nil, err
	}
	 return &transaction, nil
}

// FindByUserID finds point transactions for a user with pagination.
func (r *PointTransactionRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]*models.PointTransaction, error) {
	 opts := options.Find()
	 if page > 0 && limit > 0 {
		 opts.SetSkip(int64((page - 1) * limit))
		 opts.SetLimit(int64(limit))
	 }
	 opts.SetSort(bson.M{"transactionTimestamp": -1}) // Sort by transaction time descending

	 cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var transactions []*models.PointTransaction
	 if err := cursor.All(ctx, &transactions); err != nil {
		 return nil, err
	}
	 if transactions == nil {
		 transactions = []*models.PointTransaction{}
	 }
	 return transactions, nil
}

// FindByMSISDN finds point transactions for an MSISDN with pagination.
func (r *PointTransactionRepository) FindByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.PointTransaction, error) {
	 opts := options.Find()
	 if page > 0 && limit > 0 {
		 opts.SetSkip(int64((page - 1) * limit))
		 opts.SetLimit(int64(limit))
	 }
	 opts.SetSort(bson.M{"transactionTimestamp": -1}) // Sort by transaction time descending

	 cursor, err := r.collection.Find(ctx, bson.M{"msisdn": msisdn}, opts)
	 if err != nil {
		 return nil, err
	}
	 defer cursor.Close(ctx)

	 var transactions []*models.PointTransaction
	 if err := cursor.All(ctx, &transactions); err != nil {
		 return nil, err
	}
	 if transactions == nil {
		 transactions = []*models.PointTransaction{}
	 }
	 return transactions, nil
}

