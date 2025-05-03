package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Draw represents a draw in the system, updated based on redesign plan
type Draw struct {
	ID                          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate                    time.Time          `bson:"drawDate" json:"drawDate"`
	DrawType                    string             `bson:"drawType" json:"drawType"` // e.g., "DAILY", "SATURDAY"
	EligibleDigits              []int              `bson:"eligibleDigits" json:"eligibleDigits"`
	UseDefaultDigits            bool               `bson:"useDefaultDigits" json:"useDefaultDigits"`
	Status                      string             `bson:"status" json:"status"` // e.g., "SCHEDULED", "EXECUTING", "COMPLETED", "FAILED"
	TotalParticipants           int                `bson:"totalParticipants,omitempty" json:"totalParticipants,omitempty"` // Count of users in the initial pool (e.g., Jackpot pool)
	EligibleOptedInParticipants int                `bson:"eligibleOptedInParticipants,omitempty" json:"eligibleOptedInParticipants,omitempty"` // Count of users in the consolation pool
	Prizes                      []Prize            `bson:"prizes" json:"prizes"` // Configured prizes for this draw
	BaseJackpotAmount           float64            `bson:"baseJackpotAmount" json:"baseJackpotAmount"` // Base amount for this draw type
	RolloverAmount              float64            `bson:"rolloverAmount" json:"rolloverAmount"` // Amount rolled over *into* this draw
	CalculatedJackpotAmount     float64            `bson:"calculatedJackpotAmount" json:"calculatedJackpotAmount"` // Base + Rollover
	JackpotWinnerMsisdn         string             `bson:"jackpotWinnerMsisdn,omitempty" json:"jackpotWinnerMsisdn,omitempty"` // Initially selected MSISDN for jackpot
	JackpotWinnerValidationStatus string           `bson:"jackpotWinnerValidationStatus,omitempty" json:"jackpotWinnerValidationStatus,omitempty"` // e.g., "PENDING", "VALID", "INVALID_NOT_OPTED_IN"
	RolloverExecuted            bool               `bson:"rolloverExecuted" json:"rolloverExecuted"` // Flag indicating if rollover *from* this draw was processed
	ExecutionStartTime          time.Time          `bson:"executionStartTime,omitempty" json:"executionStartTime,omitempty"`
	ExecutionEndTime            time.Time          `bson:"executionEndTime,omitempty" json:"executionEndTime,omitempty"`
	ExecutionLog                []string           `bson:"executionLog,omitempty" json:"executionLog,omitempty"` // For audit trail
	ErrorMessage                string             `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
	CreatedAt                   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt                   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Prize represents a prize tier within a draw's configuration
type Prize struct {
	Category    string  `bson:"category" json:"category"` // e.g., "JACKPOT", "SECOND", "CONSOLATION"
	Amount      float64 `bson:"amount" json:"amount"`
	NumWinners  int     `bson:"numWinners" json:"numWinners"` // How many winners for this category
	// Winner details are stored in the separate Winner collection, linked by DrawID and Category/Tier
}

// SystemConfig stores system configuration values
type SystemConfig struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Key         string             `bson:"key" json:"key"` // e.g., "prize_structure_DAILY", "base_jackpot_SATURDAY"
	Value       interface{}        `bson:"value" json:"value"` // Can be float64, []PrizeStructure, etc.
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// PrizeStructure defines the structure for prizes stored in SystemConfig
type PrizeStructure struct {
	Category string  `bson:"category" json:"category"`
	Amount   float64 `bson:"amount" json:"amount"`
	Count    int     `bson:"count" json:"count"` // Number of winners for this category
}

// JackpotHistoryEntry stores records of jackpot amounts over time (Consider if needed or if Draw model is sufficient)
type JackpotHistoryEntry struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	DrawDate      time.Time          `bson:"drawDate" json:"drawDate"`
	DrawID        primitive.ObjectID `bson:"drawId" json:"drawId"`
	JackpotAmount float64            `bson:"jackpotAmount" json:"jackpotAmount"` // Calculated amount for the draw
	Won           bool               `bson:"won" json:"won"`
	WinnerMSISDN  string             `bson:"winnerMsisdn,omitempty" json:"winnerMsisdn,omitempty"`
	RolledOver    bool               `bson:"rolledOver" json:"rolledOver"` // Did this jackpot roll over?
}

// Constants for Draw Status and Prize Categories
const (
	DrawStatusScheduled = "SCHEDULED"
	DrawStatusExecuting = "EXECUTING"
	DrawStatusCompleted = "COMPLETED"
	DrawStatusFailed    = "FAILED"

	JackpotValidationPending        = "PENDING"
	JackpotValidationValid          = "VALID"
	JackpotValidationInvalidNotOptIn = "INVALID_NOT_OPTED_IN"

	PrizeCategoryJackpot     = "JACKPOT" // Or "FIRST" depending on convention
	PrizeCategorySecond      = "SECOND"
	PrizeCategoryThird       = "THIRD"
	PrizeCategoryConsolation = "CONSOLATION"
)

