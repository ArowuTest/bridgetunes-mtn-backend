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

// Note: PrizeStructure is defined in draw.go
// Note: JackpotHistoryEntry is defined in draw.go


