package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BlacklistEntry represents an entry in the MSISDN blacklist
type BlacklistEntry struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN    string             `bson:"msisdn" json:"msisdn"`
	Reason    string             `bson:"reason,omitempty" json:"reason,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
