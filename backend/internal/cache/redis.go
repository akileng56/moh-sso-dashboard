package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int

	// Timeouts
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRedisClient creates a configured Redis client
func NewRedisClient(cfg RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
}

// Ping checks Redis connectivity
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}

// MustPing panics on Redis connection failure (useful during startup)
func MustPing(ctx context.Context, client *redis.Client) {
	if err := Ping(ctx, client); err != nil {
		panic(fmt.Sprintf("❌ failed to connect to redis: %v", err))
	}
}
