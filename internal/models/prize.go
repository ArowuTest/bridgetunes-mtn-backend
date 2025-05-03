package models

// Prize defines the structure for a single prize category within a draw
type Prize struct {
	Category   string  `bson:"category" json:"category"`     // e.g., "JACKPOT", "2ND_PRIZE", "CONSOLATION"
	Amount     float64 `bson:"amount" json:"amount"`         // Prize amount for this category
	NumWinners int     `bson:"numWinners" json:"numWinners"` // Number of winners for this category
}
