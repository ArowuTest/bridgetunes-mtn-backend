package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ClaimStatus represents the status of a prize claim
type ClaimStatus string

const (
	ClaimStatusPending   ClaimStatus = "PENDING"   // Winner selected, claim not yet processed
	ClaimStatusProcessing ClaimStatus = "PROCESSING" // Claim is being processed
	ClaimStatusPaid       ClaimStatus = "PAID"       // Prize has been paid/disbursed
	ClaimStatusFailed     ClaimStatus = "FAILED"     // Claim processing failed
	ClaimStatusIneligible ClaimStatus = "INELIGIBLE" // Winner found ineligible post-selection
)

// Winner represents a winning entry in a draw
type Winner struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawID        primitive.ObjectID `bson:"drawId" json:"drawId"`
	UserID        primitive.ObjectID `bson:"userId" json:"userId"` // Added UserID field
	MSISDN        string             `bson:"msisdn" json:"msisdn"`
	PrizeCategory string             `bson:"prizeCategory" json:"prizeCategory"`
	PrizeAmount   float64            `bson:"prizeAmount" json:"prizeAmount"`
	WinDate       time.Time          `bson:"winDate" json:"winDate"`
	ClaimStatus   ClaimStatus        `bson:"claimStatus" json:"claimStatus"`
	ClaimNotes    string             `bson:"claimNotes,omitempty" json:"claimNotes,omitempty"`
	ClaimDate     time.Time          `bson:"claimDate,omitempty" json:"claimDate,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

