package services

import (
	"context"
	"errors"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/mongo"
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
	// If err is nil, it means the user was found
	if err == nil {
		return nil, errors.New("admin user with this email already exists")
	}
	// If the error is something other than "not found", return it
	// Assuming the repository implementation returns mongo.ErrNoDocuments when not found
	// (This should be confirmed in the actual repository implementation)
	// if err != nil && err != mongo.ErrNoDocuments {
	// 	 return nil, fmt.Errorf("error checking for existing admin user: %w", err)
	// }
	// Simplified check for now: if any error other than expected 'not found', fail.
	// A more robust implementation would explicitly check for mongo.ErrNoDocuments.
	// For now, we proceed if err is not nil (implying user not found or another error we ignore for now)

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	 if err != nil {
	 	return nil, errors.New("failed to hash password")
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
	 	return nil, errors.New("failed to create admin user")
	 }

	// Don't return password hash
	newAdminUser.Password = ""
	return newAdminUser, nil
}

// Login handles admin user login
func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (string, error) {
	// Find admin user by email using AdminUserRepository
	adminUser, err := s.adminUserRepo.FindByEmail(ctx, req.Email)
	 if err != nil {
	 	// Handle user not found (assuming mongo.ErrNoDocuments) or other errors
	 	// if err == mongo.ErrNoDocuments {
	 	// 	 return "", errors.New("invalid email or password")
	 	// } else {
	 	// 	 return "", fmt.Errorf("error finding admin user: %w", err)
	 	// }
	 	// Simplified error handling for now
	 	return "", errors.New("invalid email or password")
	 }

	// Compare submitted password with the stored hash
	 err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(req.Password))
	 if err != nil {
	 	// Passwords don't match
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

