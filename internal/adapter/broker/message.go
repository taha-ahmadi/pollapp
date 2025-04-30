package broker

// Message encapsulates the contents of the message to be sent
type Message struct {
	ID          string
	Body        []byte
	MessageType string
	QueueName   string
	DeliveryTag uint64
}
