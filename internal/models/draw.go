package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Draw represents a scheduled draw event
type Draw struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DrawDate       time.Time          `json:"draw_date" bson:"draw_date"`
	DrawType       string             `json:"draw_type" bson:"draw_type"` // e.g., DAILY, SATURDAY
	EligibleDigits []int              `json:"eligible_digits" bson:"eligible_digits"`
	UseDefault     bool               `json:"use_default" bson:"use_default"`
	Prizes         []Prize            `json:"prizes" bson:"prizes"`
	JackpotAmount  float64            `json:"jackpot_amount" bson:"jackpot_amount"`
	Status         string             `json:"status" bson:"status"` // e.g., SCHEDULED, EXECUTED, FAILED
	ExecutionTime  time.Time          `json:"execution_time,omitempty" bson:"execution_time,omitempty"`
	// Winners field removed as Winner struct is defined in winner.go
	// Winners        []Winner           `json:"winners,omitempty" bson:"winners,omitempty"` 
	ErrorMessage   string             `json:"error_message,omitempty" bson:"error_message,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

// Prize represents a prize tier within a draw
type Prize struct {
	Tier        int     `json:"tier" bson:"tier"`
	Description string  `json:"description" bson:"description"`
	Amount      float64 `json:"amount" bson:"amount"`
	NumWinners  int     `json:"num_winners" bson:"num_winners"` // How many winners for this tier
}

/* Winner struct definition removed - should be defined in winner.go
// Winner represents a winner of a specific prize in a draw
type Winner struct {
	Msisdn      string             `json:"msisdn" bson:"msisdn"`
	PrizeTier   int                `json:"prize_tier" bson:"prize_tier"`
	PrizeAmount float64            `json:"prize_amount" bson:"prize_amount"`
	DrawID      primitive.ObjectID `json:"draw_id" bson:"draw_id"`
	DrawDate    time.Time          `json:"draw_date" bson:"draw_date"`
	Notified    bool               `json:"notified" bson:"notified"`
	Paid        bool               `json:"paid" bson:"paid"`
	PaymentRef  string             `json:"payment_ref,omitempty" bson:"payment_ref,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}
*/

// SystemConfig stores system-wide configurations like prize structures
type SystemConfig struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Key          string             `json:"key" bson:"key"` // e.g., "prize_structure_daily", "prize_structure_saturday"
	Value        interface{}        `json:"value" bson:"value"` // Can store different types of config, like []PrizeStructure
	Description  string             `json:"description" bson:"description"`
	LastModified time.Time          `json:"last_modified" bson:"last_modified"`
}

// JackpotHistory stores records of jackpot amounts over time
type JackpotHistory struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Date      time.Time          `json:"date" bson:"date"`
	Amount    float64            `json:"amount" bson:"amount"`
	DrawType  string             `json:"draw_type" bson:"draw_type"` // DAILY or SATURDAY
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// PrizeStructure defines the structure for a specific prize tier
// Restored here
type PrizeStructure struct {
	Tier        int     `json:"tier" bson:"tier" binding:"required"`
	Description string  `json:"description" bson:"description" binding:"required"`
	Amount      float64 `json:"amount" bson:"amount" binding:"required"`
	NumWinners  int     `json:"num_winners" bson:"num_winners" binding:"required"`
}

