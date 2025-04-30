package redisdb

import (
	"github.com/redis/go-redis/v9"
)

// DB is the model for redis database
type DB struct {
	client *redis.Client
}

// New creates a new instance of redis database
func New(client *redis.Client) *DB {
	return &DB{client: client}
}

// Client returns the connection to the redis database
func (m *DB) Client() *redis.Client {
	return m.client
}
