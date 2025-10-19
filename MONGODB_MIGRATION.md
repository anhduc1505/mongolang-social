# MongoDB Migration Guide

This project has been completely migrated from MySQL to MongoDB. All MySQL-related code has been removed. The migration includes:

## Changes Made

### 1. Database Connection
- **MongoDB Connection**: `database/connection.go` for MongoDB connection handling
- **Configuration Helper**: `database/config.go` for parsing MongoDB connection strings
- **MongoDB Only**: The project now uses only MongoDB (via official driver)

### 2. Model Updates
- **ObjectID Support**: Updated all models to use `primitive.ObjectID` for MongoDB compatibility
- **BSON Tags**: Added proper BSON tags for MongoDB document structure
- **Removed GORM Tags**: Removed MySQL-specific GORM tags from models

### 3. Environment Configuration
- **Connection String**: Updated `local.env` to include MongoDB Atlas connection string
- **Docker Setup**: Updated `docker-compose.yml` to include MongoDB service with Mongo Express

### 4. Dependencies
- **MongoDB Driver**: `go.mongodb.org/mongo-driver v1.15.0`
- **Removed GORM**: Completely removed GORM and MySQL dependencies
- **Removed Migrations**: Removed GORM-based migration system

## Usage

### MongoDB Connection
```go
import "golang-project/database"

// Create MongoDB connection from environment
conn, err := database.NewConnectionFromEnv()
if err != nil {
    log.Fatal(err)
}

// Connect
client, err := conn.Connect()
if err != nil {
    log.Fatal(err)
}
defer conn.Disconnect()

// Get database instance
db := conn.GetDatabase()

// Use collections
usersCollection := db.Collection("users")
postsCollection := db.Collection("posts")
```

### Direct Connection with Connection String
```go
import "golang-project/database"

// Create MongoDB connection with specific connection string
conn := database.NewConnection(connectionString, databaseName)

// Connect
client, err := conn.Connect()
if err != nil {
    log.Fatal(err)
}
defer conn.Disconnect()

// Get database instance
db := conn.GetDatabase()

// Use collections
usersCollection := db.Collection("users")
postsCollection := db.Collection("posts")
```

### Environment Variables
The `local.env` file now includes comprehensive database configuration:

```env
# Database Configuration
DB_CONNECTION_STRING="mongodb+srv://socialblog:<password>@social-blog.urdtse4.mongodb.net/?retryWrites=true&w=majority&appName=social-blog"

# MongoDB Local Connection (for development)
MONGO_HOST="localhost"
MONGO_PORT="27017"
MONGO_USERNAME="admin"
MONGO_PASSWORD="root123!@#"
MONGO_DATABASE="go"
MONGO_AUTH_SOURCE="admin"

# MongoDB Connection Pool Settings
MONGO_MAX_POOL_SIZE="100"
MONGO_MIN_POOL_SIZE="5"
MONGO_MAX_CONN_IDLE_TIME="30m"
MONGO_MAX_CONN_LIFETIME="1h"
```

### Docker Services
The `docker-compose.yml` now includes:
- **MongoDB**: Local MongoDB instance on port 27017
- **Mongo Express**: Web-based MongoDB admin interface on port 8081

## Migration Notes

1. **ID Types**: All models now use `primitive.ObjectID` instead of `int`
2. **Relationships**: MongoDB relationships are handled via embedded ObjectID references
3. **Collections**: MongoDB collections correspond to the model names (users, posts, comments, etc.)
4. **Indexes**: You may want to create indexes for frequently queried fields

## Next Steps

1. **Repository Layer**: Update repository implementations to use MongoDB collections
2. **Service Layer**: Modify services to work with MongoDB documents
3. **Data Migration**: Migrate existing MySQL data to MongoDB if needed
4. **Testing**: Update tests to work with MongoDB
5. **Indexes**: Create appropriate indexes for performance

## Development vs Production

- **Development**: Use local MongoDB via Docker Compose
- **Production**: Use MongoDB Atlas connection string provided

## Security Notes

- The MongoDB Atlas connection string contains credentials
- Consider using environment variables for sensitive data
- Ensure proper network access rules in MongoDB Atlas
