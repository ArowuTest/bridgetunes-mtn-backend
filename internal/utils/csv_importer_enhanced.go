package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CSVImporterEnhanced provides enhanced functionality for importing data from CSV files
type CSVImporterEnhanced struct {
	userRepo   repositories.UserRepository
	topupRepo  repositories.TopupRepository
	drawRepo   repositories.DrawRepository
	winnerRepo repositories.WinnerRepository
	configRepo repositories.SystemConfigRepository
}

// NewCSVImporterEnhanced creates a new CSVImporterEnhanced
func NewCSVImporterEnhanced(
	userRepo repositories.UserRepository,
	topupRepo repositories.TopupRepository,
	drawRepo repositories.DrawRepository,
	winnerRepo repositories.WinnerRepository,
	configRepo repositories.SystemConfigRepository,
) *CSVImporterEnhanced {
	return &CSVImporterEnhanced{
		userRepo:   userRepo,
		topupRepo:  topupRepo,
		drawRepo:   drawRepo,
		winnerRepo: winnerRepo,
		configRepo: configRepo,
	}
}

// ImportUsersAndTopups imports users and topups from a CSV file
func (i *CSVImporterEnhanced) ImportUsersAndTopups(filePath string) (map[string]interface{}, error) {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Map column indices
	msisdnIdx := findColumnIndex(header, []string{"MSISDN", "Phone Number", "Mobile"})
	amountIdx := findColumnIndex(header, []string{"Recharge Amount", "Amount", "Topup Amount"})
	optInIdx := findColumnIndex(header, []string{"Opt-In Status", "OptIn", "Opted In"})
	dateIdx := findColumnIndex(header, []string{"Recharge Date", "Date", "Topup Date"})

	if msisdnIdx == -1 {
		return nil, fmt.Errorf("MSISDN column not found in CSV")
	}

	// Initialize result counters
	results := map[string]interface{}{
		"totalRows":     0,
		"usersCreated":  0,
		"usersUpdated":  0,
		"topupsCreated": 0,
		"errors":        []string{},
	}

	// Process each row
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Error reading row: %v", err))
			continue
		}

		results["totalRows"] = results["totalRows"].(int) + 1

		// Extract MSISDN
		msisdn := row[msisdnIdx]
		if msisdn == "" {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: No MSISDN found", results["totalRows"]))
			continue
		}

		// Clean MSISDN (remove spaces, ensure it starts with country code if needed)
		msisdn = cleanMSISDN(msisdn)

		// Extract amount
		var amount float64
		if amountIdx != -1 && row[amountIdx] != "" {
			amount, err = strconv.ParseFloat(strings.TrimSpace(row[amountIdx]), 64)
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid amount: %s", results["totalRows"], row[amountIdx]))
				amount = 0
			}
		}

		// Extract opt-in status
		var optInStatus bool
		if optInIdx != -1 && row[optInIdx] != "" {
			optInStr := strings.ToLower(strings.TrimSpace(row[optInIdx]))
			optInStatus = optInStr == "yes" || optInStr == "true" || optInStr == "1" || optInStr == "y"
		}

		// Extract date
		var date time.Time
		if dateIdx != -1 && row[dateIdx] != "" {
			date, err = parseDate(row[dateIdx])
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid date: %s", results["totalRows"], row[dateIdx]))
				date = time.Now()
			}
		} else {
			date = time.Now()
		}

		// Calculate points based on amount
		points := calculatePoints(amount)

		// Create or update user
		user := &models.User{
			MSISDN:       msisdn,
			OptInStatus:  optInStatus,
			OptInDate:    date,
			OptInChannel: "CSV_IMPORT",
			Points:       points,
			IsBlacklisted: false,
			LastActivity: date,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		existingUser, err := i.userRepo.FindByMSISDN(nil, msisdn)
		if err == nil && existingUser != nil {
			// Update existing user
			existingUser.OptInStatus = optInStatus
			if optInStatus && existingUser.OptInDate.IsZero() {
				existingUser.OptInDate = date
			}
			existingUser.Points = points
			existingUser.LastActivity = date
			existingUser.UpdatedAt = time.Now()

			err = i.userRepo.Update(nil, existingUser)
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Failed to update user: %v", results["totalRows"], err))
				continue
			}
			results["usersUpdated"] = results["usersUpdated"].(int) + 1
		} else {
			// Create new user
			err = i.userRepo.Create(nil, user)
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Failed to create user: %v", results["totalRows"], err))
				continue
			}
			results["usersCreated"] = results["usersCreated"].(int) + 1
		}

		// Create topup if amount is provided
		if amount > 0 {
			topup := &models.Topup{
				MSISDN:        msisdn,
				Amount:        amount,
				Channel:       "CSV_IMPORT",
				Date:          date,
				TransactionRef: fmt.Sprintf("CSV_IMPORT_%d_%s", time.Now().UnixNano(), primitive.NewObjectID().Hex()),
				PointsEarned:  points,
				Processed:     true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			err = i.topupRepo.Create(nil, topup)
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Failed to create topup: %v", results["totalRows"], err))
				continue
			}
			results["topupsCreated"] = results["topupsCreated"].(int) + 1
		}
	}

	return results, nil
}

