package services

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawService defines the interface for draw-related operations
type DrawService interface {
	GetDrawConfig(ctx context.Context, date time.Time) (map[string]interface{}, error)
	GetPrizeStructure(ctx context.Context, drawType string) ([]models.PrizeStructure, error)
	UpdatePrizeStructure(ctx context.Context, drawType string, structure []models.PrizeStructure) error
	ScheduleDraw(ctx context.Context, drawDate time.Time, drawType string, eligibleDigits []int, useDefaultDigits bool) (*models.Draw, error)
	ExecuteDraw(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error)
	GetDrawByID(ctx context.Context, drawID primitive.ObjectID) (*models.Draw, error)
	GetWinnersByDrawID(ctx context.Context, drawID primitive.ObjectID) ([]*models.Winner, error)
	GetDraws(ctx context.Context, startDate, endDate time.Time) ([]*models.Draw, error)
	GetJackpotHistory(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error)
}

// UserService defines the interface for user-related operations (Add other service interfaces as needed)
type UserService interface {
	// Define user service methods here
	// Example (based on LegacyUserService):
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

// TopupService defines the interface for topup-related operations
type TopupService interface {
	// Define topup service methods here
	// Example (based on LegacyTopupService):
	GetTopupByID(ctx context.Context, id primitive.ObjectID) (*models.Topup, error)
	GetTopupsByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Topup, error)
	GetTopupsByDateRange(ctx context.Context, start, end time.Time, page, limit int) ([]*models.Topup, error)
	CreateTopup(ctx context.Context, topup *models.Topup) error
	ProcessTopups(ctx context.Context, startDate, endDate time.Time) (int, error)
	GetTopupCount(ctx context.Context) (int64, error)
}

// NotificationService defines the interface for notification-related operations
type NotificationService interface {
	// Added missing method signatures based on notification_handler.go
	GetNotificationByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error)
	GetNotificationsByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Notification, error)
	GetNotificationsByCampaignID(ctx context.Context, campaignID primitive.ObjectID, page, limit int) ([]*models.Notification, error)
	GetNotificationsByStatus(ctx context.Context, status string, page, limit int) ([]*models.Notification, error)
	SendSMS(ctx context.Context, msisdn, content, notificationType string, campaignID primitive.ObjectID) (*models.Notification, error)
	CreateCampaign(ctx context.Context, campaign *models.Campaign) error
	ExecuteCampaign(ctx context.Context, campaignID primitive.ObjectID) error
	GetAllCampaigns(ctx context.Context, page, limit int) ([]models.Campaign, error)
	CreateTemplate(ctx context.Context, template *models.Template) error
	GetTemplateByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error)
	GetTemplateByName(ctx context.Context, name string) (*models.Template, error)
	GetTemplatesByType(ctx context.Context, templateType string, page, limit int) ([]*models.Template, error)
	GetAllTemplates(ctx context.Context, page, limit int) ([]*models.Template, error)
	UpdateTemplate(ctx context.Context, template *models.Template) error
	DeleteTemplate(ctx context.Context, id primitive.ObjectID) error
	GetNotificationCount(ctx context.Context) (int64, error)
	GetCampaignCount(ctx context.Context) (int64, error)
	GetTemplateCount(ctx context.Context) (int64, error)
}

