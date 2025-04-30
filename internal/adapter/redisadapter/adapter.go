package redisadapter

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Adapter is the Redis adapter
type Adapter struct {
	client *redis.Client
}

// New creates a new Redis adapter and pings the redis server
func New(config Config) (Adapter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	cmd := rdb.Ping(context.TODO())
	if cmd.Err() != nil {
		return Adapter{}, cmd.Err()
	}

	return Adapter{client: rdb}, nil
}

// Client returns the Redis client
func (a Adapter) Client() *redis.Client {
	return a.client
}
