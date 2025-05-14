package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SystemSettings represents system-wide configuration settings
type SystemSettings struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SMSGateway  string            `bson:"smsGateway" json:"smsGateway"` // MTN, KODOBE, UDUX
	CreatedAt   time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time         `bson:"updatedAt" json:"updatedAt"`
	UpdatedBy   string            `bson:"updatedBy" json:"updatedBy"`
} 