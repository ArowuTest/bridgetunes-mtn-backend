package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawStatus represents the status of a draw
type DrawStatus string

const (
	DrawStatusScheduled DrawStatus = "SCHEDULED"
	DrawStatusExecuting DrawStatus = "EXECUTING"
	DrawStatusCompleted DrawStatus = "COMPLETED"
	DrawStatusFailed    DrawStatus = "FAILED"
	DrawStatusCancelled DrawStatus = "CANCELLED"
)

// JackpotValidationStatus represents the validation status of a potential jackpot winner
type JackpotValidationStatus string

const (
	JackpotValidationPending        JackpotValidationStatus = "PENDING"
	JackpotValidationValid          JackpotValidationStatus = "VALID"
	JackpotValidationInvalidNotOptIn JackpotValidationStatus = "INVALID_NOT_OPT_IN"
	// Add other invalid reasons as needed (e.g., INVALID_BLACKLISTED)
)

// Draw represents a draw event
type Draw struct {
	ID                        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate                  time.Time          `bson:"drawDate" json:"drawDate"`
	DrawType                  string             `bson:"drawType" json:"drawType"` // e.g., "DAILY", "SATURDAY"
	EligibleDigits            []int              `bson:"eligibleDigits" json:"eligibleDigits"`
	UseDefaultDigits          bool               `bson:"useDefaultDigits" json:"useDefaultDigits"`
	Status                    DrawStatus         `bson:"status" json:"status"`
	Prizes                    []Prize            `bson:"prizes" json:"prizes"` // Uses Prize struct defined in prize.go
	BaseJackpotAmount       float64            `bson:"baseJackpotAmount" json:"baseJackpotAmount"`
	RolloverAmount            float64            `bson:"rolloverAmount" json:"rolloverAmount"` // Rollover amount *into* this draw
	CalculatedJackpotAmount float64            `bson:"calculatedJackpotAmount" json:"calculatedJackpotAmount"`
	TotalParticipants         int                `bson:"totalParticipants" json:"totalParticipants"` // Count for Pool A
	EligibleOptedInParticipants int            `bson:"eligibleOptedInParticipants" json:"eligibleOptedInParticipants"` // Count for Pool B
	JackpotWinnerMsisdn       string             `bson:"jackpotWinnerMsisdn,omitempty" json:"jackpotWinnerMsisdn,omitempty"`
	JackpotWinnerValidationStatus JackpotValidationStatus `bson:"jackpotWinnerValidationStatus,omitempty" json:"jackpotWinnerValidationStatus,omitempty"`
	RolloverExecuted          bool               `bson:"rolloverExecuted" json:"rolloverExecuted"` // Did this draw result in a rollover?
	NumWinners                int                `bson:"numWinners" json:"numWinners"`
	ExecutionStartTime        time.Time          `bson:"executionStartTime,omitempty" json:"executionStartTime,omitempty"`
	ExecutionEndTime          time.Time          `bson:"executionEndTime,omitempty" json:"executionEndTime,omitempty"`
	ExecutionLog              []string           `bson:"executionLog,omitempty" json:"executionLog,omitempty"`
	ErrorMessage              string             `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
	CreatedAt                 time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt                 time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Note: The Prize struct definition has been moved to prize.go
/*
// Prize defines the structure for a single prize category within a draw (REMOVED - Defined in prize.go)
type Prize struct {
	Category   string  `bson:"category" json:"category"`
	Amount     float64 `bson:"amount" json:"amount"`
	NumWinners int     `bson:"numWinners" json:"numWinners"`
}
*/


