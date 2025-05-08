package services

import (
	"context"
	"errors"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventServiceInterface interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	GetEvent(ctx context.Context, id primitive.ObjectID) (*models.Event, error)
	UpdateEvent(ctx context.Context, event *models.Event) error
	DeleteEvent(ctx context.Context, id primitive.ObjectID) error
	ListEvents(ctx context.Context, page, limit int) ([]*models.Event, error)
}

type EventService struct {
	eventRepo repositories.EventRepository
}

func NewEventService(eventRepo repositories.EventRepository) *EventService {
	return &EventService{
		eventRepo: eventRepo,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) error {
	if event.Title == "" {
		return errors.New("title is required")
	}
	if event.StartAt.After(event.EndAt) {
		return errors.New("start time cannot be after end time")
	}

	// Set default values if not provided
	if event.Status == "" {
		event.Status = models.EventStatusActive
	}
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	return s.eventRepo.Create(ctx, event)
}

func (s *EventService) GetEvent(ctx context.Context, id primitive.ObjectID) (*models.Event, error) {
	return s.eventRepo.FindByID(ctx, id)
}

func (s *EventService) UpdateEvent(ctx context.Context, event *models.Event) error {
	if event.Title == "" {
		return errors.New("title is required")
	}
	if event.StartAt.After(event.EndAt) {
		return errors.New("start time cannot be after end time")
	}
	event.UpdatedAt = time.Now()
	return s.eventRepo.Update(ctx, event)
}

func (s *EventService) DeleteEvent(ctx context.Context, id primitive.ObjectID) error {
	return s.eventRepo.Delete(ctx, id)
}

func (s *EventService) ListEvents(ctx context.Context, page, limit int) ([]*models.Event, error) {
	return s.eventRepo.FindAll(ctx, page, limit)
} 