package broker

type MessageBroker interface {
	Publish(message Message) error
	Consume(queueName string) (<-chan Message, error)
	ConsumeWithManualAckAndSinglePrefetch(queueName string) (<-chan Message, error)
	Ack(deliveryTag uint64) error
	Nack(tag uint64, requeue bool) error
}
