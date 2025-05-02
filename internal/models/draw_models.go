package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawConfig represents the configuration for a draw on a specific date.
// Placeholder definition.
type DrawConfig struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Date           time.Time          `bson:"date" json:"date"`
	DrawType       string             `bson:"draw_type" json:"draw_type"` // e.g., DAILY, SATURDAY
	EligibleDigits []int              `bson:"eligible_digits" json:"eligible_digits"`
	UseDefault     bool               `bson:"use_default" json:"use_default"`
	// Add other config fields as needed
}

// JackpotHistory represents a record in the jackpot history.
// Placeholder definition.
type JackpotHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate  time.Time          `bson:"draw_date" json:"draw_date"`
	DrawType  string             `bson:"draw_type" json:"draw_type"`
	Amount    float64            `bson:"amount" json:"amount"`
	WinnerID  primitive.ObjectID `bson:"winner_id,omitempty" json:"winner_id,omitempty"` // Optional: Link to winner if won
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// PrizeStructure defines the structure of prizes for a specific draw type.
// Assuming this was already defined somewhere, but adding here for completeness if not.
type PrizeStructure struct {
	Category string  `bson:"category" json:"category"` // e.g., "FIRST", "SECOND"
	Amount   float64 `bson:"amount" json:"amount"`
	// Add other fields like quantity if needed
}

