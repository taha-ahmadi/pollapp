package config

import (
	"fmt"
	"os"
	"strings"

	"pollapp/internal/adapter/broker/rabbitmq"
	"pollapp/internal/adapter/redisadapter"

	"github.com/joho/godotenv"
)

// Config is the configuration for the application containing the Postgres, Server, JWT and Redis configurations
type Config struct {
	Application *Application
	Postgres    *Postgres
	Server      *Server
	Redis       *redisadapter.Config
	RabbitMQ    rabbitmq.Config
	err         []error
}

// New creates a new Config object
func New() *Config {
	var c Config
	c.loadApplication().loadPostgres().loadServer().loadRedis().loadRabbitMQ()
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

// Helper function to get an environment variable or use a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// Load aliases LoadConfig for backward compatibility
func Load() (*Config, error) {
	return LoadConfig(".env")
}
