package services

import (
	"context"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"golang.org/x/crypto/bcrypt"
	// TODO: Add JWT library import (e.g., "github.com/golang-jwt/jwt/v5")
	"errors"
	"time"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req *models.LoginRequest) (string, error) // Returns JWT token
}

type authService struct {
	userRepo repository.UserRepository
	// TODO: Add JWT secret key from config
	// jwtSecret string
}

// NewAuthService creates a new AuthService implementation
func NewAuthService(userRepo repository.UserRepository /*, jwtSecret string*/) AuthService {
	return &authService{
		userRepo: userRepo,
		// jwtSecret: jwtSecret,
	}
}

// Register handles user registration
func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	_, err := s.userRepo.FindByEmail(ctx, req.Email)
	// Note: This assumes FindByEmail returns a specific error for "not found"
	// If it returns (nil, nil) for not found, the logic needs adjustment.
	// If it returns (nil, someOtherError), we should return that error.
	// For now, assume any error means we can't proceed, or user exists.
	// A more robust check would be needed here based on repository behavior.
	// if err == nil { // Assuming nil error means user found
	// 	 return nil, errors.New("user with this email already exists")
	// }
	// TODO: Add proper error checking based on userRepo.FindByEmail behavior

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "user", // Default role, adjust as needed
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save user
	newUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// Don't return password hash
	newUser.Password = ""
	return newUser, nil
}

// Login handles user login
func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Handle user not found or other errors
		return "", errors.New("invalid credentials")
	}

	// Compare password
	// TODO: Ensure user.Password is the hashed password from the DB
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	// if err != nil {
	// 	 return "", errors.New("invalid credentials")
	// }
	// Placeholder check:
	 if user.Password == "" { // Should be checking hash
	 	return "", errors.New("password check not implemented")
	 }

	// Generate JWT token (Placeholder)
	// TODO: Implement actual JWT generation using a library and secret key
	// claims := jwt.MapClaims{
	// 	 "sub": user.ID.Hex(),
	// 	 "email": user.Email,
	// 	 "role": user.Role,
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


