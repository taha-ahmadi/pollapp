package config

import (
	"fmt"
	"os"
	"pollapp/internal/adapter/redisadapter"
	"strconv"
)

// Postgres is the configuration for the Postgres database
type Postgres struct {
	Host     string
	Port     uint64
	Username string
	Password string
	DBName   string
	SSL      string
}

// DSN returns the Data Source Name for the Postgres database
func (p *Postgres) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.Username, p.Password, p.DBName, p.SSL)
}

// loadPostgres loads the Postgres configuration from the environment variables
func (c *Config) loadPostgres() *Config {
	host, ok := os.LookupEnv("DATABASE_HOST")
	if !ok {
		c.err = append(c.err, fmt.Errorf("DATABASE_HOST environment variable not set"))
		return c
	}

	port, ok := os.LookupEnv("DATABASE_PORT")
	if !ok {
		c.err = append(c.err, fmt.Errorf("DATABASE_HOST environment variable not set"))

		return c
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		c.err = append(c.err, fmt.Errorf("DATABASE_HOST environment variable not set"))
		return c
	}

	username, ok := os.LookupEnv("DATABASE_USERNAME")
	if !ok {
		c.err = append(c.err, fmt.Errorf("DATABASE_HOST environment variable not set"))
		return c
	}

	password, ok := os.LookupEnv("DATABASE_PASSWORD")
	if !ok {
		c.err = append(c.err, fmt.Errorf("DATABASE_HOST environment variable not set"))
		return c
	}

	dbName, ok := os.LookupEnv("DATABASE_DBNAME")
	if !ok {
		c.err = append(c.err, fmt.Errorf("DATABASE_HOST environment variable not set"))
		return c
	}

	sslMode, ok := os.LookupEnv("DATABASE_SSL_MODE")
	if !ok {
		c.err = append(c.err, fmt.Errorf("DATABASE_HOST environment variable not set"))
		return c
	}

	if sslMode != "disable" && sslMode != "require" {
		c.err = append(c.err, fmt.Errorf("DATABASE_SSL_MODE environment variable not set"))
		return c
	}

	c.Postgres = &Postgres{
		Host:     host,
		Port:     uint64(portInt),
		Username: username,
		Password: password,
		DBName:   dbName,
		SSL:      sslMode,
	}

	return c
}

// loadRedis loads the Redis configuration from the environment variables
func (c *Config) loadRedis() *Config {
	host, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		c.err = append(c.err, fmt.Errorf("REDIS_HOST environment variable not set"))
		return c
	}

	port, ok := os.LookupEnv("REDIS_PORT")
	if !ok {
		c.err = append(c.err, fmt.Errorf("REDIS_PORT environment variable not set"))
		return c
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		c.err = append(c.err, fmt.Errorf("failed to parse REDIS_PORT: %w", err))
		return c
	}

	password, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		c.err = append(c.err, fmt.Errorf("REDIS_PASSWORD environment variable not set"))
		return c
	}

	db, ok := os.LookupEnv("REDIS_DB")
	if !ok {
		c.err = append(c.err, fmt.Errorf("REDIS_DB environment variable not set"))
		return c
	}

	dbInt, err := strconv.Atoi(db)
	if err != nil {
		c.err = append(c.err, fmt.Errorf("failed to parse REDIS_DB: %w", err))
		return c
	}

	c.Redis = &redisadapter.Config{
		Host:     host,
		Port:     portInt,
		Password: password,
		DB:       dbInt,
	}

	return c
}
