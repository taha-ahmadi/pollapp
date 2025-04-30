package config

import (
	"fmt"
	"pollapp/internal/adapter/redisadapter"

	"github.com/joho/godotenv"
)

// Config is the configuration for the application containing the Postgres, Server, JWT and Redis configurations
type Config struct {
	Application *Application
	Postgres    *Postgres
	Server      *Server
	Redis       *redisadapter.Config
	err         []error
}

// New creates a new Config object
func New() *Config {
	var c Config
	c.loadApplication().loadPostgres().loadServer().loadRedis()
	return &c
}

// LoadConfig loads the configuration from the environment variables
func LoadConfig(pathFile string) (*Config, error) {
	err := godotenv.Load(pathFile)
	if err != nil {
		return nil, fmt.Errorf("cannot .env file in path: %s, err: %v", pathFile, err)
	}

	c := New()
	if len(c.err) > 0 {
		return nil, fmt.Errorf("cannot load configuration, err: %v", c.err)
	}

	return c, nil
}
