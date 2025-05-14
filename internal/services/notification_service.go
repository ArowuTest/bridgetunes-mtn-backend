package services

import (
	"context"
	"fmt" // Added missing import
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"github.com/ArowuTest/bridgetunes-mtn-backend/pkg/smsgateway"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Compile-time check to ensure LegacyNotificationService implements NotificationService
var _ NotificationService = (*LegacyNotificationService)(nil)

// LegacyNotificationService handles notification-related business logic
type LegacyNotificationService struct {
	// notificationRepo repositories.NotificationRepository // Commented out - Keep commented
	// campaignRepo     repositories.CampaignRepository // Commented out - Keep commented
	// templateRepo     repositories.TemplateRepository // Commented out - Keep commented
	userRepo         repositories.UserRepository // Use interface type for dependency
	mtnGateway       smsgateway.Gateway
	 kodobeGateway    smsgateway.Gateway
	 uduxGateway    smsgateway.Gateway
	 settingsRepo     repositories.SystemSettingsRepository
}

// NewLegacyNotificationService creates a new LegacyNotificationService
func NewLegacyNotificationService(
	// notificationRepo repositories.NotificationRepository, // Keep commented
	// templateRepo repositories.TemplateRepository, // Keep commented
	// campaignRepo repositories.CampaignRepository, // Keep commented
	 userRepo repositories.UserRepository, // Use interface type for dependency
	 mtnGateway smsgateway.Gateway,
	 kodobeGateway smsgateway.Gateway,
	 uduxGateway smsgateway.Gateway,
	 settingsRepo repositories.SystemSettingsRepository,
) *LegacyNotificationService {
	return &LegacyNotificationService{
		// notificationRepo: notificationRepo, // Keep commented
		// templateRepo:     templateRepo, // Keep commented
		// campaignRepo:     campaignRepo, // Keep commented
		 userRepo:         userRepo,
		 mtnGateway:       mtnGateway,
		 kodobeGateway:    kodobeGateway,
		 uduxGateway:      uduxGateway,
		 settingsRepo:     settingsRepo,
	}
}

// GetNotificationByID retrieves a notification by ID
func (s *LegacyNotificationService) GetNotificationByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error) {
	// return s.notificationRepo.FindByID(ctx, id) // Commented out - repo undefined
	return nil, fmt.Errorf("NotificationRepository not implemented")
}

// GetNotificationsByMSISDN retrieves notifications by MSISDN with pagination
func (s *LegacyNotificationService) GetNotificationsByMSISDN(ctx context.Context, msisdn string, page, limit int) ([]*models.Notification, error) {
	// return s.notificationRepo.FindByMSISDN(ctx, msisdn, page, limit) // Commented out - repo undefined
	return nil, fmt.Errorf("NotificationRepository not implemented")
}

// GetNotificationsByCampaignID retrieves notifications by campaign ID with pagination
func (s *LegacyNotificationService) GetNotificationsByCampaignID(ctx context.Context, campaignID primitive.ObjectID, page, limit int) ([]*models.Notification, error) {
	// return s.notificationRepo.FindByCampaignID(ctx, campaignID, page, limit) // Commented out - repo undefined
	return nil, fmt.Errorf("NotificationRepository not implemented")
}

// GetNotificationsByStatus retrieves notifications by status with pagination
func (s *LegacyNotificationService) GetNotificationsByStatus(ctx context.Context, status string, page, limit int) ([]*models.Notification, error) {
	// return s.notificationRepo.FindByStatus(ctx, status, page, limit) // Commented out - repo undefined
	return nil, fmt.Errorf("NotificationRepository not implemented")
}

// SendSMS sends an SMS notification
func (s *LegacyNotificationService) SendSMS(ctx context.Context, msisdn, content, notificationType string, campaignID primitive.ObjectID) (*models.Notification, error) {
	// Get current system settings
	settings, err := s.settingsRepo.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system settings: %w", err)
	}

	// Create notification record
	notification := &models.Notification{
		MSISDN:     msisdn,
		Content:    content,
		Type:       notificationType,
		Status:     "PENDING", // Start as PENDING
		SentDate:   time.Time{}, // Set SentDate only on successful send
		CampaignID: campaignID,
		Gateway:    settings.SMSGateway,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	// log.Print("notification:", notification) //gets here


	// Select gateway based on system settings
	var gateway smsgateway.Gateway
	switch settings.SMSGateway {
	case "MTN":
		gateway = s.mtnGateway
	case "KODOBE":
		gateway = s.kodobeGateway
	case "UDUX":
		gateway = s.uduxGateway
	default:
		gateway = s.uduxGateway
	}

	// Send SMS
	messageID, err := gateway.SendSMS(msisdn, content)
	if err != nil {
		// If default gateway fails, try the other ones as fallback
		if settings.SMSGateway != "MTN" && s.mtnGateway != nil {
			gateway = s.mtnGateway
			notification.Gateway = "MTN"
			messageID, err = gateway.SendSMS(msisdn, content)
		}
		if err != nil && settings.SMSGateway != "KODOBE" && s.kodobeGateway != nil {
			gateway = s.kodobeGateway
			notification.Gateway = "KODOBE"
			messageID, err = gateway.SendSMS(msisdn, content)
		}
		if err != nil && settings.SMSGateway != "UDUX" && s.uduxGateway != nil {
			gateway = s.uduxGateway
			notification.Gateway = "UDUX"
			messageID, err = gateway.SendSMS(msisdn, content)
		}

		// If all gateways fail
		if err != nil {
			notification.Status = "FAILED"
			return notification, err
		}
	}

	// SMS sent successfully
	notification.MessageID = messageID
	notification.Status = "SENT"
	notification.SentDate = time.Now()

	return notification, nil
}

