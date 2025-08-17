package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Processor ProcessorConfig
	Database  DatabaseConfig
	Redis     RedisConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// ProcessorConfig holds payment processor configuration
type ProcessorConfig struct {
	DefaultURL      string
	FallbackURL     string
	Timeout         time.Duration
	MaxRetries      int
	CircuitBreaker  CircuitBreakerConfig
	QueueBufferSize int
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	ConnectionString string
	MaxConnections   int
	ConnectTimeout   time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string
	PoolSize int
	QueueKey string
	UuidTTL  time.Duration
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxRequests  uint32
	Interval     time.Duration
	Timeout      time.Duration
	FailureRatio float64
	MinRequests  uint32
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", ":8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			ConnectionString: getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/payments"),
			MaxConnections:   getIntEnv("DATABASE_MAX_CONNECTIONS", 10),
			ConnectTimeout:   getDurationEnv("DATABASE_CONNECT_TIMEOUT", 5*time.Second),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://redis:6379/0"),
			PoolSize: getIntEnv("REDIS_POOL_SIZE", 10),
			QueueKey: getEnv("REDIS_QUEUE_KEY", "payment_queue"),
			UuidTTL:  getDurationEnv("REDIS_UUID_TTL", 24*time.Hour),
		},
		Processor: ProcessorConfig{
			DefaultURL:      getEnv("PROCESSOR_DEFAULT_URL", "http://payment-processor-default:8080"),
			FallbackURL:     getEnv("PROCESSOR_FALLBACK_URL", "http://payment-processor-fallback:8080"),
			Timeout:         getDurationEnv("PROCESSOR_TIMEOUT", 5*time.Second),
			MaxRetries:      getIntEnv("PROCESSOR_MAX_RETRIES", 3),
			QueueBufferSize: getIntEnv("QUEUE_BUFFER_SIZE", 100),
			CircuitBreaker: CircuitBreakerConfig{
				MaxRequests:  uint32(getIntEnv("CB_MAX_REQUESTS", 3)),
				Interval:     getDurationEnv("CB_INTERVAL", 10*time.Second),
				Timeout:      getDurationEnv("CB_TIMEOUT", 30*time.Second),
				FailureRatio: getFloatEnv("CB_FAILURE_RATIO", 0.5),
				MinRequests:  uint32(getIntEnv("CB_MIN_REQUESTS", 5)),
			},
		},
	}
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
