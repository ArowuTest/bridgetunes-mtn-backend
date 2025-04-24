package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN       string             `bson:"msisdn" json:"msisdn"`
	OptInStatus  bool               `bson:"optInStatus" json:"optInStatus"`
	OptInDate    time.Time          `bson:"optInDate,omitempty" json:"optInDate,omitempty"`
	OptInChannel string             `bson:"optInChannel,omitempty" json:"optInChannel,omitempty"`
	OptOutDate   time.Time          `bson:"optOutDate,omitempty" json:"optOutDate,omitempty"`
	Points       int                `bson:"points" json:"points"`
	IsBlacklisted bool              `bson:"isBlacklisted" json:"isBlacklisted"`
	LastActivity time.Time          `bson:"lastActivity,omitempty" json:"lastActivity,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}