// ImportPrizeStructures imports prize structures from a CSV file
func (i *CSVImporterEnhanced) ImportPrizeStructures(filePath string) (map[string]interface{}, error) {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Map column indices
	categoryIdx := findColumnIndex(header, []string{"Category", "Prize Category"})
	dailyAmountIdx := findColumnIndex(header, []string{"Daily Amount", "Daily Prize", "Daily Prizes [Mon - Fri]"})
	weeklyAmountIdx := findColumnIndex(header, []string{"Weekly Amount", "Weekly Prize", "Saturday Prizes"})
	countIdx := findColumnIndex(header, []string{"Count", "Number of Prizes"})

	if categoryIdx == -1 || (dailyAmountIdx == -1 && weeklyAmountIdx == -1) {
		return nil, fmt.Errorf("required columns not found in CSV")
	}

	// Initialize result counters
	results := map[string]interface{}{
		"totalRows":      0,
		"dailyPrizes":    0,
		"weeklyPrizes":   0,
		"errors":         []string{},
	}

	// Initialize prize structures
	var dailyPrizes []models.PrizeStructure
	var weeklyPrizes []models.PrizeStructure

	// Process each row
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Error reading row: %v", err))
			continue
		}

		results["totalRows"] = results["totalRows"].(int) + 1

		// Extract category
		category := strings.TrimSpace(row[categoryIdx])
		if category == "" {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: No category found", results["totalRows"]))
			continue
		}

		// Extract count
		count := 1
		if countIdx != -1 && row[countIdx] != "" {
			countVal, err := strconv.Atoi(strings.TrimSpace(row[countIdx]))
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid count: %s", results["totalRows"], row[countIdx]))
			} else {
				count = countVal
			}
		}

		// Extract daily amount
		if dailyAmountIdx != -1 && row[dailyAmountIdx] != "" {
			amountStr := strings.TrimSpace(row[dailyAmountIdx])
			// Remove currency symbol if present
			amountStr = strings.ReplaceAll(amountStr, "₦", "")
			amountStr = strings.ReplaceAll(amountStr, "N", "")
			amountStr = strings.ReplaceAll(amountStr, ",", "")
			
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid daily amount: %s", results["totalRows"], row[dailyAmountIdx]))
			} else {
				dailyPrizes = append(dailyPrizes, models.PrizeStructure{
					Category: category,
					Amount:   amount,
					Count:    count,
				})
				results["dailyPrizes"] = results["dailyPrizes"].(int) + 1
			}
		}

		// Extract weekly amount
		if weeklyAmountIdx != -1 && row[weeklyAmountIdx] != "" {
			amountStr := strings.TrimSpace(row[weeklyAmountIdx])
			// Remove currency symbol if present
			amountStr = strings.ReplaceAll(amountStr, "₦", "")
			amountStr = strings.ReplaceAll(amountStr, "N", "")
			amountStr = strings.ReplaceAll(amountStr, ",", "")
			
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid weekly amount: %s", results["totalRows"], row[weeklyAmountIdx]))
			} else {
				weeklyPrizes = append(weeklyPrizes, models.PrizeStructure{
					Category: category,
					Amount:   amount,
					Count:    count,
				})
				results["weeklyPrizes"] = results["weeklyPrizes"].(int) + 1
			}
		}
	}

	// Save prize structures to system config
	if len(dailyPrizes) > 0 {
		dailyConfig := &models.SystemConfig{
			Key:         "prizeStructureDaily",
			Value:       dailyPrizes,
			Description: "Daily prize structure",
			UpdatedAt:   time.Now(),
		}

		existingConfig, err := i.configRepo.FindByKey(nil, "prizeStructureDaily")
		if err == nil && existingConfig != nil {
			existingConfig.Value = dailyPrizes
			existingConfig.UpdatedAt = time.Now()
			err = i.configRepo.Update(nil, existingConfig)
		} else {
			err = i.configRepo.Create(nil, dailyConfig)
		}

		if err != nil {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Failed to save daily prize structure: %v", err))
		}
	}

	if len(weeklyPrizes) > 0 {
		weeklyConfig := &models.SystemConfig{
			Key:         "prizeStructureWeekly",
			Value:       weeklyPrizes,
			Description: "Weekly prize structure",
			UpdatedAt:   time.Now(),
		}

		existingConfig, err := i.configRepo.FindByKey(nil, "prizeStructureWeekly")
		if err == nil && existingConfig != nil {
			existingConfig.Value = weeklyPrizes
			existingConfig.UpdatedAt = time.Now()
			err = i.configRepo.Update(nil, existingConfig)
		} else {
			err = i.configRepo.Create(nil, weeklyConfig)
		}

		if err != nil {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Failed to save weekly prize structure: %v", err))
		}
	}

	return results, nil
}

// Helper functions

// findColumnIndex finds the index of a column by possible names
func findColumnIndex(header []string, possibleNames []string) int {
	for i, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		for _, name := range possibleNames {
			if strings.ToLower(name) == h {
				return i
			}
		}
	}
	return -1
}

// cleanMSISDN cleans an MSISDN
func cleanMSISDN(msisdn string) string {
	// Remove spaces and other non-numeric characters
	msisdn = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, msisdn)

	// Ensure it starts with country code if needed
	if len(msisdn) == 10 && msisdn[0] == '0' {
		// Nigerian number starting with 0, add 234 country code
		msisdn = "234" + msisdn[1:]
	}

	return msisdn
}

// parseDate parses a date string in various formats
func parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	
	// Try various date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"Jan 2, 2006",
		"2 Jan 2006",
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
		"02/01/2006 15:04:05",
	}

	for _, format := range formats {
		date, err := time.Parse(format, dateStr)
		if err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// calculatePoints calculates points based on topup amount
func calculatePoints(amount float64) int {
	switch {
	case amount >= 1000:
		return 10
	case amount >= 900:
		return 9
	case amount >= 800:
		return 8
	case amount >= 700:
		return 7
	case amount >= 600:
		return 6
	case amount >= 500:
		return 5
	case amount >= 400:
		return 4
	case amount >= 300:
		return 3
	case amount >= 200:
		return 2
	case amount >= 100:
		return 1
	default:
		return 0
	}
}
