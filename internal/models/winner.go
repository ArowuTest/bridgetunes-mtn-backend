package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Winner represents a winner in a draw
type Winner struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN       string             `bson:"msisdn" json:"msisdn"`
	DrawID       primitive.ObjectID `bson:"drawId" json:"drawId"`
	PrizeCategory string            `bson:"prizeCategory" json:"prizeCategory"`
	PrizeAmount  float64            `bson:"prizeAmount" json:"prizeAmount"`
	WinDate      time.Time          `bson:"winDate" json:"winDate"`
	ClaimStatus  string             `bson:"claimStatus" json:"claimStatus"` // PENDING, CLAIMED, FORFEITED
	ClaimDate    time.Time          `bson:"claimDate,omitempty" json:"claimDate,omitempty"`
	NotifiedAt   time.Time          `bson:"notifiedAt,omitempty" json:"notifiedAt,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}
