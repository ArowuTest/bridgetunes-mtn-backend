package services

import (
	"context"
	"errors"
	"fmt"
	"log" // Import the log package
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"github.com/golang-jwt/jwt/v5" // Import JWT library
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.AdminUser, error) // Return AdminUser
	Login(ctx context.Context, req *models.LoginRequest) (string, error)                  // Returns JWT token
}

type authService struct {
	adminUserRepo repositories.AdminUserRepository // Use AdminUserRepository
	jwtSecret     string
	jwtExpiresIn  int
}

// NewAuthService creates a new AuthService implementation
func NewAuthService(adminUserRepo repositories.AdminUserRepository, jwtSecret string, jwtExpiresIn int) AuthService { // Accept AdminUserRepository and JWT config
	return &authService{
		adminUserRepo: adminUserRepo,
		jwtSecret:     jwtSecret,
		jwtExpiresIn:  jwtExpiresIn,
	}
}

// Register handles admin user registration
func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AdminUser, error) {
	log.Printf("[DEBUG] Register: Attempting to register user with email: %s", req.Email)
	// Check if admin user already exists
	_, err := s.adminUserRepo.FindByEmail(ctx, req.Email)

	// Handle potential errors from FindByEmail
	 switch {
	 case err == nil:
	 	log.Printf("[DEBUG] Register: User with email %s already exists.", req.Email)
	 	return nil, errors.New("admin user with this email already exists")
	 case err != mongo.ErrNoDocuments:
	 	log.Printf("[ERROR] Register: Error checking for existing user %s: %v", req.Email, err)
	 	return nil, fmt.Errorf("error checking for existing admin user: %w", err)
	 case err == mongo.ErrNoDocuments:
	 	log.Printf("[DEBUG] Register: User with email %s not found. Proceeding with registration.", req.Email)
	 }

	// Hash password
	 hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	 if err != nil {
	 	log.Printf("[ERROR] Register: Failed to hash password for %s: %v", req.Email, err)
	 	return nil, fmt.Errorf("failed to hash password: %w", err)
	 }
	log.Printf("[DEBUG] Register: Password hashed successfully for %s.", req.Email)

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
	 	log.Printf("[ERROR] Register: Failed to create admin user %s: %v", req.Email, err)
	 	return nil, fmt.Errorf("failed to create admin user: %w", err)
	 }
	log.Printf("[DEBUG] Register: Admin user %s created successfully with ID: %s", req.Email, newAdminUser.ID.Hex())

	// Don't return password hash
	newAdminUser.Password = ""
	return newAdminUser, nil
}

// Login handles admin user login
func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (string, error) {
	log.Printf("[DEBUG] Login: Attempting login for email: %s", req.Email)

	// Find admin user by email using AdminUserRepository
	adminUser, err := s.adminUserRepo.FindByEmail(ctx, req.Email)

	// Handle potential errors from FindByEmail
	 switch {
	 case err == mongo.ErrNoDocuments:
	 	log.Printf("[DEBUG] Login: User not found for email: %s", req.Email)
	 	return "", errors.New("invalid email or password") // Keep generic error for security
	 case err != nil:
	 	log.Printf("[ERROR] Login: Error finding user %s: %v", req.Email, err)
	 	return "", fmt.Errorf("error finding admin user: %w", err)
	 case err == nil:
	 	log.Printf("[DEBUG] Login: User found for email: %s. User ID: %s", req.Email, adminUser.ID.Hex())
	 	log.Printf("[DEBUG] Login: Stored hash for %s: %s", req.Email, adminUser.Password) // Log the stored hash
	 }

	// Compare submitted password with the stored hash
	log.Printf("[DEBUG] Login: Comparing provided password with stored hash for %s", req.Email)
	 err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(req.Password))
	 if err != nil {
	 	// Passwords don't match (or bcrypt error)
	 	log.Printf("[DEBUG] Login: Password comparison failed for %s: %v", req.Email, err) // Log the specific bcrypt error
	 	return "", errors.New("invalid email or password") // Keep generic error for security
	 }

	log.Printf("[DEBUG] Login: Password comparison successful for %s", req.Email)

	// Generate JWT token
	log.Printf("[DEBUG] Login: Generating JWT token for %s", req.Email)
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   adminUser.ID.Hex(), // Subject (user ID)
		"email": adminUser.Email,
		"role":  adminUser.Role,
		"iat":   now.Unix(),                                       // Issued At
		"exp":   now.Add(time.Second * time.Duration(s.jwtExpiresIn)).Unix(), // Expiration Time
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	 if err != nil {
	 	log.Printf("[ERROR] Login: Failed to sign JWT token for %s: %v", req.Email, err)
	 	return "", fmt.Errorf("failed to generate token: %w", err)
	 }

	log.Printf("[DEBUG] Login: JWT token generated successfully for %s", req.Email)
	return tokenString, nil
}


