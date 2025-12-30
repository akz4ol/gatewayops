// Package config handles configuration loading for the gateway.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the gateway.
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	ClickHouse ClickHouseConfig
	Auth       AuthConfig
	RateLimit  RateLimitConfig
	Logging    LoggingConfig
	MCPServers map[string]MCPServerConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port            string
	Env             string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds PostgreSQL configuration.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	URL          string
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

// ClickHouseConfig holds ClickHouse configuration.
type ClickHouseConfig struct {
	DSN string
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	BcryptCost int
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	DefaultRPM int
	Burst      int
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string
	Format string // json or console
}

// MCPServerConfig holds configuration for an MCP server.
type MCPServerConfig struct {
	Name       string
	URL        string
	Timeout    time.Duration
	MaxRetries int
	Pricing    MCPPricing
}

// MCPPricing holds pricing configuration for an MCP server.
type MCPPricing struct {
	PerCall        float64 `json:"per_call"`
	PerInputToken  float64 `json:"per_input_token"`
	PerOutputToken float64 `json:"per_output_token"`
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			Env:             getEnv("ENV", "development"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gatewayops?sslmode=disable"),
			MaxOpenConns:    getIntEnv("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL:          getEnv("REDIS_URL", "redis://localhost:6379"),
			MaxRetries:   getIntEnv("REDIS_MAX_RETRIES", 3),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
		},
		ClickHouse: ClickHouseConfig{
			DSN: getEnv("CLICKHOUSE_DSN", "clickhouse://localhost:9000/gatewayops"),
		},
		Auth: AuthConfig{
			BcryptCost: getIntEnv("API_KEY_BCRYPT_COST", 12),
		},
		RateLimit: RateLimitConfig{
			DefaultRPM: getIntEnv("RATE_LIMIT_DEFAULT_RPM", 1000),
			Burst:      getIntEnv("RATE_LIMIT_BURST", 50),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		MCPServers: make(map[string]MCPServerConfig),
	}

	// Load MCP servers from environment
	// Format: MCP_SERVER_{NAME}_URL, MCP_SERVER_{NAME}_TIMEOUT
	if mockURL := getEnv("MCP_SERVER_MOCK_URL", ""); mockURL != "" {
		cfg.MCPServers["mock"] = MCPServerConfig{
			Name:       "mock",
			URL:        mockURL,
			Timeout:    getDurationEnv("MCP_SERVER_MOCK_TIMEOUT", 30*time.Second),
			MaxRetries: getIntEnv("MCP_SERVER_MOCK_RETRIES", 3),
			Pricing: MCPPricing{
				PerCall: 0.001,
			},
		}
	}

	return cfg, nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
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

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}
