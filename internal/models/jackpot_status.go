package models

import (
	"time"
)

// JackpotStatus represents the current status of the jackpot
type JackpotStatus struct {
	CurrentAmount    float64   `json:"currentAmount"`
	LastDrawDate     time.Time `json:"lastDrawDate"` // Date of the last relevant draw
	LastWinnerMSISDN string    `json:"lastWinnerMsisdn,omitempty"` // MSISDN of the last jackpot winner (if any)
	LastWinAmount    float64   `json:"lastWinAmount,omitempty"`    // Amount won by the last jackpot winner (if any)
	LastUpdatedAt    time.Time `json:"lastUpdatedAt"`
	// Add other relevant fields like next draw date if needed
}

