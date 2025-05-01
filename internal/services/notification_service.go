package services

import (
	"context"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"github.com/ArowuTest/bridgetunes-mtn-backend/pkg/smsgateway"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LegacyNotificationService handles notification-related business logic // Renamed from NotificationService
type LegacyNotificationService struct {
	notificationRepo repositories.NotificationRepository
	campaignRepo     repositories.CampaignRepository
	 templateRepo     repositories.TemplateRepository
	userRepo         repositories.UserRepository
	mtnGateway       smsgateway.Gateway
	 kodobeGateway    smsgateway.Gateway
	defaultGateway   string
}

// NewLegacyNotificationService creates a new LegacyNotificationService // Renamed from NewNotificationService
func NewLegacyNotificationService(
	notificationRepo repositories.NotificationRepository,
	 templateRepo repositories.TemplateRepository,
	 campaignRepo repositories.CampaignRepository,
	 userRepo repositories.UserRepository,
	 mtnGateway smsgateway.Gateway,
	 kodobeGateway smsgateway.Gateway,
	 defaultGateway string,
) *LegacyNotificationService { // Renamed return type
	return &LegacyNotificationService{ // Renamed struct type
		notificationRepo: notificationRepo,
		 templateRepo:     templateRepo,
		 campaignRepo:     campaignRepo,
		 userRepo:         userRepo,
		 mtnGateway:       mtnGateway,
		 kodobeGateway:    kodobeGateway,
		 defaultGateway:   defaultGateway,
	}
}

// GetNotificationByID retrieves a notification by ID
func (s *LegacyNotificationService) GetNotificationByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error) {
	return s.notificationRepo.FindByID(ctx, id)
}

// GetNotificationsByMSISDN retrieves notifications by MSISDN with pagination
func (s *LegacyNotificationService) GetNotificationsByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.FindByMSISDN(ctx, msisdn, page, limit)
}

// GetNotificationsByCampaignID retrieves notifications by campaign ID with pagination
func (s *LegacyNotificationService) GetNotificationsByCampaignID(ctx context.Context, campaignID primitive.ObjectID, page, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.FindByCampaignID(ctx, campaignID, page, limit)
}

// GetNotificationsByStatus retrieves notifications by status with pagination
func (s *LegacyNotificationService) GetNotificationsByStatus(ctx context.Context, status string, page, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.FindByStatus(ctx, status, page, limit)
}

