package broker

type MessageBroker interface {
	Publish(message Message) error
	Consume(queueName string) (<-chan Message, error)
	Ack(deliveryTag uint64) error
	Nack(deliveryTag uint64, requeue bool) error
	Close() error
}
