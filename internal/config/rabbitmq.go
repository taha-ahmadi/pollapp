package config

import (
	"fmt"
	"pollapp/internal/adapter/broker/rabbitmq"
)

// RabbitMQ configuration loader
func (c *Config) loadRabbitMQ() *Config {
	host := getEnvWithDefault("RABBITMQ_HOST", "localhost")
	portStr := getEnvWithDefault("RABBITMQ_PORT", "5672")
	username := getEnvWithDefault("RABBITMQ_USERNAME", "guest")
	password := getEnvWithDefault("RABBITMQ_PASSWORD", "guest")
	vhost := getEnvWithDefault("RABBITMQ_VHOST", "/")

	port, err := parseInt(portStr)
	if err != nil {
		c.err = append(c.err, fmt.Errorf("cannot parse RABBITMQ_PORT: %v", err))
	}

	c.RabbitMQ = rabbitmq.Config{
		Host:        host,
		Port:        port,
		Username:    username,
		Password:    password,
		VirtualHost: vhost,
		QueueOptions: &rabbitmq.QueueOptions{
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
		},
		ConsumerOptions: &rabbitmq.ConsumerOptions{
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
		},
	}

	return c
}
