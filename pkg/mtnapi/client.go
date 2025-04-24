package mtnapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Client represents an MTN API client
type Client struct {
	BaseURL   string
	APIKey    string
	APISecret string
	MockAPI   bool
	client    *http.Client
}

// TopupResponse represents a topup response from the MTN API
type TopupResponse struct {
	MSISDN         string    `json:"msisdn"`
	Amount         float64   `json:"amount"`
	TransactionRef string    `json:"transactionRef"`
	Date           time.Time `json:"date"`
	Status         string    `json:"status"`
}

// NewClient creates a new MTN API client
func NewClient(baseURL, apiKey, apiSecret string, mockAPI bool) *Client {
	return &Client{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		APISecret: apiSecret,
		MockAPI:   mockAPI,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// GetTopups retrieves topups for a given date range
func (c *Client) GetTopups(startDate, endDate time.Time) ([]TopupResponse, error) {
	if c.MockAPI {
		return c.mockGetTopups(startDate, endDate)
	}

	// In a real implementation, this would make an HTTP request to the MTN API
	// For now, we'll just return an error
	return nil, errors.New("real MTN API not implemented")
}

// mockGetTopups mocks the GetTopups method for testing
func (c *Client) mockGetTopups(startDate, endDate time.Time) ([]TopupResponse, error) {
	// Generate random topups for testing
	var topups []TopupResponse
	
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Generate between 10 and 50 random topups
	numTopups := rand.Intn(41) + 10
	
	for i := 0; i < numTopups; i++ {
		// Generate a random MSISDN starting with 080
		msisdn := fmt.Sprintf("080%08d", rand.Intn(100000000))
		
		// Generate a random amount between 100 and 1000
		amount := float64(rand.Intn(10) + 1) * 100
		
		// Generate a random date between startDate and endDate
		duration := endDate.Sub(startDate)
		randomDuration := time.Duration(rand.Int63n(int64(duration)))
		randomDate := startDate.Add(randomDuration)
		
		// Generate a random transaction reference
		transactionRef := fmt.Sprintf("TXN%012d", rand.Intn(1000000000000))
		
		topups = append(topups, TopupResponse{
			MSISDN:         msisdn,
			Amount:         amount,
			TransactionRef: transactionRef,
			Date:           randomDate,
			Status:         "SUCCESS",
		})
	}
	
	return topups, nil
}

// VerifyMSISDN verifies if an MSISDN is valid
func (c *Client) VerifyMSISDN(msisdn string) (bool, error) {
	if c.MockAPI {
		return c.mockVerifyMSISDN(msisdn)
	}

	// In a real implementation, this would make an HTTP request to the MTN API
	// For now, we'll just return an error
	return false, errors.New("real MTN API not implemented")
}

// mockVerifyMSISDN mocks the VerifyMSISDN method for testing
func (c *Client) mockVerifyMSISDN(msisdn string) (bool, error) {
	// For testing, we'll consider any MSISDN starting with 080 as valid
	if len(msisdn) == 11 && msisdn[:3] == "080" {
		return true, nil
	}
	return false, nil
}
