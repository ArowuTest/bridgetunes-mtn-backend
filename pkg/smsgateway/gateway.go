package smsgateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/config"
	"github.com/ArowuTest/bridgetunes-mtn-backend/pkg/jwt"
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

// UduxGateway represents the Udux SMS gateway
type UduxGateway struct {
	BaseURL      string
	APISecret    string
	tokenService *jwt.SMSTokenService
	httpClient   *http.Client
}

// MockGateway represents a mock SMS gateway for testing
type MockGateway struct {
	Name string
}

// NewMTNGateway creates a new MTN SMS gateway
func NewMTNGateway(cfg *config.Config, mockSMS bool) Gateway {
	return &MTNGateway{
		BaseURL:   cfg.SMS.MTNGateway.BaseURL,
		APIKey:    cfg.SMS.MTNGateway.APIKey,
		APISecret: cfg.SMS.MTNGateway.APISecret,
		MockSMS:   mockSMS,
	}
}

// NewKodobeGateway creates a new Kodobe SMS gateway
func NewKodobeGateway(cfg *config.Config, mockSMS bool) Gateway {
	return &KodobeGateway{
		BaseURL: cfg.SMS.KodobeGateway.BaseURL,
		APIKey:  cfg.SMS.KodobeGateway.APIKey,
		MockSMS: mockSMS,
	} 
}

// NewUduxGateway creates a new UduxGateway
func NewUduxGateway(cfg *config.Config) Gateway {
	return &UduxGateway{
		BaseURL:      cfg.UduxGateway.BaseURL,
		APISecret:    cfg.UduxGateway.APISecret,
		tokenService: jwt.NewSMSTokenService(cfg),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

// SendSMS sends an SMS using the Udux gateway
func (g *UduxGateway) SendSMS(msisdn, message string) (string, error) {
	// Get SMS tokens
	tokens, err := g.tokenService.GetSMSTokens("music", "admin")
	if err != nil {
		return "", fmt.Errorf("failed to get SMS tokens: %w", err)
	}

	log.Println("encrypted token  ", tokens.AccessToken)

	// Prepare the request body
	requestBody := map[string]interface{}{
		"phoneNumber":           msisdn,
		"message":          message,
		"uduxService": "music",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s", g.BaseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))

	// Send the request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var response struct {
		MessageID string `json:"messageId"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.MessageID, nil
}

// GetDeliveryStatus gets the delivery status of an SMS from Udux
func (g *UduxGateway) GetDeliveryStatus(messageID string) (string, error) {
	// Get SMS tokens
	tokens, err := g.tokenService.GetSMSTokens("UDUX_ADMIN", "ADMIN")
	if err != nil {
		return "", fmt.Errorf("failed to get SMS tokens: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notifications/status/%s", g.BaseURL, messageID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))

	// Send the request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var response struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Status, nil
}

