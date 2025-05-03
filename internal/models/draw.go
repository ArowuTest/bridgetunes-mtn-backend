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
)

// Prize defines the structure for a prize category within a draw
type Prize struct {
	Category   string  `bson:"category" json:"category"`     // e.g., "JACKPOT", "2ND_PRIZE", "CONSOLATION"
	Amount     float64 `bson:"amount" json:"amount"`         // Prize amount
	NumWinners int     `bson:"numWinners" json:"numWinners"` // Number of winners for this category
}

// JackpotValidationStatus represents the validation status of a potential jackpot winner
type JackpotValidationStatus string

const (
	JackpotValidationPending        JackpotValidationStatus = "PENDING"
	JackpotValidationValid          JackpotValidationStatus = "VALID"
	JackpotValidationInvalidNotOptIn JackpotValidationStatus = "INVALID_NOT_OPT_IN"
	// Add other invalid reasons as needed
)

// Define Prize Categories as constants for consistency
const (
	JackpotCategory     string = "JACKPOT"
	SecondPrizeCategory string = "2ND_PRIZE"
	ThirdPrizeCategory  string = "3RD_PRIZE"
	ConsolationCategory string = "CONSOLATION"
)

// Draw represents a draw event in the database
type Draw struct {
	ID                        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate                  time.Time          `bson:"drawDate" json:"drawDate"`
	DrawType                  string             `bson:"drawType" json:"drawType"` // e.g., "DAILY", "SATURDAY"
	EligibleDigits            []int              `bson:"eligibleDigits" json:"eligibleDigits"`
	UseDefaultDigits          bool               `bson:"useDefaultDigits" json:"useDefaultDigits"`
	Status                    DrawStatus         `bson:"status" json:"status"`
	Prizes                    []Prize            `bson:"prizes" json:"prizes"` // Embed prize structure
	BaseJackpotAmount         float64            `bson:"baseJackpotAmount" json:"baseJackpotAmount"`
	RolloverAmount            float64            `bson:"rolloverAmount" json:"rolloverAmount"` // Rollover amount *into* this draw
	CalculatedJackpotAmount   float64            `bson:"calculatedJackpotAmount" json:"calculatedJackpotAmount"`
	JackpotWinnerMsisdn       string             `bson:"jackpotWinnerMsisdn,omitempty" json:"jackpotWinnerMsisdn,omitempty"`
	JackpotWinnerValidationStatus JackpotValidationStatus `bson:"jackpotWinnerValidationStatus,omitempty" json:"jackpotWinnerValidationStatus,omitempty"`
	RolloverExecuted          bool               `bson:"rolloverExecuted" json:"rolloverExecuted"` // Flag if rollover *from* this draw was triggered
	ExecutionStartTime        time.Time          `bson:"executionStartTime,omitempty" json:"executionStartTime,omitempty"`
	ExecutionEndTime          time.Time          `bson:"executionEndTime,omitempty" json:"executionEndTime,omitempty"`
	ExecutionLog              []string           `bson:"executionLog,omitempty" json:"executionLog,omitempty"`
	ErrorMessage              string             `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
	TotalParticipants         int                `bson:"totalParticipants,omitempty" json:"totalParticipants,omitempty"`           // Pool A count
	EligibleOptedInParticipants int              `bson:"eligibleOptedInParticipants,omitempty" json:"eligibleOptedInParticipants,omitempty"` // Pool B count
	NumWinners                int                `bson:"numWinners,omitempty" json:"numWinners,omitempty"`                     // Total winners created for this draw
	CreatedAt                 time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt                 time.Time          `bson:"updatedAt" json:"updatedAt"`
}


