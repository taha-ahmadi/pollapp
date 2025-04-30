package testbroker

import (
	"log"

	"pollapp/internal/adapter/broker"
)

type Adapter struct {
}

func New() (*Adapter, error) {

	return &Adapter{}, nil
}

func (a *Adapter) Publish(message broker.Message) error {

	log.Println("Sending message to Rabbitmq ...")
	return nil
}

func (a *Adapter) Consume(queueName string) (<-chan broker.Message, error) {

	messageChan := make(chan broker.Message)

	return messageChan, nil
}

func (a *Adapter) Ack(deliveryTag uint64) error {
	return nil
}
