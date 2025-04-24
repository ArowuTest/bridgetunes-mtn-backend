package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Template represents a notification template
type Template struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Content     string             `bson:"content" json:"content"`
	Type        string             `bson:"type" json:"type"` // WINNER, TOPUP, OPT_IN, etc.
	Variables   []string           `bson:"variables" json:"variables"`
	IsActive    bool               `bson:"isActive" json:"isActive"`
	CreatedBy   string             `bson:"createdBy" json:"createdBy"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