// SendSMS sends an SMS notification
func (s *LegacyNotificationService) SendSMS(ctx context.Context, msisdn, content, notificationType string, campaignID primitive.ObjectID) (*models.Notification, error) {
	// Create notification record
	notification := &models.Notification{
		MSISDN:     msisdn,
		Content:    content,
		Type:       notificationType,
		Status:     "SENT",
		SentDate:   time.Now(),
		CampaignID: campaignID,
		Gateway:    s.defaultGateway,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Select gateway
	var gateway smsgateway.Gateway
	 if s.defaultGateway == "MTN" {
	 	gateway = s.mtnGateway
	 	notification.Gateway = "MTN"
	 } else {
	 	gateway = s.kodobeGateway
	 	notification.Gateway = "KODOBE"
	 }

	// Send SMS
	 messageID, err := gateway.SendSMS(msisdn, content)
	 if err != nil {
	 	// If MTN gateway fails, try Kodobe as fallback
	 	 if s.defaultGateway == "MTN" {
	 	 	messageID, err = s.kodobeGateway.SendSMS(msisdn, content)
	 	 	 if err != nil {
	 	 	 	notification.Status = "FAILED"
	 	 	 	 s.notificationRepo.Create(ctx, notification)
	 	 	 	 return notification, err
	 	 	 }
	 	 	 notification.Gateway = "KODOBE"
	 	 } else {
	 	 	 notification.Status = "FAILED"
	 	 	 s.notificationRepo.Create(ctx, notification)
	 	 	 return notification, err
	 	 }
	 }

	notification.MessageID = messageID
	 err = s.notificationRepo.Create(ctx, notification)
	 if err != nil {
	 	 return nil, err
	 }

	return notification, nil
}

// CreateCampaign creates a new notification campaign
func (s *LegacyNotificationService) CreateCampaign(ctx context.Context, campaign *models.Campaign) error {
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	return s.campaignRepo.Create(ctx, campaign)
}

// GetAllCampaigns retrieves all campaigns with pagination
func (s *LegacyNotificationService) GetAllCampaigns(ctx context.Context, page, limit int) ([]models.Campaign, error) {
	 return s.campaignRepo.FindAll(ctx, page, limit)
}

// ExecuteCampaign executes a notification campaign
func (s *LegacyNotificationService) ExecuteCampaign(ctx context.Context, campaignID primitive.ObjectID) error {
	// Get the campaign
	 campaign, err := s.campaignRepo.FindByID(ctx, campaignID)
	 if err != nil {
	 	 return err
	 }

	// Check if campaign is already running or completed
	 if campaign.Status == "RUNNING" || campaign.Status == "COMPLETED" {
	 	 return nil
	 }

	// Get the template
	 template, err := s.templateRepo.FindByID(ctx, campaign.TemplateID)
	 if err != nil {
	 	 return err
	 }

	// Get target users based on segment
	 var targetUsers []*models.User
	 // Placeholder: Fetch all opted-in users for now
	 targetUsers, err = s.userRepo.FindByOptInStatus(ctx, true, 1, 10000) // Increased limit for testing

	 if err != nil {
	 	 return err
	 }

	// Update campaign status
	 campaign.Status = "RUNNING"
	 err = s.campaignRepo.Update(ctx, campaign)
	 if err != nil {
	 	 return err
	 }

	// Send notifications to target users
	 var totalSent, delivered, failed int
	 for _, user := range targetUsers {
	 	 if !user.OptInStatus {
	 	 	 continue
	 	 }

	 	 _, sendErr := s.SendSMS(ctx, user.MSISDN, template.Content, template.Type, campaign.ID)
	 	 totalSent++
	 	 if sendErr == nil {
	 	 	 delivered++
	 	 } else {
	 	 	 failed++
	 	 }
	 }

	// Update campaign status and stats
	 campaign.Status = "COMPLETED"
	 // Add stats update here if fields exist in Campaign model
	 return s.campaignRepo.Update(ctx, campaign)
}

// CreateTemplate creates a new notification template
func (s *LegacyNotificationService) CreateTemplate(ctx context.Context, template *models.Template) error {
	 template.CreatedAt = time.Now()
	 template.UpdatedAt = time.Now()
	 return s.templateRepo.Create(ctx, template)
}

// GetTemplateByID retrieves a template by ID
func (s *LegacyNotificationService) GetTemplateByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	 return s.templateRepo.FindByID(ctx, id)
}

// GetTemplateByName retrieves a template by name
func (s *LegacyNotificationService) GetTemplateByName(ctx context.Context, name string) (*models.Template, error) {
	 return s.templateRepo.FindByName(ctx, name)
}

// GetTemplatesByType retrieves templates by type with pagination
func (s *LegacyNotificationService) GetTemplatesByType(ctx context.Context, templateType string, page, limit int) ([]*models.Template, error) {
	 return s.templateRepo.FindByType(ctx, templateType, page, limit)
}

// GetAllTemplates retrieves all templates with pagination
func (s *LegacyNotificationService) GetAllTemplates(ctx context.Context, page, limit int) ([]*models.Template, error) {
	 return s.templateRepo.FindAll(ctx, page, limit)
}

// UpdateTemplate updates a template
func (s *LegacyNotificationService) UpdateTemplate(ctx context.Context, template *models.Template) error {
	 template.UpdatedAt = time.Now()
	 return s.templateRepo.Update(ctx, template)
}

// DeleteTemplate deletes a template
func (s *LegacyNotificationService) DeleteTemplate(ctx context.Context, id primitive.ObjectID) error {
	 return s.templateRepo.Delete(ctx, id)
}

// GetNotificationCount gets the total number of notifications
func (s *LegacyNotificationService) GetNotificationCount(ctx context.Context) (int64, error) {
	 return s.notificationRepo.Count(ctx)
}

// GetCampaignCount gets the total number of campaigns
func (s *LegacyNotificationService) GetCampaignCount(ctx context.Context) (int64, error) {
	 return s.campaignRepo.Count(ctx)
}

// GetTemplateCount gets the total number of templates
func (s *LegacyNotificationService) GetTemplateCount(ctx context.Context) (int64, error) {
	 return s.templateRepo.Count(ctx)
}

