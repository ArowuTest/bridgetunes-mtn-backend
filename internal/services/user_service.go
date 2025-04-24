package services

import (
	"context"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// GetUserByMSISDN retrieves a user by MSISDN
func (s *UserService) GetUserByMSISDN(ctx context.Context, msisdn string) (*models.User, error) {
	return s.userRepo.FindByMSISDN(ctx, msisdn)
}

// GetAllUsers retrieves all users with pagination
func (s *UserService) GetAllUsers(ctx context.Context, page, limit int) ([]*models.User, error) {
	return s.userRepo.FindAll(ctx, page, limit)
}

// GetUsersByOptInStatus retrieves users by opt-in status with pagination
func (s *UserService) GetUsersByOptInStatus(ctx context.Context, optInStatus bool, page, limit int) ([]*models.User, error) {
	return s.userRepo.FindByOptInStatus(ctx, optInStatus, page, limit)
}

// GetUsersByEligibleDigits retrieves users by eligible digits (last digits of MSISDN)
func (s *UserService) GetUsersByEligibleDigits(ctx context.Context, digits []int, optInStatus bool) ([]*models.User, error) {
	return s.userRepo.FindByEligibleDigits(ctx, digits, optInStatus)
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return s.userRepo.Create(ctx, user)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	return s.userRepo.Delete(ctx, id)
}

// OptIn opts a user in to the promotion
func (s *UserService) OptIn(ctx context.Context, msisdn string, channel string) error {
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
func (s *UserService) OptOut(ctx context.Context, msisdn string) error {
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
func (s *UserService) AddPoints(ctx context.Context, msisdn string, points int) error {
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return err
	}

	user.Points += points
	user.LastActivity = time.Now()
	return s.userRepo.Update(ctx, user)
}

// GetUserCount gets the total number of users
func (s *UserService) GetUserCount(ctx context.Context) (int64, error) {
	return s.userRepo.Count(ctx)
}
