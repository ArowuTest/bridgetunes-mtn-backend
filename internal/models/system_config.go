package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SystemConfig represents a configuration setting stored in the database
type SystemConfig struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key         string             `bson:"key" json:"key"` // Unique key for the config setting (e.g., "prize_structure_DAILY")
	Value       interface{}        `bson:"value" json:"value"` // Can store various types (string, number, bool, array, object)
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
