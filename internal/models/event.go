package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventStatus string

const (
	EventStatusActive      EventStatus = "ACTIVE"
	EventStatusDeactivated EventStatus = "DEACTIVATED"
)

type Event struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title                 string            `json:"title" bson:"title" validate:"required"`
	ChatURL              string            `json:"chatUrl" bson:"chat_url"`
	StreamURL            string            `json:"streamUrl" bson:"stream_url"`
	BannerURL            string            `json:"bannerUrl" bson:"banner_url"`
	StartAt              time.Time         `json:"startAt" bson:"start_at"`
	EndAt                time.Time         `json:"endAt" bson:"end_at"`
	EventMode            string            `json:"eventMode" bson:"event_mode"`
	Status               EventStatus       `json:"status" bson:"status" default:"ACTIVE"`
	HeroSectionTitle     string            `json:"heroSectionTitle" bson:"hero_section_title"`
	HeroSectionDescription string          `json:"heroSectionDescription" bson:"hero_section_description"`
	HeroSectionURLs      string            `json:"heroSectionUrls" bson:"hero_section_urls"`
	VideoCarouselTitle   string            `json:"videoCarouselTitle" bson:"video_carousel_title"`
	VideoCarouselDescription string        `json:"videoCarouselDescription" bson:"video_carousel_description"`
	CreatedAt            time.Time         `json:"createdAt" bson:"created_at"`
	UpdatedAt            time.Time         `json:"updatedAt" bson:"updated_at"`
}

// NewEvent creates a new Event with default values
func NewEvent() *Event {
	return &Event{
		Status:    EventStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
} 