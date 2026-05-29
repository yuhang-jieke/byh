package redis

import (
	"context"
	"fmt"

	"driver/taketaxi/pkg/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Sprintf("failed to connect to Redis: %v", err))
	}
	return rdb
}
