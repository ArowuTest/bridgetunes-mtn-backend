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
	MockSMS   bool // Added this field to match constructor
}

// KodobeGateway represents a Kodobe SMS gateway
type KodobeGateway struct {
	BaseURL string
	APIKey  string
	MockSMS bool // Added this field to match constructor
}

// MockGateway represents a mock SMS gateway for testing
type MockGateway struct {
	Name string
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

// NewMockGateway creates a new Mock SMS gateway
func NewMockGateway(name string) Gateway {
	return &MockGateway{Name: name}
}

// SendSMS sends an SMS using the MTN gateway
func (g *MTNGateway) SendSMS(msisdn, message string) (string, error) {
	 if g.MockSMS {
		 return fmt.Sprintf("MTN-MOCK-MSG-%d", time.Now().UnixNano()), nil
	 }
	 // Placeholder for real implementation
	 fmt.Printf("[MTN Gateway] Sending SMS to %s: %s\n", msisdn, message)
	 return "", errors.New("real MTN SMS gateway not implemented")
}

// GetDeliveryStatus gets the delivery status of an SMS from MTN
func (g *MTNGateway) GetDeliveryStatus(messageID string) (string, error) {
	 if g.MockSMS {
		 return "DELIVERED", nil
	 }
	 // Placeholder for real implementation
	 fmt.Printf("[MTN Gateway] Getting status for %s\n", messageID)
	 return "", errors.New("real MTN SMS gateway status check not implemented")
}

// SendSMS sends an SMS using the Kodobe gateway
func (g *KodobeGateway) SendSMS(msisdn, message string) (string, error) {
	 if g.MockSMS {
		 return fmt.Sprintf("KODOBE-MOCK-MSG-%d", time.Now().UnixNano()), nil
	 }
	 // Placeholder for real implementation
	 fmt.Printf("[Kodobe Gateway] Sending SMS to %s: %s\n", msisdn, message)
	 return "", errors.New("real Kodobe SMS gateway not implemented")
}

// GetDeliveryStatus gets the delivery status of an SMS from Kodobe
func (g *KodobeGateway) GetDeliveryStatus(messageID string) (string, error) {
	 if g.MockSMS {
		 return "DELIVERED", nil
	 }
	 // Placeholder for real implementation
	 fmt.Printf("[Kodobe Gateway] Getting status for %s\n", messageID)
	 return "", errors.New("real Kodobe SMS gateway status check not implemented")
}

// SendSMS sends an SMS using the Mock gateway
func (g *MockGateway) SendSMS(msisdn, message string) (string, error) {
	 msgID := fmt.Sprintf("%s-MOCK-MSG-%d", g.Name, time.Now().UnixNano())
	 fmt.Printf("[%s Mock Gateway] Simulating SendSMS to %s: %s -> %s\n", g.Name, msisdn, message, msgID)
	 return msgID, nil
}

// GetDeliveryStatus gets the delivery status of an SMS from the Mock gateway
func (g *MockGateway) GetDeliveryStatus(messageID string) (string, error) {
	 fmt.Printf("[%s Mock Gateway] Simulating GetDeliveryStatus for %s -> DELIVERED\n", g.Name, messageID)
	 return "DELIVERED", nil
}


