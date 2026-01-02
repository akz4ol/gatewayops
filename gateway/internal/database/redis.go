// Package database provides database connection management.
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
	cfg    config.RedisConfig
}

// NewRedis creates a new Redis connection.
func NewRedis(cfg config.RedisConfig, logger zerolog.Logger) (*Redis, error) {
	logger.Info().
		Str("url", maskRedisURL(cfg.URL)).
		Int("pool_size", cfg.PoolSize).
		Msg("Connecting to Redis")

	// Parse Redis URL
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	// Configure client options
	opts.MaxRetries = cfg.MaxRetries
	opts.PoolSize = cfg.PoolSize
	opts.MinIdleConns = cfg.MinIdleConns

	client := redis.NewClient(opts)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	logger.Info().Msg("Redis connected successfully")

	return &Redis{
		Client: client,
		logger: logger,
		cfg:    cfg,
	}, nil
}

// Close closes the Redis connection.
func (r *Redis) Close() error {
	if r.Client != nil {
		r.logger.Info().Msg("Closing Redis connection")
		return r.Client.Close()
	}
	return nil
}

// Health checks if Redis is healthy.
func (r *Redis) Health() bool {
	if r.Client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.Client.Ping(ctx).Err(); err != nil {
		r.logger.Warn().Err(err).Msg("Redis health check failed")
		return false
	}
	return true
}

// Ready checks if Redis is ready to accept commands.
func (r *Redis) Ready() bool {
	return r.Health()
}

// Get retrieves a value by key.
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Set sets a value with optional expiration.
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Del deletes one or more keys.
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Incr increments a key's value.
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Expire sets a key's expiration.
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// TTL returns the remaining time to live of a key.
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Exists checks if keys exist.
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// HSet sets a hash field.
func (r *Redis) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet gets a hash field.
func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields.
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// XAdd adds to a stream (for trace publishing).
func (r *Redis) XAdd(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	return r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()
}

// XRead reads from a stream.
func (r *Redis) XRead(ctx context.Context, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	return r.Client.XRead(ctx, &redis.XReadArgs{
		Streams: streams,
		Count:   count,
		Block:   block,
	}).Result()
}

// Publish publishes a message to a channel.
func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels.
func (r *Redis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}

// Pool returns the connection pool stats.
func (r *Redis) Pool() *redis.PoolStats {
	if r.Client == nil {
		return nil
	}
	stats := r.Client.PoolStats()
	return stats
}

// maskRedisURL masks sensitive information in the Redis URL for logging.
func maskRedisURL(url string) string {
	if len(url) > 20 {
		return url[:15] + "..."
	}
	return "redis://***"
}
