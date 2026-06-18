package repository

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(addr, password string) *RedisRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  200 * time.Millisecond,
		WriteTimeout: 200 * time.Millisecond,
		MaxRetries:   -1,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis unavailable at startup, continuing without cache", "error", err)
	}

	return &RedisRepository{rdb: rdb}
}

const cacheTTL = 7 * 24 * time.Hour

func sanitizeKey(key string) string {
	return strings.ReplaceAll(key, ":", "-")
}

func (r *RedisRepository) SetCache(ctx context.Context, key string, value any) error {
	key = "location:" + sanitizeKey(key)
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, b, cacheTTL).Err()
}

func (r *RedisRepository) GetCache(ctx context.Context, key string, dest any) error {
	key = "location:" + sanitizeKey(key)
	b, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
