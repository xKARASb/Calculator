package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host string `env:"REDIS_HOST" env-default:"localhost"`
	Port string `env:"REDIS_PORT" env-default:"6379"`
}

type RedisClient struct {
	Client *redis.Client
}

func New(config RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.Host, config.Port),
	})
	return &RedisClient{Client: client}
}

func (c *RedisClient) Ping(ctx context.Context) (string, error) {
	return c.Client.Ping(ctx).Result()
}

func (c *RedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

func (c *RedisClient) SAdd(ctx context.Context, key string, members ...string) error {
	return c.Client.SAdd(ctx, key, members).Err()
}

func (c *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.Client.SMembers(ctx, key).Result()
}
