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

// --- ADDED CODE START ---
// Campaign represents a notification campaign
type Campaign struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name" binding:"required"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	TemplateID  primitive.ObjectID `bson:"templateId" json:"templateId" binding:"required"`
	// Add other relevant fields based on your requirements, e.g.:
	// TargetSegmentID primitive.ObjectID `bson:"targetSegmentId,omitempty" json:"targetSegmentId,omitempty"`
	// ScheduleType string             `bson:"scheduleType" json:"scheduleType"` // e.g., IMMEDIATE, SCHEDULED
	// ScheduledTime time.Time          `bson:"scheduledTime,omitempty" json:"scheduledTime,omitempty"`
	Status      string             `bson:"status" json:"status"` // e.g., DRAFT, SCHEDULED, EXECUTING, COMPLETED, FAILED
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
	ExecutedAt  time.Time          `bson:"executedAt,omitempty" json:"executedAt,omitempty"`
}

// Template represents a notification message template
type Template struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name" binding:"required"`
	Content   string             `bson:"content" json:"content" binding:"required"`
	Type      string             `bson:"type" json:"type" binding:"required"` // SMS, PUSH, EMAIL
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// --- ADDED CODE END ---


