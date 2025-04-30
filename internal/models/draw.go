package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Draw represents a draw in the system
// Note: This struct definition might need review based on overall application logic,
// but it includes fields seen in previous versions or implied by errors.
type Draw struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate            time.Time          `bson:"drawDate" json:"drawDate"`
	DrawType            string             `bson:"drawType" json:"drawType"` // DAILY or WEEKLY or SATURDAY?
	EligibleDigits      []int              `bson:"eligibleDigits" json:"eligibleDigits"`
	UseDefault          bool               `bson:"useDefault" json:"useDefault"` // Added based on handler code
	Status              string             `bson:"status" json:"status"` // SCHEDULED, IN_PROGRESS, COMPLETED, CANCELLED, EXECUTED, FAILED?
	TotalParticipants   int                `bson:"totalParticipants,omitempty" json:"totalParticipants,omitempty"` // Added based on potential need
	OptedInParticipants int                `bson:"optedInParticipants,omitempty" json:"optedInParticipants,omitempty"` // Added based on potential need
	Prizes              []Prize            `bson:"prizes" json:"prizes"`
	JackpotAmount       float64            `bson:"jackpotAmount" json:"jackpotAmount"` // Calculated jackpot for this draw
	RolloverSource      []RolloverInfo     `bson:"rolloverSource,omitempty" json:"rolloverSource,omitempty"` // Info about rollovers contributing to this draw
	RolloverTarget      *primitive.ObjectID `bson:"rolloverTarget,omitempty" json:"rolloverTarget,omitempty"` // ID of the draw this jackpot rolled over to (if applicable)
	ExecutionTime       time.Time          `bson:"executionTime,omitempty" json:"executionTime,omitempty"` // Added based on handler code
	ErrorMessage        string             `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"` // Added based on handler code
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Prize represents a prize in a draw
// Includes 'Category' field based on service layer errors
type Prize struct {
	Tier        int     `bson:"tier,omitempty" json:"tier,omitempty"` // Kept from repo version
	Description string  `bson:"description,omitempty" json:"description,omitempty"` // Kept from repo version
	Category    string  `bson:"category" json:"category"` // Added based on service error
	Amount      float64 `bson:"amount" json:"amount"`
	NumWinners  int     `bson:"numWinners,omitempty" json:"numWinners,omitempty"` // Kept from repo version
	WinnerID    primitive.ObjectID `bson:"winnerId,omitempty" json:"winnerId,omitempty"` // Added based on potential need/previous versions
	IsValid     *bool              `bson:"isValid,omitempty" json:"isValid,omitempty"` // Added based on potential need/previous versions
}

// RolloverInfo tracks jackpot rollovers contributing to a draw
// Added based on service error 'undefined: models.RolloverInfo'
type RolloverInfo struct {
	SourceDrawID primitive.ObjectID `bson:"sourceDrawId" json:"sourceDrawId"`
	Amount       float64            `bson:"amount" json:"amount"`
	Reason       string             `bson:"reason" json:"reason"` // e.g., "INVALID_WINNER"
}

// SystemConfig stores system configuration values
// Includes 'UpdatedAt' field based on utils layer errors
type SystemConfig struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key          string             `bson:"key" json:"key"` // e.g., "prize_structure_daily", "prize_structure_weekly"
	Value        interface{}        `bson:"value" json:"value"`
	Description  string             `bson:"description" json:"description"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"` // Changed from LastModified based on utils error
	// LastModified time.Time          `bson:"last_modified" json:"last_modified"` // Original field from repo, commented out
}

// PrizeStructure defines the structure for prizes stored in SystemConfig
// Includes 'Category' and 'Count' fields based on utils layer errors
type PrizeStructure struct {
	Tier        int     `bson:"tier,omitempty" json:"tier,omitempty"` // Kept from repo version
	Description string  `bson:"description,omitempty" json:"description,omitempty"` // Kept from repo version
	Category    string  `bson:"category" json:"category"` // Added based on utils error
	Amount      float64 `bson:"amount" json:"amount"`
	Count       int     `bson:"count" json:"count"` // Added based on utils error (Number of prizes for this category)
	// NumWinners  int     `bson:"num_winners" json:"num_winners"` // Original field from repo, potentially replaced by Count?
}

// JackpotHistory stores records of jackpot amounts over time
// Kept from repo version
type JackpotHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Date      time.Time          `bson:"date" json:"date"`
	Amount    float64            `bson:"amount" json:"amount"`
	DrawType  string             `bson:"draw_type" json:"draw_type"` // DAILY or SATURDAY
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Note: The Winner struct is assumed to be correctly defined in internal/models/winner.go


