package utils

import (
	"context" // Added context
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/models"
	"github.com/ArowuTest/bridgetunes-mtn-backend/internal/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
)

// CSVImporterEnhanced provides enhanced functionality for importing data from CSV files
type CSVImporterEnhanced struct {
	userRepo   repositories.UserRepository
	// topupRepo  repositories.TopupRepository // Assuming Topup repo is not directly needed here
	// drawRepo   repositories.DrawRepository
	// winnerRepo repositories.WinnerRepository
	configRepo repositories.SystemConfigRepository
}

// NewCSVImporterEnhanced creates a new CSVImporterEnhanced
func NewCSVImporterEnhanced(
	userRepo repositories.UserRepository,
	// topupRepo repositories.TopupRepository,
	// drawRepo repositories.DrawRepository,
	// winnerRepo repositories.WinnerRepository,
	configRepo repositories.SystemConfigRepository,
) *CSVImporterEnhanced {
	return &CSVImporterEnhanced{
		userRepo:   userRepo,
		// topupRepo:  topupRepo,
		// drawRepo:   drawRepo,
		// winnerRepo: winnerRepo,
		configRepo: configRepo,
	}
}

// ImportUsersAndTopups imports users and topups from a CSV file
// NOTE: This function seems to handle both user creation/update and topup creation.
// It might be better split, but keeping original logic for now.
// It also directly calculates and sets user points, bypassing the DrawService logic.
func (i *CSVImporterEnhanced) ImportUsersAndTopups(ctx context.Context, filePath string) (map[string]interface{}, error) {
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
		"topupsCreated": 0, // Assuming topups are still created here
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
				 date = time.Now() // Use current time as fallback?
			 }
		 } else {
			 date = time.Now() // Use current time if no date provided
		 }

		// Calculate points based on amount (Direct calculation, bypasses service logic)
		points := calculatePoints(amount)

		// Create or update user
		user := &models.User{
			MSISDN:       msisdn,
			OptInStatus:  optInStatus,
			OptInDate:    time.Time{}, // Initialize, set only if optInStatus is true
			OptInChannel: "CSV_IMPORT",
			Points:       0, // Initialize, will be updated
			IsBlacklisted: false,
			LastActivity: date,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		 if optInStatus {
			 user.OptInDate = date
		 }

		 existingUser, err := i.userRepo.FindByMSISDN(ctx, msisdn)
		 if err == nil && existingUser != nil {
			 // Update existing user
			 existingUser.OptInStatus = optInStatus
			 if optInStatus && existingUser.OptInDate.IsZero() {
				 existingUser.OptInDate = date
			 }
			 // WARNING: Overwriting points directly. Consider using IncrementPoints or service method.
			 existingUser.Points = points
			 existingUser.LastActivity = date
			 existingUser.UpdatedAt = time.Now()

			 err = i.userRepo.Update(ctx, existingUser)
			 if err != nil {
				 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Failed to update user %s: %v", results["totalRows"], msisdn, err))
				 continue
			 }
			 results["usersUpdated"] = results["usersUpdated"].(int) + 1
		 } else {
			 // Create new user
			 user.Points = points // Set initial points
			 err = i.userRepo.Create(ctx, user)
			 if err != nil {
				 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Failed to create user %s: %v", results["totalRows"], msisdn, err))
				 continue
			 }
			 results["usersCreated"] = results["usersCreated"].(int) + 1
		 }

		// Create topup if amount is provided (Assuming Topup model and repo exist)
		/* // Commenting out Topup creation as repo is not injected
		 if amount > 0 && i.topupRepo != nil {
			 topup := &models.Topup{
				 MSISDN:        msisdn,
				 Amount:        amount,
				 Channel:       "CSV_IMPORT",
				 Date:          date,
				 TransactionRef: fmt.Sprintf("CSV_IMPORT_%d_%s", time.Now().UnixNano(), primitive.NewObjectID().Hex()),
				 PointsEarned:  points,
				 Processed:     true, // Assuming processed as points are added directly
				 CreatedAt:     time.Now(),
				 UpdatedAt:     time.Now(),
			 }

			 err = i.topupRepo.Create(ctx, topup)
			 if err != nil {
				 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Failed to create topup for %s: %v", results["totalRows"], msisdn, err))
				 // continue // Don't stop user import if topup fails
			 }
			 results["topupsCreated"] = results["topupsCreated"].(int) + 1
		 }
		*/
	}

	return results, nil
}

