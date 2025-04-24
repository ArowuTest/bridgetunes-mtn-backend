package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client represents a MongoDB client
type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewClient creates a new MongoDB client
func NewClient(uri string) (*Client, error) {
	// Create client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

// Database returns a database
func (c *Client) Database(name string) *mongo.Database {
	if c.db == nil || c.db.Name() != name {
		c.db = c.client.Database(name)
	}
	return c.db
}

// Disconnect disconnects from MongoDB
func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
