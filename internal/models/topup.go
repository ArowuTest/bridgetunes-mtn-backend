package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Topup represents a topup transaction
type Topup struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN         string             `bson:"msisdn" json:"msisdn"`
	Amount         float64            `bson:"amount" json:"amount"`
	Channel        string             `bson:"channel" json:"channel"`
	Date           time.Time          `bson:"date" json:"date"`
	TransactionRef string             `bson:"transactionRef" json:"transactionRef"`
	PointsEarned   int                `bson:"pointsEarned" json:"pointsEarned"`
	Processed      bool               `bson:"processed" json:"processed"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}
