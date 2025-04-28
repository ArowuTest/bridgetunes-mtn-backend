package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Draw represents a draw in the system
type Draw struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate            time.Time          `bson:"drawDate" json:"drawDate"`
	DrawType            string             `bson:"drawType" json:"drawType"` // DAILY or WEEKLY
	EligibleDigits      []int              `bson:"eligibleDigits" json:"eligibleDigits"`
	Status              string             `bson:"status" json:"status"` // SCHEDULED, IN_PROGRESS, COMPLETED, CANCELLED
	TotalParticipants   int                `bson:"totalParticipants" json:"totalParticipants"`
	OptedInParticipants int                `bson:"optedInParticipants" json:"optedInParticipants"` // Participants in the consolation prize pool
	Prizes              []Prize            `bson:"prizes" json:"prizes"`
	JackpotAmount       float64            `bson:"jackpotAmount" json:"jackpotAmount"` // Calculated jackpot for this draw
	RolloverSource      []RolloverInfo     `bson:"rolloverSource,omitempty" json:"rolloverSource,omitempty"` // Info about rollovers contributing to this draw
	RolloverTarget      *primitive.ObjectID `bson:"rolloverTarget,omitempty" json:"rolloverTarget,omitempty"` // ID of the draw this jackpot rolled over to (if applicable)
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Prize represents a prize in a draw
type Prize struct {
	Category  string             `bson:"category" json:"category"`
	Amount    float64            `bson:"amount" json:"amount"`
	WinnerID  primitive.ObjectID `bson:"winnerId,omitempty" json:"winnerId,omitempty"`
	IsValid   *bool              `bson:"isValid,omitempty" json:"isValid,omitempty"` // Pointer to bool to distinguish between false and not set
}

// RolloverInfo tracks jackpot rollovers contributing to a draw
type RolloverInfo struct {
	SourceDrawID primitive.ObjectID `bson:"sourceDrawId" json:"sourceDrawId"`
	Amount       float64            `bson:"amount" json:"amount"`
	Reason       string             `bson:"reason" json:"reason"` // e.g., "INVALID_WINNER"
}

// Winner represents a winner in a draw
type Winner struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN         string             `bson:"msisdn" json:"msisdn"`
	MaskedMSISDN   string             `bson:"maskedMsisdn" json:"maskedMsisdn"` // e.g., "080*****78"
	DrawID         primitive.ObjectID `bson:"drawId" json:"drawId"`
	PrizeCategory  string             `bson:"prizeCategory" json:"prizeCategory"`
	PrizeAmount    float64            `bson:"prizeAmount" json:"prizeAmount"`
	IsOptedIn      bool               `bson:"isOptedIn" json:"isOptedIn"` // User's opt-in status at time of draw
	IsValid        bool               `bson:"isValid" json:"isValid"` // Whether the win is valid (e.g., opted-in for jackpot)
	Points         int                `bson:"points" json:"points"` // User's points at time of draw
	WinDate        time.Time          `bson:"winDate" json:"winDate"`
	ClaimStatus    string             `bson:"claimStatus" json:"claimStatus"` // PENDING, CLAIMED, EXPIRED
	ClaimDate      *time.Time         `bson:"claimDate,omitempty" json:"claimDate,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// SystemConfig stores system configuration values
type SystemConfig struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key         string             `bson:"key" json:"key"` // e.g., "prizeStructureDaily", "prizeStructureWeekly"
	Value       interface{}        `bson:"value" json:"value"`
	Description string             `bson:"description" json:"description"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// PrizeStructure defines the structure for prizes stored in SystemConfig
type PrizeStructure struct {
	Category string  `bson:"category" json:"category"`
	Amount   float64 `bson:"amount" json:"amount"`
	Count    int     `bson:"count" json:"count"` // Number of prizes for this category (e.g., 7 for consolation)
}


