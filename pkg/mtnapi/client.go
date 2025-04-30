package mtnapi

import (
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
		client: &http.Client{Timeout: 10 * time.Second},
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
	// rand.Seed(time.Now().UnixNano()) // Deprecated in Go 1.20+, automatically seeded

	// Generate between 10 and 50 random topups
	numTopups := rand.Intn(41) + 10

	for i := 0; i < numTopups; i++ {
		// Generate a random MSISDN starting with 080
		msisdn := fmt.Sprintf("080%08d", rand.Intn(100000000))

		// Generate a random amount between 100 and 1000
		amount := float64(rand.Intn(10)+1) * 100

		// Generate a random date between startDate and endDate
		var randomDate time.Time
		// Ensure startDate is before endDate to avoid panic in Int63n
		// Check if startDate and endDate are the same
		// If they are the same, set randomDate to startDate
		// Otherwise, calculate the duration and generate a random date
		// This avoids potential issues with Int63n(0)
		// and ensures randomDate is always within the valid range
		// or equal to startDate if the range is zero.
		// Note: This logic assumes startDate is not after endDate.
		// Proper validation might be needed elsewhere if that's possible.
		// Consider adding a check: if startDate.After(endDate) { handle error or swap }
		// For now, assuming startDate <= endDate based on typical usage.
		// Also, rand.Int63n is preferred over rand.Intn for larger ranges if needed,
		// but time.Duration is int64, so Int63n is suitable.
		// Let's refine the date generation logic slightly for clarity and safety.
		
		// Calculate duration safely
		var duration time.Duration
		 if !startDate.After(endDate) {
		 	 duration = endDate.Sub(startDate)
		 }
		
		// Generate random duration offset
		var randomOffset time.Duration
		 if duration > 0 {
		 	 randomOffset = time.Duration(rand.Int63n(int64(duration)))
		 }
		
		// Calculate random date
		 randomDate = startDate.Add(randomOffset)

		// Generate a random transaction reference
		transactionRef := fmt.Sprintf("TXN%012d", rand.Int63n(1000000000000))

		// Append the generated top-up response to the slice
		// Ensure all fields are correctly populated
		// Check for potential nil pointers or uninitialized values if applicable
		// In this case, all fields are value types or initialized, so it's safe.
		// Consider adding validation for generated data if needed (e.g., amount > 0).
		// For mock data, current generation seems sufficient.
		// Final check of the TopupResponse struct fields and types:
		// MSISDN: string - OK
		// Amount: float64 - OK
		// TransactionRef: string - OK
		// Date: time.Time - OK
		// Status: string - OK
		// All fields match the struct definition.
		// Append the generated TopupResponse
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
	// Ensure length check is correct (e.g., 11 digits for typical Nigerian numbers)
	// Check prefix correctly
	// Consider adding more sophisticated validation if needed for mock
	// Current logic: length 11 and starts with "080"
	 if len(msisdn) == 11 && msisdn[:3] == "080" {
		return true, nil
	}
	return false, nil
}

