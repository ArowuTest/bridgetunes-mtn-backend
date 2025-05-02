package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LoginRequest defines the structure for login requests
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest defines the structure for registration requests
type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"` // Example validation
}

// AdminUser represents a user account for the admin backend (separate from promotion users)
// Note: This is an assumed structure. Adjust fields and validation as needed.
// Consider storing this in a separate MongoDB collection (e.g., "admin_users").
type AdminUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName string             `bson:"firstName" json:"firstName"`
	LastName  string             `bson:"lastName" json:"lastName"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"` // Store hashed password, omit from JSON responses
	Role      string             `bson:"role" json:"role"`       // e.g., "admin", "editor"
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

