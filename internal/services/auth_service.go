package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/mongo" // Now used for error checking
	"golang.org/x/crypto/bcrypt"
	// TODO: Add JWT library import (e.g., "github.com/golang-jwt/jwt/v5")
	// import "github.com/golang-jwt/jwt/v5"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.AdminUser, error) // Return AdminUser
	Login(ctx context.Context, req *models.LoginRequest) (string, error)                  // Returns JWT token
}

type authService struct {
	adminUserRepo repositories.AdminUserRepository // Use AdminUserRepository
	// TODO: Add JWT secret key from config
	// jwtSecret string
}

// NewAuthService creates a new AuthService implementation
func NewAuthService(adminUserRepo repositories.AdminUserRepository /*, jwtSecret string*/) AuthService { // Accept AdminUserRepository
	return &authService{
		adminUserRepo: adminUserRepo,
		// jwtSecret: jwtSecret,
	}
}

// Register handles admin user registration
func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AdminUser, error) {
	// Check if admin user already exists
	_, err := s.adminUserRepo.FindByEmail(ctx, req.Email)

	// Handle potential errors from FindByEmail
	// Use a switch statement for clarity
	 switch {
	 case err == nil:
	 	// User found, return error
	 	return nil, errors.New("admin user with this email already exists")
	 case err != mongo.ErrNoDocuments:
	 	// An unexpected error occurred during the database query
	 	return nil, fmt.Errorf("error checking for existing admin user: %w", err)
	 // case err == mongo.ErrNoDocuments:
	 	// User not found, proceed with registration (do nothing here)
	 }

	// Hash password
	 hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	 if err != nil {
	 	return nil, fmt.Errorf("failed to hash password: %w", err)
	 }

	adminUser := &models.AdminUser{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "admin", // Default role for registration, adjust as needed
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save admin user using AdminUserRepository
	newAdminUser, err := s.adminUserRepo.Create(ctx, adminUser)
	 if err != nil {
	 	return nil, fmt.Errorf("failed to create admin user: %w", err)
	 }

	// Don't return password hash
	newAdminUser.Password = ""
	return newAdminUser, nil
}

// Login handles admin user login
func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (string, error) {
	// Find admin user by email using AdminUserRepository
	adminUser, err := s.adminUserRepo.FindByEmail(ctx, req.Email)
	
	// Handle potential errors from FindByEmail
	 switch {
	 case err == mongo.ErrNoDocuments:
	 	// User not found
	 	return "", errors.New("invalid email or password")
	 case err != nil:
	 	// An unexpected error occurred during the database query
	 	return "", fmt.Errorf("error finding admin user: %w", err)
	 // case err == nil:
	 	// User found, proceed with password check (do nothing here)
	 }

	// Compare submitted password with the stored hash
	 err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(req.Password))
	 if err != nil {
	 	// Passwords don't match (or bcrypt error)
	 	return "", errors.New("invalid email or password")
	 }

	// Generate JWT token (Placeholder)
	// TODO: Implement actual JWT generation using a library and secret key
	// claims := jwt.MapClaims{
	// 	 "sub": adminUser.ID.Hex(),
	// 	 "email": adminUser.Email,
	// 	 "role": adminUser.Role,
	// 	 "exp": time.Now().Add(time.Hour * 72).Unix(), // Example expiration
	// }
	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// tokenString, err := token.SignedString([]byte(s.jwtSecret))
	// if err != nil {
	// 	 return "", errors.New("failed to generate token")
	// }
	// return tokenString, nil

	// Return a dummy token for now
	return "dummy-jwt-token-replace-with-real-one", nil
}


