package config

import (
	"os"
	"pollapp/internal/adapter/redisadapter"
	"testing"
)

func TestLoadPostgres(t *testing.T) {
	os.Setenv("DATABASE_HOST", "test_db_host")
	os.Setenv("DATABASE_PORT", "5432")
	os.Setenv("DATABASE_USERNAME", "test_user")
	os.Setenv("DATABASE_PASSWORD", "test_password")
	os.Setenv("DATABASE_DBNAME", "test_db")
	os.Setenv("DATABASE_SSL_MODE", "disable")

	config := &Config{}

	config.loadPostgres()

	expectedPostgres := &Postgres{
		Host:     "test_db_host",
		Port:     5432,
		Username: "test_user",
		Password: "test_password",
		DBName:   "test_db",
		SSL:      "disable",
	}

	if config.Postgres.Host != expectedPostgres.Host {
		t.Errorf("Expected Postgres Host to be %s, got %s", expectedPostgres.Host, config.Postgres.Host)
	}
}

func TestLoadRedis(t *testing.T) {
	os.Setenv("REDIS_HOST", "test_redis_host")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "test_redis_password")
	os.Setenv("REDIS_DB", "0")

	config := &Config{}

	config.loadRedis()

	expectedRedis := &redisadapter.Config{
		Host:     "test_redis_host",
		Port:     6379,
		Password: "test_redis_password",
		DB:       0,
	}

	if config.Redis.Host != expectedRedis.Host {
		t.Errorf("Expected Redis Host to be %s, got %s", expectedRedis.Host, config.Redis.Host)
	}

}

func TestLoadPostgres_MissingHost(t *testing.T) {
	os.Unsetenv("DATABASE_HOST")

	config := &Config{}

	config.loadPostgres()

	if len(config.err) == 0 {
		t.Error("Expected an error for missing DATABASE_HOST environment variable")
	}
}

func TestLoadPostgres_InvalidPort(t *testing.T) {
	os.Setenv("DATABASE_HOST", "test_db_host")
	os.Setenv("DATABASE_PORT", "invalid_port")

	config := &Config{}

	config.loadPostgres()

	if len(config.err) == 0 {
		t.Error("Expected an error for invalid DATABASE_PORT value")
	}
}

func TestLoadRedis_MissingHost(t *testing.T) {
	os.Unsetenv("REDIS_HOST")

	config := &Config{}

	config.loadRedis()

	if len(config.err) == 0 {
		t.Error("Expected an error for missing REDIS_HOST environment variable")
	}
}

func TestLoadRedis_InvalidPort(t *testing.T) {
	os.Setenv("REDIS_HOST", "test_redis_host")
	os.Setenv("REDIS_PORT", "invalid_port")

	config := &Config{}

	config.loadRedis()

	if len(config.err) == 0 {
		t.Error("Expected an error for invalid REDIS_PORT value")
	}
}

func TestLoadRedis_InvalidDB(t *testing.T) {
	os.Setenv("REDIS_HOST", "test_redis_host")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "test_redis_password")
	os.Setenv("REDIS_DB", "invalid_db")

	config := &Config{}

	config.loadRedis()

	if len(config.err) == 0 {
		t.Error("Expected an error for invalid REDIS_DB value")
	}
}
