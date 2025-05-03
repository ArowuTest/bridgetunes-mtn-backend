package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JackpotRollover records the details when a jackpot amount is rolled over
// from one draw to another, typically due to an invalid winner.
type JackpotRollover struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SourceDrawID        primitive.ObjectID `bson:"sourceDrawId" json:"sourceDrawId"` // The draw where the jackpot was initially won (but invalid)
	SourceDrawDate      time.Time          `bson:"sourceDrawDate" json:"sourceDrawDate"`
	RolloverAmount      float64            `bson:"rolloverAmount" json:"rolloverAmount"` // The amount being rolled over
	DestinationDrawID   primitive.ObjectID `bson:"destinationDrawId,omitempty" json:"destinationDrawId,omitempty"` // The draw the amount is rolled over *to* (e.g., next Saturday)
	DestinationDrawDate time.Time          `bson:"destinationDrawDate" json:"destinationDrawDate"`
	Reason              string             `bson:"reason" json:"reason"` // e.g., "INVALID_WINNER_NOT_OPTED_IN"
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"` // Timestamp when the rollover was recorded
}

