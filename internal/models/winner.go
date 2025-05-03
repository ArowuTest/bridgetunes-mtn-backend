package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Winner represents a winner in a draw
type Winner struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN         string             `bson:"msisdn" json:"msisdn"`
	MaskedMSISDN   string             `bson:"maskedMsisdn" json:"maskedMsisdn"` // Stores masked number, e.g., "080*****178"
	DrawID         primitive.ObjectID `bson:"drawId" json:"drawId"`
	PrizeCategory  string             `bson:"prizeCategory" json:"prizeCategory"`
	PrizeAmount    float64            `bson:"prizeAmount" json:"prizeAmount"`
	IsOptedIn      bool               `bson:"isOptedIn" json:"isOptedIn"` // User's opt-in status at time of draw
	IsValid        bool               `bson:"isValid" json:"isValid"` // Whether the win is valid (e.g., opted-in for jackpot)
	Points         int                `bson:"points" json:"points"` // User's points at time of draw
	WinDate        time.Time          `bson:"winDate" json:"winDate"`
	ClaimStatus    string             `bson:"claimStatus" json:"claimStatus"` // PENDING, CLAIMED, EXPIRED
	ClaimDate      *time.Time         `bson:"claimDate,omitempty" json:"claimDate,omitempty"` // Pointer to allow null
	NotifiedAt     *time.Time         `bson:"notifiedAt,omitempty" json:"notifiedAt,omitempty"` // Kept for compatibility, changed to pointer
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}
