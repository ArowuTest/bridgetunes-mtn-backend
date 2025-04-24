package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Draw represents a draw in the system
type Draw struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate          time.Time          `bson:"drawDate" json:"drawDate"`
	DrawType          string             `bson:"drawType" json:"drawType"` // DAILY or WEEKLY
	EligibleDigits    []int              `bson:"eligibleDigits" json:"eligibleDigits"`
	Status            string             `bson:"status" json:"status"` // SCHEDULED, COMPLETED, CANCELLED
	TotalParticipants int                `bson:"totalParticipants" json:"totalParticipants"`
	Prizes            []Prize            `bson:"prizes" json:"prizes"`
	CreatedAt         time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Prize represents a prize in a draw
type Prize struct {
	Category  string             `bson:"category" json:"category"`
	Amount    float64            `bson:"amount" json:"amount"`
	WinnerID  primitive.ObjectID `bson:"winnerId,omitempty" json:"winnerId,omitempty"`
}
