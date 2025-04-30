package rabbitmq

// Config holds the configuration for RabbitMQ
type Config struct {
	Host            string
	Port            int
	Username        string
	Password        string
	VirtualHost     string
	QueueOptions    *QueueOptions
	ConsumerOptions *ConsumerOptions
}

// QueueOptions defines options for RabbitMQ queues
type QueueOptions struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

// ConsumerOptions defines options for RabbitMQ consumers
type ConsumerOptions struct {
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
}
