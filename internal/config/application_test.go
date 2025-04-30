package config

import (
	"os"
	"testing"
)

func TestLoadApplication(t *testing.T) {
	os.Setenv("APPLICATION_RUN_MODE", "development")

	config := &Config{}

	config.loadApplication()

	expectedApplication := &Application{
		RunMode: "development",
	}

	if config.Application.RunMode != expectedApplication.RunMode {
		t.Errorf("Expected Application RunMode to be %s, got %s", expectedApplication.RunMode, config.Application.RunMode)
	}
}

func TestLoadApplication_MissingRunMode(t *testing.T) {
	os.Unsetenv("APPLICATION_RUN_MODE")

	config := &Config{}

	config.loadApplication()

	if len(config.err) == 0 {
		t.Error("Expected an error for missing APPLICATION_RUN_MODE environment variable")
	}
}
