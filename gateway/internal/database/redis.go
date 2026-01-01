package database

import (
	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/rs/zerolog"
)

// Redis wraps the Redis client.
type Redis struct {
	Client interface{} // Mock client for demo mode
	logger zerolog.Logger
}

// NewRedis creates a new Redis connection.
// In demo mode, returns a mock connection that always succeeds.
func NewRedis(cfg config.RedisConfig, logger zerolog.Logger) (*Redis, error) {
	logger.Info().
		Msg("Demo mode: Using mock Redis connection")

	return &Redis{
		Client: nil, // Mock - no actual Redis
		logger: logger,
	}, nil
}

// Close closes the Redis connection.
func (r *Redis) Close() error {
	return nil
}

// Health checks if Redis is healthy.
func (r *Redis) Health() bool {
	return true // Demo mode always healthy
}

// Ready checks if Redis is ready to accept commands.
func (r *Redis) Ready() bool {
	return true // Demo mode always ready
}
