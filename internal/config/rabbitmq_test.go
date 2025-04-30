package config

import (
	"os"
	"testing"
)

func TestLoadRabbitMQ(t *testing.T) {
	// Set environment variables for RabbitMQ
	os.Setenv("RABBITMQ_HOST", "rabbitmq-server")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USERNAME", "test_user")
	os.Setenv("RABBITMQ_PASSWORD", "test_password")
	os.Setenv("RABBITMQ_VHOST", "/test")

	config := &Config{}

	config.loadRabbitMQ()

	// Verify the configuration is loaded correctly
	if config.RabbitMQ.Host != "rabbitmq-server" {
		t.Errorf("Expected RabbitMQ Host to be %s, got %s", "rabbitmq-server", config.RabbitMQ.Host)
	}

	if config.RabbitMQ.Port != 5672 {
		t.Errorf("Expected RabbitMQ Port to be %d, got %d", 5672, config.RabbitMQ.Port)
	}

	if config.RabbitMQ.Username != "test_user" {
		t.Errorf("Expected RabbitMQ Username to be %s, got %s", "test_user", config.RabbitMQ.Username)
	}

	if config.RabbitMQ.Password != "test_password" {
		t.Errorf("Expected RabbitMQ Password to be %s, got %s", "test_password", config.RabbitMQ.Password)
	}

	if config.RabbitMQ.VirtualHost != "/test" {
		t.Errorf("Expected RabbitMQ VirtualHost to be %s, got %s", "/test", config.RabbitMQ.VirtualHost)
	}

	// Test default queue options
	if !config.RabbitMQ.QueueOptions.Durable {
		t.Error("Expected QueueOptions.Durable to be true")
	}

	if config.RabbitMQ.QueueOptions.AutoDelete {
		t.Error("Expected QueueOptions.AutoDelete to be false")
	}
}

func TestLoadRabbitMQ_InvalidPort(t *testing.T) {
	// Set valid environment variables except for port
	os.Setenv("RABBITMQ_HOST", "rabbitmq-server")
	os.Setenv("RABBITMQ_PORT", "invalid-port")
	os.Setenv("RABBITMQ_USERNAME", "test_user")
	os.Setenv("RABBITMQ_PASSWORD", "test_password")
	os.Setenv("RABBITMQ_VHOST", "/test")

	config := &Config{}

	config.loadRabbitMQ()

	// Verify that an error was recorded
	if len(config.err) == 0 {
		t.Error("Expected an error for invalid RABBITMQ_PORT value")
	}
}

func TestLoadRabbitMQ_Defaults(t *testing.T) {
	// Unset all environment variables to test defaults
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_USERNAME")
	os.Unsetenv("RABBITMQ_PASSWORD")
	os.Unsetenv("RABBITMQ_VHOST")

	config := &Config{}

	config.loadRabbitMQ()

	// Verify default values are used
	if config.RabbitMQ.Host != "localhost" {
		t.Errorf("Expected default RabbitMQ Host to be %s, got %s", "localhost", config.RabbitMQ.Host)
	}

	if config.RabbitMQ.Port != 5672 {
		t.Errorf("Expected default RabbitMQ Port to be %d, got %d", 5672, config.RabbitMQ.Port)
	}

	if config.RabbitMQ.Username != "guest" {
		t.Errorf("Expected default RabbitMQ Username to be %s, got %s", "guest", config.RabbitMQ.Username)
	}

	if config.RabbitMQ.Password != "guest" {
		t.Errorf("Expected default RabbitMQ Password to be %s, got %s", "guest", config.RabbitMQ.Password)
	}

	if config.RabbitMQ.VirtualHost != "/" {
		t.Errorf("Expected default RabbitMQ VirtualHost to be %s, got %s", "/", config.RabbitMQ.VirtualHost)
	}
}
