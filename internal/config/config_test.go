package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("../../.env.docker")

	assert.Nil(t, err, "Expected no error")
	assert.NotNil(t, config, "Expected non-nil Config")

	assert.NotNil(t, config.Postgres, "Expected non-nil Postgres configuration")
	assert.NotNil(t, config.Server, "Expected non-nil Server configuration")
	assert.NotNil(t, config.Redis, "Expected non-nil Redis configuration")
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	config, err := LoadConfig("non_existent_file.env")

	assert.Error(t, err, "Expected an error")
	assert.Nil(t, config, "Expected nil Config")
}
