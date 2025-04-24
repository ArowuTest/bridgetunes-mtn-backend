package services

import (
	"context"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"github.com/bridgetunes/mtn-backend/pkg/smsgateway"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationService handles notification-related business logic
type NotificationService struct {
	notificationRepo repositories.NotificationRepository
	templateRepo     repositories.TemplateRepository
	campaignRepo     repositories.CampaignRepository
	userRepo         repositories.UserRepository
	mtnGateway       smsgateway.Gateway
	kodobeGateway    smsgateway.Gateway
	defaultGateway   string
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	templateRepo repositories.TemplateRepository,
	campaignRepo repositories.CampaignRepository,
	userRepo repositories.UserRepository,
	mtnGateway smsgateway.Gateway,
	kodobeGateway smsgateway.Gateway,
	defaultGateway string,
) *NotificationService {
	return &NotificationService{
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
func (s *NotificationService) GetNotificationByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error) {
	return s.notificationRepo.FindByID(ctx, id)
}

// GetNotificationsByMSISDN retrieves notifications by MSISDN with pagination
func (s *NotificationService) GetNotificationsByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.FindByMSISDN(ctx, msisdn, page, limit)
}

// GetNotificationsByCampaignID retrieves notifications by campaign ID with pagination
func (s *NotificationService) GetNotificationsByCampaignID(ctx context.Context, campaignID primitive.ObjectID, page, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.FindByCampaignID(ctx, campaignID, page, limit)
}

// GetNotificationsByStatus retrieves notifications by status with pagination
func (s *NotificationService) GetNotificationsByStatus(ctx context.Context, status string, page, limit int) ([]*models.Notification, error) {
	return s.notificationRepo.FindByStatus(ctx, status, page, limit)
}

// SendSMS sends an SMS notification
func (s *NotificationService) SendSMS(ctx context.Context, msisdn, content, notificationType string, campaignID primitive.ObjectID) (*models.Notification, error) {
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
func (s *NotificationService) CreateCampaign(ctx context.Context, campaign *models.Campaign) error {
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	return s.campaignRepo.Create(ctx, campaign)
}

// ExecuteCampaign executes a notification campaign
func (s *NotificationService) ExecuteCampaign(ctx context.Context, campaignID primitive.ObjectID) error {
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
	if campaign.Segment.Type == "ALL" {
		targetUsers, err = s.userRepo.FindAll(ctx, 1, 1000) // Get all users (paginated)
	} else if campaign.Segment.Type == "OPT_IN" {
		targetUsers, err = s.userRepo.FindByOptInStatus(ctx, true, 1, 1000) // Get opted-in users
	} else if campaign.Segment.Type == "CUSTOM" && len(campaign.Segment.MSISDNs) > 0 {
		// Get users by MSISDNs
		for _, msisdn := range campaign.Segment.MSISDNs {
			user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
			if err == nil && user != nil {
				targetUsers = append(targetUsers, user)
			}
		}
	}

	if err != nil {
		return err
	}

	// Update campaign status
	campaign.Status = "RUNNING"
	campaign.StartedAt = time.Now()
	err = s.campaignRepo.Update(ctx, campaign)
	if err != nil {
		return err
	}

	// Send notifications to target users
	for _, user := range targetUsers {
		// Skip users who have opted out
		if !user.OptInStatus {
			continue
		}

		// Send notification
		_, err := s.SendSMS(ctx, user.MSISDN, template.Content, template.Type, campaign.ID)
		if err == nil {
			campaign.TotalSent++
			campaign.Delivered++
		} else {
			campaign.Failed++
		}
	}

	// Update campaign status
	campaign.Status = "COMPLETED"
	campaign.CompletedAt = time.Now()
	return s.campaignRepo.Update(ctx, campaign)
}

// CreateTemplate creates a new notification template
func (s *NotificationService) CreateTemplate(ctx context.Context, template *models.Template) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	return s.templateRepo.Create(ctx, template)
}

// GetTemplateByID retrieves a template by ID
func (s *NotificationService) GetTemplateByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	return s.templateRepo.FindByID(ctx, id)
}

// GetTemplateByName retrieves a template by name
func (s *NotificationService) GetTemplateByName(ctx context.Context, name string) (*models.Template, error) {
	return s.templateRepo.FindByName(ctx, name)
}

// GetTemplatesByType retrieves templates by type with pagination
func (s *NotificationService) GetTemplatesByType(ctx context.Context, templateType string, page, limit int) ([]*models.Template, error) {
	return s.templateRepo.FindByType(ctx, templateType, page, limit)
}

// GetAllTemplates retrieves all templates with pagination
func (s *NotificationService) GetAllTemplates(ctx context.Context, page, limit int) ([]*models.Template, error) {
	return s.templateRepo.FindAll(ctx, page, limit)
}

// UpdateTemplate updates a template
func (s *NotificationService) UpdateTemplate(ctx context.Context, template *models.Template) error {
	template.UpdatedAt = time.Now()
	return s.templateRepo.Update(ctx, template)
}

// DeleteTemplate deletes a template
func (s *NotificationService) DeleteTemplate(ctx context.Context, id primitive.ObjectID) error {
	return s.templateRepo.Delete(ctx, id)
}

// GetNotificationCount gets the total number of notifications
func (s *NotificationService) GetNotificationCount(ctx context.Context) (int64, error) {
	return s.notificationRepo.Count(ctx)
}

// GetCampaignCount gets the total number of campaigns
func (s *NotificationService) GetCampaignCount(ctx context.Context) (int64, error) {
	return s.campaignRepo.Count(ctx)
}

// GetTemplateCount gets the total number of templates
func (s *NotificationService) GetTemplateCount(ctx context.Context) (int64, error) {
	return s.templateRepo.Count(ctx)
}
