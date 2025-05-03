package models

import (
	"time"
)

// JackpotStatus represents the current status of the jackpot
type JackpotStatus struct {
	CurrentAmount float64   `json:"currentAmount"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	// Add other relevant fields like next draw date if needed
}

