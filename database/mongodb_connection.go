package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConnection represents the MongoDB database connection
type MongoDBConnection interface {
	Connect() (*mongo.Client, error)
	Disconnect() error
	GetDatabase() *mongo.Database
	Ping() error
}

// mongodbConnection is an implementation of the MongoDB database connection
type mongodbConnection struct {
	connectionString string
	databaseName     string
	client           *mongo.Client
	database         *mongo.Database
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewMongoDBConnection creates and returns a MongoDB connection instance
func NewMongoDBConnection(connectionString, databaseName string) MongoDBConnection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return &mongodbConnection{
		connectionString: connectionString,
		databaseName:     databaseName,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// NewMongoDBConnectionFromEnv creates and returns a MongoDB connection instance from environment variables
func NewMongoDBConnectionFromEnv() (MongoDBConnection, error) {
	connectionString, databaseName, err := BuildMongoDBConnectionStringFromEnv()
	if err != nil {
		return nil, err
	}

	return NewMongoDBConnection(connectionString, databaseName), nil
}

// Connect initializes a new MongoDB client
func (c *mongodbConnection) Connect() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(c.connectionString)

	// Set connection pool options from environment
	maxPoolSize, minPoolSize, maxConnIdleTime, _ := GetMongoDBPoolConfig()
	clientOptions.SetMaxPoolSize(maxPoolSize)
	clientOptions.SetMinPoolSize(minPoolSize)

	// Parse time durations
	if idleTime, err := time.ParseDuration(maxConnIdleTime); err == nil {
		clientOptions.SetMaxConnIdleTime(idleTime)
	} else {
		clientOptions.SetMaxConnIdleTime(30 * time.Minute) // default
	}

	// Set Server API for Atlas compatibility
	// This ensures compatibility with Atlas and future MongoDB versions
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions.SetServerAPIOptions(serverAPI)

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
func (c *mongodbConnection) Disconnect() error {
	if c.client == nil {
		return nil
	}

	if c.cancel != nil {
		c.cancel()
	}

	return c.client.Disconnect(c.ctx)
}

// GetDatabase returns the MongoDB database instance
func (c *mongodbConnection) GetDatabase() *mongo.Database {
	return c.database
}

// Ping verifies if the current MongoDB client is active and healthy
func (c *mongodbConnection) Ping() error {
	if c.client == nil {
		return ErrUninitializedDatabase
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.client.Ping(ctx, nil)
}
