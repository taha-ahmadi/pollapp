package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pollapp/internal/adapter/broker"
	"pollapp/pkg/zlog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	reconnectDelay      = 5 * time.Second
	resendDelay         = 5 * time.Second
	maxReconnectRetries = 20
)

type Broker struct {
	config        Config
	connection    *amqp.Connection
	channel       *amqp.Channel
	queues        map[string]amqp.Queue
	clientName    string
	mu            sync.Mutex
	isClosed      bool
	notifyClose   chan *amqp.Error
	notifyReturn  chan amqp.Return
	notifyConfirm chan amqp.Confirmation
	consumers     map[string]<-chan amqp.Delivery
}

func New(config Config, clientName string, queues []string) (*Broker, error) {
	broker := &Broker{
		config:     config,
		clientName: clientName,
		queues:     make(map[string]amqp.Queue),
		consumers:  make(map[string]<-chan amqp.Delivery),
		isClosed:   false,
	}

	if err := broker.connect(); err != nil {
		return nil, err
	}

	for _, queueName := range queues {
		if err := broker.declareQueue(queueName); err != nil {
			broker.Close()
			return nil, err
		}
	}

	return broker, nil
}

func (b *Broker) connect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return fmt.Errorf("broker is closed")
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		b.config.Username,
		b.config.Password,
		b.config.Host,
		b.config.Port,
		b.config.VirtualHost)

	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	b.connection = conn

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	b.channel = channel

	b.notifyClose = make(chan *amqp.Error)
	b.notifyReturn = make(chan amqp.Return)
	b.notifyConfirm = make(chan amqp.Confirmation)

	b.channel.NotifyClose(b.notifyClose)
	b.channel.NotifyReturn(b.notifyReturn)

	if err := b.channel.Confirm(false); err != nil {
		b.Close()
		return fmt.Errorf("failed to enable publish confirmations: %w", err)
	}
	b.channel.NotifyPublish(b.notifyConfirm)

	return nil
}

func (b *Broker) reconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return fmt.Errorf("broker is closed")
	}

	if b.channel != nil {
		b.channel.Close()
	}
	if b.connection != nil {
		b.connection.Close()
	}

	return b.connect()
}

func (b *Broker) declareQueue(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return fmt.Errorf("broker is closed")
	}

	// Declare the queue
	queue, err := b.channel.QueueDeclare(
		name,                             // name
		b.config.QueueOptions.Durable,    // durable
		b.config.QueueOptions.AutoDelete, // auto-delete
		b.config.QueueOptions.Exclusive,  // exclusive
		b.config.QueueOptions.NoWait,     // no-wait
		nil,                              // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", name, err)
	}

	b.queues[name] = queue
	return nil
}

func (b *Broker) Publish(message broker.Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return fmt.Errorf("broker is closed")
	}

	if _, ok := b.queues[message.QueueName]; !ok {
		if err := b.declareQueue(message.QueueName); err != nil {
			return err
		}
	}

	pub := amqp.Publishing{
		MessageId:     message.ID,
		ContentType:   "application/json",
		Body:          message.Body,
		DeliveryMode:  amqp.Persistent,
		Timestamp:     time.Now(),
		Type:          message.MessageType,
		CorrelationId: message.ID,
	}

	err := b.channel.PublishWithContext(
		context.Background(),
		"",                // exchange
		message.QueueName, // routing key (queue name)
		true,              // mandatory
		false,             // immediate
		pub,               // message
	)
	if err != nil {
		if err = b.reconnect(); err != nil {
			return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
		}

		err = b.channel.PublishWithContext(
			context.Background(),
			"",                // exchange
			message.QueueName, // routing key (queue name)
			true,              // mandatory
			false,             // immediate
			pub,               // message
		)
		if err != nil {
			return fmt.Errorf("failed to publish message after reconnect: %w", err)
		}
	}

	// Wait for confirmation
	select {
	case confirm := <-b.notifyConfirm:
		if !confirm.Ack {
			return fmt.Errorf("failed to publish message: negative acknowledgment")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("failed to publish message: confirmation timeout")
	}

	return nil
}

func (b *Broker) Consume(queueName string) (<-chan broker.Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return nil, fmt.Errorf("broker is closed")
	}

	if _, ok := b.queues[queueName]; !ok {
		if err := b.declareQueue(queueName); err != nil {
			return nil, err
		}
	}

	consumerTag := fmt.Sprintf("%s-%s", b.clientName, queueName)
	deliveries, err := b.channel.Consume(
		queueName,                          // queue
		consumerTag,                        // consumer
		b.config.ConsumerOptions.AutoAck,   // auto-ack
		b.config.ConsumerOptions.Exclusive, // exclusive
		b.config.ConsumerOptions.NoLocal,   // no-local
		b.config.ConsumerOptions.NoWait,    // no-wait
		nil,                                // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume from queue %s: %w", queueName, err)
	}

	b.consumers[queueName] = deliveries

	messageChan := make(chan broker.Message)

	go func() {
		defer close(messageChan)

		for {
			select {
			case delivery, ok := <-deliveries:
				if !ok {
					zlog.L.Info("Delivery channel closed")
					return
				}

				messageChan <- broker.Message{
					ID:          delivery.MessageId,
					Body:        delivery.Body,
					MessageType: delivery.Type,
					QueueName:   queueName,
					DeliveryTag: delivery.DeliveryTag,
				}

			case <-b.notifyClose:
				zlog.L.Warn("AMQP connection closed during consume")
				return
			}
		}
	}()

	return messageChan, nil
}

func (b *Broker) Ack(deliveryTag uint64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return fmt.Errorf("broker is closed")
	}

	return b.channel.Ack(deliveryTag, false)
}

func (b *Broker) Nack(deliveryTag uint64, requeue bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return fmt.Errorf("broker is closed")
	}

	return b.channel.Nack(deliveryTag, false, requeue)
}

func (b *Broker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return nil
	}

	b.isClosed = true

	var errs []error

	if b.channel != nil {
		if err := b.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing channel: %w", err))
		}
	}

	if b.connection != nil {
		if err := b.connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing RabbitMQ connection: %v", errs)
	}

	return nil
}