// CreateCampaign creates a new notification campaign
func (s *LegacyNotificationService) CreateCampaign(ctx context.Context, campaign *models.Campaign) error {
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	// return s.campaignRepo.Create(ctx, campaign) // Commented out - repo undefined
	return fmt.Errorf("CampaignRepository not implemented")
}

// GetAllCampaigns retrieves all campaigns with pagination
func (s *LegacyNotificationService) GetAllCampaigns(ctx context.Context, page, limit int) ([]models.Campaign, error) {
	 // return s.campaignRepo.FindAll(ctx, page, limit) // Commented out - repo undefined
	 return nil, fmt.Errorf("CampaignRepository not implemented")
}

// ExecuteCampaign executes a notification campaign
func (s *LegacyNotificationService) ExecuteCampaign(ctx context.Context, campaignID primitive.ObjectID) error {
	// Get the campaign
	 // campaign, err := s.campaignRepo.FindByID(ctx, campaignID) // Commented out - repo undefined
	 // if err != nil {
	 // 	 return err
	 // }
	 return fmt.Errorf("CampaignRepository not implemented") // Return error as campaign cannot be fetched

	/* // Commenting out rest of the function as it depends on campaign and template repos
	// Check if campaign is already running or completed
	 if campaign.Status == "RUNNING" || campaign.Status == "COMPLETED" {
	 	 return nil // Or return an error?
	 }

	// Get the template
	 template, err := s.templateRepo.FindByID(ctx, campaign.TemplateID)
	 if err != nil {
	 	 return err
	 }

	// Get target users based on segment
	 var targetUsers []*models.User
	 // Placeholder: Fetch all opted-in users for now. Needs proper segmentation logic.
	 targetUsers, err = s.userRepo.FindByOptInStatus(ctx, true, 1, 10000) // Consider pagination for large user bases

	 if err != nil {
	 	 return err
	 }

	// Update campaign status to RUNNING
	 campaign.Status = "RUNNING"
	 err = s.campaignRepo.Update(ctx, campaign)
	 if err != nil {
	 	 return err
	 }

	// Send notifications asynchronously?
	// For now, sending synchronously
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
	 		 // Log individual send errors
	 	 }
	 }

	// Update campaign status to COMPLETED and stats
	 campaign.Status = "COMPLETED"
	 // Add stats update here if fields exist in Campaign model (e.g., campaign.TotalSent = totalSent)
	 return s.campaignRepo.Update(ctx, campaign)
	*/
}

// CreateTemplate creates a new notification template
func (s *LegacyNotificationService) CreateTemplate(ctx context.Context, template *models.Template) error {
	 template.CreatedAt = time.Now()
	 template.UpdatedAt = time.Now()
	 // return s.templateRepo.Create(ctx, template) // Commented out - repo undefined
	 return fmt.Errorf("TemplateRepository not implemented")
}

// GetTemplateByID retrieves a template by ID
func (s *LegacyNotificationService) GetTemplateByID(ctx context.Context, id primitive.ObjectID) (*models.Template, error) {
	 // return s.templateRepo.FindByID(ctx, id) // Commented out - repo undefined
	 return nil, fmt.Errorf("TemplateRepository not implemented")
}

// GetTemplateByName retrieves a template by name
func (s *LegacyNotificationService) GetTemplateByName(ctx context.Context, name string) (*models.Template, error) {
	 // return s.templateRepo.FindByName(ctx, name) // Commented out - repo undefined
	 return nil, fmt.Errorf("TemplateRepository not implemented")
}

// GetTemplatesByType retrieves templates by type with pagination
func (s *LegacyNotificationService) GetTemplatesByType(ctx context.Context, templateType string, page, limit int) ([]*models.Template, error) {
	 // return s.templateRepo.FindByType(ctx, templateType, page, limit) // Commented out - repo undefined
	 return nil, fmt.Errorf("TemplateRepository not implemented")
}

// GetAllTemplates retrieves all templates with pagination
func (s *LegacyNotificationService) GetAllTemplates(ctx context.Context, page, limit int) ([]*models.Template, error) {
	 // return s.templateRepo.FindAll(ctx, page, limit) // Commented out - repo undefined
	 return nil, fmt.Errorf("TemplateRepository not implemented")
}

// UpdateTemplate updates a template
func (s *LegacyNotificationService) UpdateTemplate(ctx context.Context, template *models.Template) error {
	 template.UpdatedAt = time.Now()
	 // return s.templateRepo.Update(ctx, template) // Commented out - repo undefined
	 return fmt.Errorf("TemplateRepository not implemented")
}

// DeleteTemplate deletes a template
func (s *LegacyNotificationService) DeleteTemplate(ctx context.Context, id primitive.ObjectID) error {
	 // return s.templateRepo.Delete(ctx, id) // Commented out - repo undefined
	 return fmt.Errorf("TemplateRepository not implemented")
}

// GetNotificationCount gets the total number of notifications
func (s *LegacyNotificationService) GetNotificationCount(ctx context.Context) (int64, error) {
	 // return s.notificationRepo.Count(ctx) // Commented out - repo undefined
	 return 0, fmt.Errorf("NotificationRepository not implemented")
}

// GetCampaignCount gets the total number of campaigns
func (s *LegacyNotificationService) GetCampaignCount(ctx context.Context) (int64, error) {
	 // return s.campaignRepo.Count(ctx) // Commented out - repo undefined
	 return 0, fmt.Errorf("CampaignRepository not implemented")
}

// GetTemplateCount gets the total number of templates
func (s *LegacyNotificationService) GetTemplateCount(ctx context.Context) (int64, error) {
	 // return s.templateRepo.Count(ctx) // Commented out - repo undefined
	 return 0, fmt.Errorf("TemplateRepository not implemented")
}

