package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Server is the configuration for the server
type Server struct {
	Host string
	Port uint64
	SSL  bool
}

// loadServer loads the server configuration from the environment variables
func (c *Config) loadServer() *Config {
	host, ok := os.LookupEnv("SERVER_HOST")
	if !ok {
		c.err = append(c.err, fmt.Errorf("SERVER_HOST environment variable not set"))
		return c
	}

	port, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		c.err = append(c.err, fmt.Errorf("SERVER_PORT environment variable not set"))
		return c
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		c.err = append(c.err, fmt.Errorf("failed to parse PORT: %w", err))
		return c
	}

	ssl, ok := os.LookupEnv("SERVER_SSL")
	if !ok {
		c.err = append(c.err, fmt.Errorf("SERVER_SSL environment variable not set"))
		return c
	}

	var sslBool bool
	if strings.ToLower(ssl) == "true" {
		sslBool = true
	} else if strings.ToLower(ssl) == "false" {
		sslBool = false
	} else {
		c.err = append(c.err, fmt.Errorf("failed to parse SSL: %w", err))
		return c
	}

	c.Server = &Server{
		Host: host,
		Port: uint64(portInt),
		SSL:  sslBool,
	}

	return c
}
