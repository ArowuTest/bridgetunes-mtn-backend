package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a notification sent to a user
type Notification struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MSISDN       string             `bson:"msisdn" json:"msisdn"`
	Content      string             `bson:"content" json:"content"`
	Type         string             `bson:"type" json:"type"` // WINNER, TOPUP, OPT_IN, etc.
	Status       string             `bson:"status" json:"status"` // SENT, DELIVERED, FAILED
	SentDate     time.Time          `bson:"sentDate" json:"sentDate"`
	DeliveryDate time.Time          `bson:"deliveryDate,omitempty" json:"deliveryDate,omitempty"`
	CampaignID   primitive.ObjectID `bson:"campaignId,omitempty" json:"campaignId,omitempty"`
	Gateway      string             `bson:"gateway" json:"gateway"` // MTN, KODOBE
	MessageID    string             `bson:"messageId,omitempty" json:"messageId,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}