// ImportPrizeStructures imports prize structures from a CSV file
func (i *CSVImporterEnhanced) ImportPrizeStructures(ctx context.Context, filePath string) (map[string]interface{}, error) {
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
	categoryIdx := findColumnIndex(header, []string{"Category", "Prize Category", "Prize Pool Breakdown"})
	 dailyAmountIdx := findColumnIndex(header, []string{"Daily Amount", "Daily Prize", "Daily Prizes [Mon - Fri]"})
	 weeklyAmountIdx := findColumnIndex(header, []string{"Weekly Amount", "Weekly Prize", "Saturday Prizes"})
	 // Assuming NumWinners is 1 unless specified? Or should it be a column?
	 // countIdx := findColumnIndex(header, []string{"Count", "Number of Prizes", "NumWinners"})

	 if categoryIdx == -1 || (dailyAmountIdx == -1 && weeklyAmountIdx == -1) {
		 return nil, fmt.Errorf("required columns (Category, Daily Amount/Weekly Amount) not found in CSV")
	 }

	// Initialize result counters
	results := map[string]interface{}{
		"totalRows":      0,
		"dailyPrizes":    0,
		"weeklyPrizes":   0,
		"errors":         []string{},
	}

	// Initialize prize structures
	var dailyPrizes []models.Prize // Changed from PrizeStructure
	var weeklyPrizes []models.Prize // Changed from PrizeStructure

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
		 if category == "" || strings.ToLower(category) == "total" { // Skip empty or total rows
			 continue
		 }
		 // Standardize category names (optional)
		 category = standardizeCategoryName(category)

		// Extract count (Assuming 1 winner per category row unless specified)
		count := 1
		/* // If count column exists
		 if countIdx != -1 && row[countIdx] != "" {
			 countVal, err := strconv.Atoi(strings.TrimSpace(row[countIdx]))
			 if err != nil {
				 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid count: %s", results["totalRows"], row[countIdx]))
			 } else {
				 count = countVal
			 }
		 }
		*/

		// Extract daily amount
		 if dailyAmountIdx != -1 && row[dailyAmountIdx] != "" {
			 amountStr := strings.TrimSpace(row[dailyAmountIdx])
			 // Remove currency symbol and commas if present
			 amountStr = strings.ReplaceAll(amountStr, "₦", "")
			 amountStr = strings.ReplaceAll(amountStr, "N", "")
			 amountStr = strings.ReplaceAll(amountStr, ",", "")

			 amount, err := strconv.ParseFloat(amountStr, 64)
			 if err != nil {
				 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid daily amount ", results["totalRows"], row[dailyAmountIdx], err))
			 } else {
				 dailyPrizes = append(dailyPrizes, models.Prize{ // Changed from PrizeStructure
					 Category: category,
					 Amount:   amount,
					 NumWinners: count, // Use NumWinners field
				 })
				 results["dailyPrizes"] = results["dailyPrizes"].(int) + 1
			 }
		 }

		// Extract weekly amount
		 if weeklyAmountIdx != -1 && row[weeklyAmountIdx] != "" {
			 amountStr := strings.TrimSpace(row[weeklyAmountIdx])
			 // Remove currency symbol and commas if present
			 amountStr = strings.ReplaceAll(amountStr, "₦", "")
			 amountStr = strings.ReplaceAll(amountStr, "N", "")
			 amountStr = strings.ReplaceAll(amountStr, ",", "")

			 amount, err := strconv.ParseFloat(amountStr, 64)
			 if err != nil {
				 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Row %d: Invalid weekly amount ", results["totalRows"], row[weeklyAmountIdx], err))
			 } else {
				 weeklyPrizes = append(weeklyPrizes, models.Prize{ // Changed from PrizeStructure
					 Category: category,
					 Amount:   amount,
					 NumWinners: count, // Use NumWinners field
				 })
				 results["weeklyPrizes"] = results["weeklyPrizes"].(int) + 1
			 }
		 }
	}

	// Save prize structures to system config using UpsertByKey
	// Corrected calls to UpsertByKey: removed description argument
	 if len(dailyPrizes) > 0 {
		 err = i.configRepo.UpsertByKey(ctx, "prize_structure_DAILY", dailyPrizes)
		 if err != nil {
			 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Failed to save daily prize structure: %v", err))
			 slog.Error("Failed to upsert daily prize structure", "error", err)
		 }
	 }

	 if len(weeklyPrizes) > 0 {
		 err = i.configRepo.UpsertByKey(ctx, "prize_structure_SATURDAY", weeklyPrizes)
		 if err != nil {
			 results["errors"] = append(results["errors"].([]string), fmt.Sprintf("Failed to save weekly prize structure: %v", err))
			 slog.Error("Failed to upsert weekly prize structure", "error", err)
		 }
	 }

	return results, nil
}

// --- Helper functions ---

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

	// Ensure it starts with country code if needed (e.g., 234 for Nigeria)
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
		"1/2/2006",        // M/D/YYYY
		"01/02/2006",      // MM/DD/YYYY
		"1-2-2006",        // M-D-YYYY
		"01-02-2006",      // MM-DD-YYYY
		"2/1/2006",        // D/M/YYYY (Common outside US)
		"02/01/2006",      // DD/MM/YYYY
		"2-1-2006",        // D-M-YYYY
		"02-01-2006",      // DD-MM-YYYY
		"Jan 2, 2006",
		"2 Jan 2006",
		"2006-01-02 15:04:05", // With time
		"1/2/2006 15:04:05",
		"01/02/2006 15:04:05",
		"2/1/2006 15:04:05",
		"02/01/2006 15:04:05",
		 time.RFC3339,      // Standard format
	}

	for _, format := range formats {
		 date, err := time.Parse(format, dateStr)
		 if err == nil {
			 return date, nil
		 }
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// calculatePoints calculates points based on topup amount (Matches DrawService logic)
func calculatePoints(amount float64) int {
	pointsToAdd := 0
	 if amount >= 1000 {
		 pointsToAdd = 10
	 } else {
		 pointsToAdd = int(amount / 100) // Integer division gives points per N100
	 }
	 return pointsToAdd
}

// standardizeCategoryName standardizes prize category names
func standardizeCategoryName(name string) string {
	name = strings.ToUpper(strings.ReplaceAll(name, " ", "_"))
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, "#", "")
	name = strings.ReplaceAll(name, "PRIZE", "")
	name = strings.Trim(name, "_")

	// Specific mappings
	 switch name {
	 case "JACKPOT_(1ST)", "JACKPOT_1ST", "1ST":
		 return "JACKPOT"
	 case "2ND":
		 return "2ND_PRIZE"
	 case "3RD":
		 return "3RD_PRIZE"
	 case "CONCESSION_1", "CONCESSION_2", "CONCESSION_3", "CONCESSION_4", "CONCESSION_5", "CONCESSION_6", "CONCESSION_7", "CONSOLATION":
		 return "CONSOLATION" // Group all consolations
	 default:
		 return name // Return standardized name if no specific mapping
	 }
}


