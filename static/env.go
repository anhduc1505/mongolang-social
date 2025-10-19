package static

// Server environment variable name
const (
	EnvServerEnv     = "SERVER_ENV"
	EnvServerAddress = "SERVER_ADDRESS"
)

// Database environment variable name
const (
	EnvMongoConnectionString = "DB_CONNECTION_STRING"
	EnvMongoHost             = "MONGO_HOST"
	EnvMongoPort             = "MONGO_PORT"
	EnvMongoUsername         = "MONGO_USERNAME"
	EnvMongoPassword         = "MONGO_PASSWORD"
	EnvMongoDatabase         = "MONGO_DATABASE"
	EnvMongoAuthSource       = "MONGO_AUTH_SOURCE"
	EnvMongoMaxPoolSize      = "MONGO_MAX_POOL_SIZE"
	EnvMongoMinPoolSize      = "MONGO_MIN_POOL_SIZE"
	EnvMongoMaxConnIdleTime  = "MONGO_MAX_CONN_IDLE_TIME"
	EnvMongoMaxConnLifetime  = "MONGO_MAX_CONN_LIFETIME"
)

// Auth environment variable name
const (
	EnvAuthType     = "AUTH_TYPE"
	EnvAuthSecret   = "AUTH_SECRET"
	EnvAuthLifeTime = "AUTH_LIFE_TIME"
	EnvAuthAudience = "AUTH_AUDIENCE"
	EnvAuthIssuer   = "AUTH_ISSUER"
	EnvAuthSubject  = "AUTH_SUBJECT"
)
