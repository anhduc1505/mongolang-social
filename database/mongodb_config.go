package database

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang-project/static"
)

// MongoDBConfig represents MongoDB connection configuration
type MongoDBConfig struct {
	ConnectionString string
	DatabaseName     string
}

// ParseMongoDBConnectionString parses a MongoDB connection string and extracts database name
func ParseMongoDBConnectionString(connectionString string) (*MongoDBConfig, error) {
	// Parse the connection string
	parsedURL, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}

	// Extract database name from path
	databaseName := strings.TrimPrefix(parsedURL.Path, "/")
	if databaseName == "" {
		// Default database name if not specified
		databaseName = "social-blog"
	}

	return &MongoDBConfig{
		ConnectionString: connectionString,
		DatabaseName:     databaseName,
	}, nil
}

// BuildMongoDBConnectionStringFromEnv builds MongoDB connection string from environment variables
func BuildMongoDBConnectionStringFromEnv() (string, string, error) {
	// Check if connection string is provided directly
	connectionString := os.Getenv(static.EnvMongoConnectionString)
	if connectionString != "" {
		config, err := ParseMongoDBConnectionString(connectionString)
		if err != nil {
			return "", "", err
		}
		return config.ConnectionString, config.DatabaseName, nil
	}

	// Build connection string from individual components
	host := os.Getenv(static.EnvMongoHost)
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv(static.EnvMongoPort)
	if port == "" {
		port = "27017"
	}

	username := os.Getenv(static.EnvMongoUsername)
	password := os.Getenv(static.EnvMongoPassword)
	database := os.Getenv(static.EnvMongoDatabase)
	if database == "" {
		database = "go"
	}

	authSource := os.Getenv(static.EnvMongoAuthSource)
	if authSource == "" {
		authSource = "admin"
	}

	var connectionStr string
	if username != "" && password != "" {
		// Authenticated connection
		connectionStr = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
			username, password, host, port, database, authSource)
	} else {
		// Unauthenticated connection
		connectionStr = fmt.Sprintf("mongodb://%s:%s/%s", host, port, database)
	}

	return connectionStr, database, nil
}

// GetMongoDBPoolConfig gets MongoDB connection pool configuration from environment
func GetMongoDBPoolConfig() (maxPoolSize, minPoolSize uint64, maxConnIdleTime, maxConnLifetime string) {
	maxPoolSizeStr := os.Getenv(static.EnvMongoMaxPoolSize)
	if maxPoolSizeStr != "" {
		if val, err := strconv.ParseUint(maxPoolSizeStr, 10, 32); err == nil {
			maxPoolSize = val
		}
	}
	if maxPoolSize == 0 {
		maxPoolSize = 100 // default
	}

	minPoolSizeStr := os.Getenv(static.EnvMongoMinPoolSize)
	if minPoolSizeStr != "" {
		if val, err := strconv.ParseUint(minPoolSizeStr, 10, 32); err == nil {
			minPoolSize = val
		}
	}
	if minPoolSize == 0 {
		minPoolSize = 5 // default
	}

	maxConnIdleTime = os.Getenv(static.EnvMongoMaxConnIdleTime)
	if maxConnIdleTime == "" {
		maxConnIdleTime = "30m" // default
	}

	maxConnLifetime = os.Getenv(static.EnvMongoMaxConnLifetime)
	if maxConnLifetime == "" {
		maxConnLifetime = "1h" // default
	}

	return maxPoolSize, minPoolSize, maxConnIdleTime, maxConnLifetime
}
