package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	WebSocket WebSocketConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	Logging   LoggingConfig
}

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret                 string
	Expiration             time.Duration
	RefreshTokenExpiration time.Duration
}

type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	MaxMessageSize  int64
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxConnPerUser  int
}

type RateLimitConfig struct {
	RequestsPerMinute int
	Enabled           bool
}

type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
}

type LoggingConfig struct {
	Level string
}

func Load() (*Config, error) {
	godotenv.Load()

	jwtExp, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}

	refreshExp, err := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRATION", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_EXPIRATION: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5984"),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "inkdown"),
		},
		JWT: JWTConfig{
			Secret:                 getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			Expiration:             jwtExp,
			RefreshTokenExpiration: refreshExp,
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  getEnvAsInt("WS_READ_BUFFER_SIZE", 4096),
			WriteBufferSize: getEnvAsInt("WS_WRITE_BUFFER_SIZE", 4096),
			MaxMessageSize:  int64(getEnvAsInt("WS_MAX_MESSAGE_SIZE", 10485760)),
			WriteWait:       10 * time.Second,
			PongWait:        60 * time.Second,
			PingPeriod:      54 * time.Second,
			MaxConnPerUser:  getEnvAsInt("WS_MAX_CONN_PER_USER", 5),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			Enabled:           getEnvAsBool("RATE_LIMIT_ENABLED", true),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
