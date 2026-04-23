package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// Client wraps a *mongo.Client together with the target database name.
type Client struct {
	client *mongo.Client
	dbName string
}

// NewClient creates a Client from an already-connected *mongo.Client.
func NewClient(client *mongo.Client, dbName string) *Client {
	return &Client{client: client, dbName: dbName}
}

// Collection returns a handle to the named collection in the configured database.
func (c *Client) Collection(name string) *mongo.Collection {
	return c.client.Database(c.dbName).Collection(name)
}
