package scripts

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bridgetunes/mtn-backend/internal/models"
	"github.com/bridgetunes/mtn-backend/internal/utils"
	"github.com/bridgetunes/mtn-backend/pkg/mongodb"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ImportCSVData imports data from a CSV file into MongoDB
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get MongoDB connection string from environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is required")
	}

	// Get database name from environment
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "bridgetunes"
	}

	// Get CSV file path from command line arguments
	if len(os.Args) < 2 {
		log.Fatal("CSV file path is required as a command line argument")
	}
	csvFilePath := os.Args[1]

	// Connect to MongoDB
	client, err := mongodb.NewClient(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Get database
	db := client.Database(dbName)

	// Import data
	err = importData(db, csvFilePath)
	if err != nil {
		log.Fatalf("Failed to import data: %v", err)
	}

	log.Println("Data imported successfully")
}

// importData imports data from a CSV file into MongoDB
func importData(db *mongo.Database, csvFilePath string) error {
	// Open CSV file
	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	// Parse CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV file: %v", err)
	}

	// Check if CSV file has header
	if len(records) < 2 {
		return fmt.Errorf("CSV file is empty or has only header")
	}

	// Get collections
	usersCollection := db.Collection("users")
	topupsCollection := db.Collection("topups")

	// Process records
	for i, record := range records {
		// Skip header
		if i == 0 {
			continue
		}

		// Check if record has enough fields
		if len(record) < 3 {
			log.Printf("Warning: Record %d has less than 3 fields, skipping", i)
			continue
		}

		// Parse record
		msisdn := record[0]
		amount, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Printf("Warning: Record %d has invalid amount, skipping", i)
			continue
		}
		dateStr := record[2]
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Printf("Warning: Record %d has invalid date format, skipping", i)
			continue
		}

		// Calculate points
		points := utils.CalculatePoints(amount)

		// Check if user exists
		var user models.User
		err = usersCollection.FindOne(context.Background(), bson.M{"msisdn": msisdn}).Decode(&user)
		if err != nil {
			// User doesn't exist, create a new one
			user = models.User{
				MSISDN:        msisdn,
				OptInStatus:   true,
				OptInDate:     time.Now(),
				OptInChannel:  "CSV_IMPORT",
				Points:        points,
				IsBlacklisted: false,
				LastActivity:  time.Now(),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			_, err = usersCollection.InsertOne(context.Background(), user)
			if err != nil {
				log.Printf("Warning: Failed to create user for record %d: %v", i, err)
				continue
			}
		} else {
			// User exists, update points
			_, err = usersCollection.UpdateOne(
				context.Background(),
				bson.M{"msisdn": msisdn},
				bson.M{
					"$inc": bson.M{"points": points},
					"$set": bson.M{
						"lastActivity": time.Now(),
						"updatedAt":    time.Now(),
					},
				},
			)
			if err != nil {
				log.Printf("Warning: Failed to update user for record %d: %v", i, err)
				continue
			}
		}

		// Create topup
		topup := models.Topup{
			MSISDN:         msisdn,
			Amount:         amount,
			Channel:        "CSV_IMPORT",
			Date:           date,
			TransactionRef: fmt.Sprintf("CSV_IMPORT_%d", i),
			PointsEarned:   points,
			Processed:      true,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		_, err = topupsCollection.InsertOne(context.Background(), topup)
		if err != nil {
			log.Printf("Warning: Failed to create topup for record %d: %v", i, err)
			continue
		}
	}

	return nil
}
