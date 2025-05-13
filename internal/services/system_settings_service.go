package services

import (
	"context"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
)

// SystemSettingsServiceImpl implements SystemSettingsService
type SystemSettingsServiceImpl struct {
	settingsRepo repositories.SystemSettingsRepository
}

// NewSystemSettingsService creates a new SystemSettingsService
func NewSystemSettingsService(settingsRepo repositories.SystemSettingsRepository) SystemSettingsService {
	return &SystemSettingsServiceImpl{
		settingsRepo: settingsRepo,
	}
}

// GetSettings retrieves the current system settings
func (s *SystemSettingsServiceImpl) GetSettings(ctx context.Context) (*models.SystemSettings, error) {
	return s.settingsRepo.GetSettings(ctx)
}

// UpdateSettings updates all system settings
func (s *SystemSettingsServiceImpl) UpdateSettings(ctx context.Context, settings *models.SystemSettings) error {
	return s.settingsRepo.UpdateSettings(ctx, settings)
}

// UpdateSMSGateway updates only the SMS gateway setting
func (s *SystemSettingsServiceImpl) UpdateSMSGateway(ctx context.Context, gateway string, updatedBy string) error {
	return s.settingsRepo.UpdateSMSGateway(ctx, gateway, updatedBy)
} 