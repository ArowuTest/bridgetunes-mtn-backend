package services

import (
	"context"
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
	// notificationRepo repositories.NotificationRepository // Commented out - undefined
	// campaignRepo     repositories.CampaignRepository // Commented out - undefined
	// templateRepo     repositories.TemplateRepository // Commented out - undefined
	userRepo         repositories.UserRepository // Use interface type for dependency
	mtnGateway       smsgateway.Gateway
	 kodobeGateway    smsgateway.Gateway
	defaultGateway   string
}

// NewLegacyNotificationService creates a new LegacyNotificationService
func NewLegacyNotificationService(
	// notificationRepo repositories.NotificationRepository, // Commented out - undefined
	// templateRepo repositories.TemplateRepository, // Commented out - undefined
	// campaignRepo repositories.CampaignRepository, // Commented out - undefined
	 userRepo repositories.UserRepository, // Use interface type for dependency
	 mtnGateway smsgateway.Gateway,
	 kodobeGateway smsgateway.Gateway,
	 defaultGateway string,
) *LegacyNotificationService {
	return &LegacyNotificationService{
		// notificationRepo: notificationRepo, // Commented out - undefined
		// templateRepo:     templateRepo, // Commented out - undefined
		// campaignRepo:     campaignRepo, // Commented out - undefined
		 userRepo:         userRepo,
		 mtnGateway:       mtnGateway,
		 kodobeGateway:    kodobeGateway,
		 defaultGateway:   defaultGateway,
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
	// Create notification record
	notification := &models.Notification{
		MSISDN:     msisdn,
		Content:    content,
		Type:       notificationType,
		Status:     "PENDING", // Start as PENDING
		SentDate:   time.Time{}, // Set SentDate only on successful send
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
	 	// If default gateway fails, try the other one as fallback
	 	 if s.defaultGateway == "MTN" && s.kodobeGateway != nil {
	 	 	gateway = s.kodobeGateway
	 	 	notification.Gateway = "KODOBE"
	 	 	messageID, err = gateway.SendSMS(msisdn, content)
	 	 } else if s.defaultGateway != "MTN" && s.mtnGateway != nil {
	 	 	gateway = s.mtnGateway
	 	 	notification.Gateway = "MTN"
	 	 	messageID, err = gateway.SendSMS(msisdn, content)
	 	 }

	 	 // If fallback also fails
	 	 if err != nil {
	 	 	notification.Status = "FAILED"
	 	 	// s.notificationRepo.Create(ctx, notification) // Save failed attempt - Commented out - repo undefined
	 	 	 return notification, err // Return the failed notification and the error
	 	 }
	 }

	// SMS sent successfully (either primary or fallback)
	notification.MessageID = messageID
	notification.Status = "SENT"
	notification.SentDate = time.Now()
	 // err = s.notificationRepo.Create(ctx, notification) // Save successful attempt - Commented out - repo undefined
	 // if err != nil {
	 // 	 // Log error saving notification, but SMS was sent
	 // 	 return notification, err // Return notification and DB error
	 // }

	return notification, nil // Return nil error if repo save is commented out
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

