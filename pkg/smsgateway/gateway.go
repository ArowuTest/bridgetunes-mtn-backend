package smsgateway

import (
	"errors"
	"fmt"
	"time"
)

// Gateway represents an SMS gateway interface
type Gateway interface {
	SendSMS(msisdn, message string) (string, error)
	GetDeliveryStatus(messageID string) (string, error)
}

// MTNGateway represents an MTN SMS gateway
type MTNGateway struct {
	BaseURL   string
	APIKey    string
	APISecret string
	MockSMS   bool
}

// KodobeGateway represents a Kodobe SMS gateway
type KodobeGateway struct {
	BaseURL string
	APIKey  string
	MockSMS bool
}

// NewMTNGateway creates a new MTN SMS gateway
func NewMTNGateway(baseURL, apiKey, apiSecret string, mockSMS bool) Gateway {
	return &MTNGateway{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		APISecret: apiSecret,
		MockSMS:   mockSMS,
	}
}

// NewKodobeGateway creates a new Kodobe SMS gateway
func NewKodobeGateway(baseURL, apiKey string, mockSMS bool) Gateway {
	return &KodobeGateway{
		BaseURL: baseURL,
		APIKey:  apiKey,
		MockSMS: mockSMS,
	}
}

// SendSMS sends an SMS using the MTN gateway
func (g *MTNGateway) SendSMS(msisdn, message string) (string, error) {
	if g.MockSMS {
		return fmt.Sprintf("MTN-MSG-%d", time.Now().UnixNano()), nil
	}
	return "", errors.New("real MTN SMS gateway not implemented")
}

// GetDeliveryStatus gets the delivery status of an SMS
func (g *MTNGateway) GetDeliveryStatus(messageID string) (string, error) {
	if g.MockSMS {
		return "DELIVERED", nil
	}
	return "", errors.New("real MTN SMS gateway not implemented")
}

// SendSMS sends an SMS using the Kodobe gateway
func (g *KodobeGateway) SendSMS(msisdn, message string) (string, error) {
	if g.MockSMS {
		return fmt.Sprintf("KODOBE-MSG-%d", time.Now().UnixNano()), nil
	}
	return "", errors.New("real Kodobe SMS gateway not implemented")
}

// GetDeliveryStatus gets the delivery status of an SMS
func (g *KodobeGateway) GetDeliveryStatus(messageID string) (string, error) {
	if g.MockSMS {
		return "DELIVERED", nil
	}
	return "", errors.New("real Kodobe SMS gateway not implemented")
}
