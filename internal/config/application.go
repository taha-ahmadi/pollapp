package config

import (
	"fmt"
	"os"
)

type Application struct {
	RunMode string
}

func (c *Config) loadApplication() *Config {
	runMode, ok := os.LookupEnv("APPLICATION_RUN_MODE")
	if !ok {
		c.err = append(c.err, fmt.Errorf("APPLICATION_RUN_MODE not found"))
		return c
	}

	c.Application = &Application{
		RunMode: runMode,
	}
	return c
}
