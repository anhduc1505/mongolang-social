package database

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"golang-project/static"
)

// Config represents database connection configuration
type Config struct {
	ConnectionString string
	DatabaseName     string
}

// ParseConnectionString parses a connection string and extracts database name
func ParseConnectionString(connectionString string) (*Config, error) {
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

	return &Config{
		ConnectionString: connectionString,
		DatabaseName:     databaseName,
	}, nil
}

// BuildConnectionStringFromEnv builds connection string from environment variables
func BuildConnectionStringFromEnv() (string, string, error) {
	// Check if connection string is provided directly
	connectionString := viper.GetString(static.EnvMongoConnectionString)
	if connectionString != "" {
		config, err := ParseConnectionString(connectionString)
		if err != nil {
			return "", "", err
		}
		return config.ConnectionString, config.DatabaseName, nil
	}

	// Build connection string from individual components
	host := viper.GetString(static.EnvMongoHost)
	if host == "" {
		host = "localhost"
	}

	port := viper.GetString(static.EnvMongoPort)
	if port == "" {
		port = "27017"
	}

	username := viper.GetString(static.EnvMongoUsername)
	password := viper.GetString(static.EnvMongoPassword)
	database := viper.GetString(static.EnvMongoDatabase)
	if database == "" {
		database = "go"
	}

	authSource := viper.GetString(static.EnvMongoAuthSource)
	if authSource == "" {
		authSource = "admin"
	}

	var connectionStr string
	if username != "" && password != "" {
		// Authenticated connection - URL encode password
		encodedPassword := url.QueryEscape(password)
		connectionStr = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
			username, encodedPassword, host, port, database, authSource)
	} else {
		// Unauthenticated connection
		connectionStr = fmt.Sprintf("mongodb://%s:%s/%s", host, port, database)
	}

	return connectionStr, database, nil
}

// GetPoolConfig gets connection pool configuration from environment
func GetPoolConfig() (maxPoolSize, minPoolSize uint64, maxConnIdleTime, maxConnLifetime string) {
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
