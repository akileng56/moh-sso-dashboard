package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(host, port, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})
}

func PingRedis(client *redis.Client) error {
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	return err
}
