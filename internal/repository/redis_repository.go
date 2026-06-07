package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(addr, password string) (*RedisRepository, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisRepository{rdb: rdb}, nil
}

const cacheTTL = 7 * 24 * time.Hour

func (r *RedisRepository) SetCache(ctx context.Context, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, b, cacheTTL).Err()
}

func (r *RedisRepository) GetCache(ctx context.Context, key string, dest any) error {
	b, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
