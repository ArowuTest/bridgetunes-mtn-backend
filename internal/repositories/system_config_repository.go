package repositories // This is the interface file

import (
	"context"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
)

// SystemConfigRepository defines the interface for system configuration repository
type SystemConfigRepository interface {
	// FindByKey finds a system configuration by key
	FindByKey(ctx context.Context, key string) (*models.SystemConfig, error)
	
	// Create creates a new system configuration
	Create(ctx context.Context, config *models.SystemConfig) error
	
	// Update updates a system configuration
	Update(ctx context.Context, config *models.SystemConfig) error
	
	// Delete deletes a system configuration
	Delete(ctx context.Context, key string) error
	
	// FindAll finds all system configurations
	FindAll(ctx context.Context) ([]*models.SystemConfig, error)
}
