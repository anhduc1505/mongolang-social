package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrUninitializedDatabase = errors.New("database instance is not initialized")
)

// Connection represents the database connection
type Connection interface {
	Connect() (*mongo.Client, error)
	Disconnect() error
	GetDatabase() *mongo.Database
	Ping() error
}

// connection is an implementation of the database connection
type connection struct {
	connectionString string
	databaseName     string
	client           *mongo.Client
	database         *mongo.Database
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewConnection creates and returns a database connection instance
func NewConnection(connectionString, databaseName string) Connection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return &connection{
		connectionString: connectionString,
		databaseName:     databaseName,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// NewConnectionFromEnv creates and returns a database connection instance from environment variables
func NewConnectionFromEnv() (Connection, error) {
	connectionString, databaseName, err := BuildConnectionStringFromEnv()
	if err != nil {
		return nil, err
	}

	return NewConnection(connectionString, databaseName), nil
}

// Connect initializes a new MongoDB client
func (c *connection) Connect() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(c.connectionString)

	// Set connection pool options from environment
	maxPoolSize, minPoolSize, maxConnIdleTime, _ := GetPoolConfig()
	clientOptions.SetMaxPoolSize(maxPoolSize)
	clientOptions.SetMinPoolSize(minPoolSize)

	// Parse time durations
	if idleTime, err := time.ParseDuration(maxConnIdleTime); err == nil {
		clientOptions.SetMaxConnIdleTime(idleTime)
	} else {
		clientOptions.SetMaxConnIdleTime(30 * time.Minute) // default
	}

	// Note: SetMaxConnLifetime is not available in this version of the MongoDB driver
	// Connection lifetime is managed by the MongoDB server

	client, err := mongo.Connect(c.ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	c.client = client
	c.database = client.Database(c.databaseName)

	return client, nil
}

// Disconnect closes the current MongoDB client
func (c *connection) Disconnect() error {
	if c.client == nil {
		return nil
	}

	if c.cancel != nil {
		c.cancel()
	}

	return c.client.Disconnect(c.ctx)
}

// GetDatabase returns the MongoDB database instance
func (c *connection) GetDatabase() *mongo.Database {
	return c.database
}

// Ping verifies if the current MongoDB client is active and healthy
func (c *connection) Ping() error {
	if c.client == nil {
		return ErrUninitializedDatabase
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.client.Ping(ctx, nil)
}
