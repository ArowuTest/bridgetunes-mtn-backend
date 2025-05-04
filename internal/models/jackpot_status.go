package models

import (
	"time"
)

// JackpotStatus represents the current status of the jackpot
type JackpotStatus struct {
	CurrentAmount    float64   `json:"currentAmount"`
	LastDrawDate     time.Time `json:"lastDrawDate"` // Date of the last relevant draw
	LastWinnerMSISDN string    `json:"lastWinnerMsisdn,omitempty"` // Added: MSISDN of the last jackpot winner (if any)
	LastWinAmount    float64   `json:"lastWinAmount,omitempty"`    // Added: Amount won by the last jackpot winner (if any)
	NextDrawDate     time.Time `json:"nextDrawDate,omitempty"`    // Added: Date of the next scheduled draw
	LastUpdatedAt    time.Time `json:"lastUpdatedAt"`
}


