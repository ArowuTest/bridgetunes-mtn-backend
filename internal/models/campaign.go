package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Campaign represents a notification campaign
type Campaign struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	TemplateID  primitive.ObjectID `bson:"templateId" json:"templateId"`
	Segment     Segment            `bson:"segment" json:"segment"`
	Status      string             `bson:"status" json:"status"` // DRAFT, SCHEDULED, RUNNING, COMPLETED, CANCELLED
	ScheduledAt time.Time          `bson:"scheduledAt,omitempty" json:"scheduledAt,omitempty"`
	StartedAt   time.Time          `bson:"startedAt,omitempty" json:"startedAt,omitempty"`
	CompletedAt time.Time          `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	TotalSent   int                `bson:"totalSent" json:"totalSent"`
	Delivered   int                `bson:"delivered" json:"delivered"`
	Failed      int                `bson:"failed" json:"failed"`
	CreatedBy   string             `bson:"createdBy" json:"createdBy"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Segment represents a user segment for targeting notifications
type Segment struct {
	Type       string   `bson:"type" json:"type"` // ALL, OPT_IN, CUSTOM
	Conditions []string `bson:"conditions,omitempty" json:"conditions,omitempty"`
	MSISDNs    []string `bson:"msisdns,omitempty" json:"msisdns,omitempty"`
}
