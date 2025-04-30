package redisdb

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	NotExist = errors.New("entity not exist")
)

// Get gets the value of a key
func (r *DB) Get(ctx context.Context, key string) (string, error) {
	data, err := r.Client().Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", NotExist
	} else if err != nil {
		return "", err
	}

	return data, nil
}

// Set sets the value of a key
func (r *DB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client().Set(ctx, key, data, expiration).Err()
}

// Del deletes a key
func (r *DB) Del(ctx context.Context, key string) error {
	return r.Client().Del(ctx, key).Err()
}

// Exist checks if a key exists
func (r *DB) Exist(ctx context.Context, key string) (bool, error) {
	_, err := r.Client().Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
