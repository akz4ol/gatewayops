package database

import (
	"context"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Redis wraps the Redis client.
type Redis struct {
	Client *redis.Client
	logger zerolog.Logger
}

// NewRedis creates a new Redis connection.
func NewRedis(cfg config.RedisConfig, logger zerolog.Logger) (*Redis, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	// Apply additional configuration
	opt.MaxRetries = cfg.MaxRetries
	opt.PoolSize = cfg.PoolSize
	opt.MinIdleConns = cfg.MinIdleConns

	client := redis.NewClient(opt)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	logger.Info().
		Str("url", maskRedisURL(cfg.URL)).
		Int("pool_size", cfg.PoolSize).
		Msg("Connected to Redis")

	return &Redis{
		Client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis connection.
func (r *Redis) Close() error {
	return r.Client.Close()
}

// Health checks if Redis is healthy.
func (r *Redis) Health() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.Client.Ping(ctx).Err() == nil
}

// Ready checks if Redis is ready to accept commands.
func (r *Redis) Ready() bool {
	return r.Health()
}

// maskRedisURL masks sensitive parts of a Redis URL.
func maskRedisURL(url string) string {
	if len(url) > 20 {
		return url[:20] + "..."
	}
	return url
}
