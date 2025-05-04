package models

import (
	"time"
)

// JackpotStatus represents the current status of the jackpot
type JackpotStatus struct {
	CurrentAmount float64   `json:"currentAmount"`
	LastDrawDate  time.Time `json:"lastDrawDate"` // Added field for the date of the last relevant draw
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	// Add other relevant fields like next draw date if needed
}

