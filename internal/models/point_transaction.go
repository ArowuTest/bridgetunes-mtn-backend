package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PointTransaction records points awarded for a specific top-up event.
type PointTransaction struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID             primitive.ObjectID `bson:"userId" json:"userId"` // Link to the User model
	MSISDN             string             `bson:"msisdn" json:"msisdn"` // Denormalized for easier querying?
	TopupAmount        float64            `bson:"topupAmount" json:"topupAmount"`
	PointsAwarded      int                `bson:"pointsAwarded" json:"pointsAwarded"`
	TransactionTimestamp time.Time          `bson:"transactionTimestamp" json:"transactionTimestamp"` // Time of the top-up event
	CreatedAt          time.Time          `bson:"createdAt" json:"createdAt"` // Time this record was created
}

