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

// NOTE: The Winner struct definition has been removed from this file.
// It should be defined in internal/models/winner.go


