package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Blacklist represents a blacklisted user
type Blacklist struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN      string             `bson:"msisdn" json:"msisdn"`
	Reason      string             `bson:"reason" json:"reason"`
	BlacklistedAt time.Time        `bson:"blacklistedAt" json:"blacklistedAt"`
	BlacklistedBy string           `bson:"blacklistedBy" json:"blacklistedBy"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
