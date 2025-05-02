package services

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LegacyUserService handles user-related business logic // Renamed from UserService
type LegacyUserService struct {
	 userRepo repositories.UserRepository
}

// NewLegacyUserService creates a new LegacyUserService // Renamed from NewUserService
func NewLegacyUserService(userRepo repositories.UserRepository) *LegacyUserService { // Renamed return type
	return &LegacyUserService{ // Renamed struct type
		 userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by ID
func (s *LegacyUserService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// GetUserByMSISDN retrieves a user by MSISDN
func (s *LegacyUserService) GetUserByMSISDN(ctx context.Context, msisdn string) (*models.User, error) {
	return s.userRepo.FindByMSISDN(ctx, msisdn)
}

// GetAllUsers retrieves all users with pagination
func (s *LegacyUserService) GetAllUsers(ctx context.Context, page, limit int) ([]*models.User, error) {
	return s.userRepo.FindAll(ctx, page, limit)
}

// GetUsersByOptInStatus retrieves users by opt-in status with pagination
func (s *LegacyUserService) GetUsersByOptInStatus(ctx context.Context, optInStatus bool, page, limit int) ([]*models.User, error) {
	return s.userRepo.FindByOptInStatus(ctx, optInStatus, page, limit)
}

// GetUsersByEligibleDigits retrieves users by eligible digits (last digits of MSISDN)
func (s *LegacyUserService) GetUsersByEligibleDigits(ctx context.Context, digits []int, optInStatus bool) ([]*models.User, error) {
	return s.userRepo.FindByEligibleDigits(ctx, digits, optInStatus)
}

// CreateUser creates a new user
func (s *LegacyUserService) CreateUser(ctx context.Context, user *models.User) error {
	 user.CreatedAt = time.Now()
	 user.UpdatedAt = time.Now()
	 return s.userRepo.Create(ctx, user)
}

// UpdateUser updates a user
func (s *LegacyUserService) UpdateUser(ctx context.Context, user *models.User) error {
	 user.UpdatedAt = time.Now()
	 return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *LegacyUserService) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	 return s.userRepo.Delete(ctx, id)
}

// OptIn opts a user in to the promotion
func (s *LegacyUserService) OptIn(ctx context.Context, msisdn string, channel string) error {
	 user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	 if err != nil {
		 // User doesn't exist, create a new one
		 user = &models.User{
			 MSISDN:       msisdn,
			 OptInStatus:  true,
			 OptInDate:    time.Now(),
			 OptInChannel: channel,
			 Points:       0,
			 IsBlacklisted: false,
			 LastActivity: time.Now(),
		 }
		 return s.userRepo.Create(ctx, user)
	 }

	 // User exists, update opt-in status
	 user.OptInStatus = true
	 user.OptInDate = time.Now()
	 user.OptInChannel = channel
	 user.LastActivity = time.Now()
	 return s.userRepo.Update(ctx, user)
}

// OptOut opts a user out of the promotion
func (s *LegacyUserService) OptOut(ctx context.Context, msisdn string) error {
	 user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	 if err != nil {
		 return err
	 }

	 user.OptInStatus = false
	 user.OptOutDate = time.Now()
	 user.LastActivity = time.Now()
	 return s.userRepo.Update(ctx, user)
}

// AddPoints adds points to a user
func (s *LegacyUserService) AddPoints(ctx context.Context, msisdn string, points int) error {
	 user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	 if err != nil {
		 return err
	 }

	 user.Points += points
	 user.LastActivity = time.Now()
	 return s.userRepo.Update(ctx, user)
}

// GetUserCount gets the total number of users
func (s *LegacyUserService) GetUserCount(ctx context.Context) (int64, error) {
	return s.userRepo.Count(ctx)
}


// UserService defines the interface for user-related operations (placeholder)
type UserService interface {
	GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetUserByMSISDN(ctx context.Context, msisdn string) (*models.User, error)
	GetAllUsers(ctx context.Context, page, limit int) ([]*models.User, error)
	GetUsersByOptInStatus(ctx context.Context, optInStatus bool, page, limit int) ([]*models.User, error)
	GetUsersByEligibleDigits(ctx context.Context, digits []int, optInStatus bool) ([]*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id primitive.ObjectID) error
	OptIn(ctx context.Context, msisdn string, channel string) error
	OptOut(ctx context.Context, msisdn string) error
	AddPoints(ctx context.Context, msisdn string, points int) error
	GetUserCount(ctx context.Context) (int64, error)
}

// NewUserService is a wrapper to maintain compatibility with main.go
// It returns the LegacyUserService implementation, cast to the UserService interface.
// NOTE: This assumes LegacyUserService implements the UserService interface.
func NewUserService(userRepo repositories.UserRepository) UserService {
	// LegacyUserService only requires userRepo, which main.go provides.
	 return NewLegacyUserService(userRepo)
}

