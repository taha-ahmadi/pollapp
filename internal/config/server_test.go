package config

import (
	"os"
	"testing"
)

func TestLoadServer(t *testing.T) {
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_SSL", "true")

	config := &Config{}

	config.loadServer()

	expectedServer := &Server{
		Host: "localhost",
		Port: 8080,
		SSL:  true,
	}

	if config.Server.Host != expectedServer.Host {
		t.Errorf("Expected Host to be %s, got %s", expectedServer.Host, config.Server.Host)
	}

	if config.Server.Port != expectedServer.Port {
		t.Errorf("Expected Port to be %d, got %d", expectedServer.Port, config.Server.Port)
	}

	if config.Server.SSL != expectedServer.SSL {
		t.Errorf("Expected SSL to be %t, got %t", expectedServer.SSL, config.Server.SSL)
	}
}
func TestLoadServer_MissingHost(t *testing.T) {
	os.Unsetenv("SERVER_HOST")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_SSL", "true")

	config := &Config{}

	config.loadServer()

	if len(config.err) == 0 {
		t.Error("Expected an error for missing SERVER_HOST environment variable")
	}
}

func TestLoadServer_MissingPort(t *testing.T) {
	os.Setenv("SERVER_HOST", "localhost")
	os.Unsetenv("SERVER_PORT")
	os.Setenv("SERVER_SSL", "true")

	config := &Config{}

	config.loadServer()

	if len(config.err) == 0 {
		t.Error("Expected an error for missing SERVER_PORT environment variable")
	}
}

func TestLoadServer_InvalidPort(t *testing.T) {
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("SERVER_PORT", "invalid_port")
	os.Setenv("SERVER_SSL", "true")

	config := &Config{}

	config.loadServer()

	if len(config.err) == 0 {
		t.Error("Expected an error for invalid SERVER_PORT")
	}
}

func TestLoadServer_MissingSSL(t *testing.T) {
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("SERVER_PORT", "8080")
	os.Unsetenv("SERVER_SSL")

	config := &Config{}

	config.loadServer()

	if len(config.err) == 0 {
		t.Error("Expected an error for missing SERVER_SSL environment variable")
	}
}

func TestLoadServer_InvalidSSL(t *testing.T) {
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_SSL", "invalid_bool")

	config := &Config{}

	config.loadServer()

	if len(config.err) == 0 {
		t.Error("Expected an error for invalid SERVER_SSL value")
	}
}
