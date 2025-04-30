package rabbitmq

import (
	"fmt"
	"log"
	"log/slog"
	"pollapp/internal/adapter/broker"
	"pollapp/pkg/zlog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const RabbitMQNS = "rabbitmq"

// TODO: add ack and publish metrics
type Config struct {
	Host            string           `koanf:"host"`
	Port            int              `koanf:"port"`
	Username        string           `koanf:"username"`
	Password        string           `koanf:"password"`
	VirtualHost     string           `koanf:"virtual_host"`
	QueueOptions    *QueueOptions    `koanf:"queue_options"`
	ConsumerOptions *ConsumerOptions `koanf:"consumer_options"`
}

type QueueOptions struct {
	Durable    bool `koanf:"durable"`
	AutoDelete bool `koanf:"auto_delete"`
	Exclusive  bool `koanf:"exclusive"`
	NoWait     bool `koanf:"no_wait"`
}

var DefaultQueueOptions = QueueOptions{
	Durable:    true,
	AutoDelete: false,
	Exclusive:  false,
	NoWait:     false,
}

type ConsumerOptions struct {
	AutoAck   bool `koanf:"auto_ack"`
	Exclusive bool `koanf:"exclusive"`
	NoLocal   bool `koanf:"no_local"`
	NoWait    bool `koanf:"no_wait"`
}

var DefaultConsumerOptions = ConsumerOptions{
	AutoAck:   true,
	Exclusive: false,
	NoLocal:   false,
	NoWait:    false,
}

type Adapter struct {
	cfg     Config
	conn    *amqp.Connection
	channel *amqp.Channel
	runMod  string
}

func New(config Config, runMod string, queueNames []string) (*Adapter, error) {
	fmt.Println("New RabbitMQ", config)
	if config.QueueOptions == nil {
		config.QueueOptions = &DefaultQueueOptions
		zlog.L.Info("queue options is nil, using default", zlog.Any("options", config.QueueOptions))
	}
	if config.ConsumerOptions == nil {
		config.ConsumerOptions = &DefaultConsumerOptions
		zlog.L.Info("consumer options is nil, using default", zlog.Any("options", config.ConsumerOptions))
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/%s", config.Username, config.Password, config.Host, config.Port, config.VirtualHost))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}
	for _, queueName := range queueNames {
		fmt.Printf("queue declare %s \n", queueName)
		_, err := channel.QueueDeclare(
			queueName,
			config.QueueOptions.Durable,    // durable
			config.QueueOptions.AutoDelete, // delete when unused
			config.QueueOptions.Exclusive,  // exclusive
			config.QueueOptions.NoWait,     // no-wait
			nil,                            // arguments
		)
		if err != nil {
			return nil, fmt.Errorf("queue declare %s %v", queueName, slog.String("error", err.Error()))
		}
	}
	adapter := &Adapter{
		conn:    conn,
		channel: channel,
		cfg:     config,
		runMod:  runMod,
	}

	return adapter, nil
}

func (a *Adapter) Publish(message broker.Message) error {

	// base64EncodedMessage := base64.StdEncoding.EncodeToString(message.Body)
	// TODO: add publishOptions

	err := a.channel.Publish(
		"",                // exchange
		message.QueueName, // routing key (same as queue name)
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			MessageId:    message.ID,
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/protobuf",
			Type:         message.MessageType,
			Body:         message.Body,
		})
	// If there is an error publishing the message, a log will be displayed in the terminal.
	if err != nil {
		log.Println("publish rabbit - publish", slog.String("error", err.Error()))
		return err
	}

	log.Println("Sending message to Rabbitmq ...")
	return nil
}

func (a *Adapter) Consume(queueName string) (<-chan broker.Message, error) {
	messages, err := a.channel.Consume(
		queueName,                       // queue
		"",                              // consumer
		a.cfg.ConsumerOptions.AutoAck,   // auto-ack
		a.cfg.ConsumerOptions.Exclusive, // exclusive
		a.cfg.ConsumerOptions.NoLocal,   // no-local
		a.cfg.ConsumerOptions.NoWait,    // no-wait
		nil,                             // arguments
	)
	if err != nil {
		zlog.L.Error("can't consume", zlog.String("queueName", queueName), zlog.Any("error", err))
		return nil, err
	}

	messageChan := make(chan broker.Message)
	//TODO: Add communication with  supervisor
	go func() {
		defer func() {
			if rErr := recover(); rErr != nil {
				zlog.L.Debug("rabbit consumer panic", zlog.String("queueName", queueName), zlog.Any("panic", rErr))
			}
		}()
		for delivery := range messages {
			zlog.L.Debug("Consume message", zlog.String("queueName", queueName))
			message := broker.Message{
				ID:          delivery.MessageId,
				Body:        delivery.Body,
				MessageType: delivery.Type,
				QueueName:   delivery.RoutingKey,
				DeliveryTag: delivery.DeliveryTag,
			}

			messageChan <- message

			zlog.L.Debug("Consumer Options", zlog.String("queueName", queueName), zlog.Any("ConsumerOptions", a.cfg.ConsumerOptions))
			if !a.cfg.ConsumerOptions.AutoAck {
				// Acknowledge the message to remove it from the queue
				err = a.Ack(delivery.DeliveryTag)
				if err != nil {
					zlog.L.Error("Error acknowledging message", zlog.String("queueName", queueName), zlog.Any("error", err))
				}
				fmt.Println("Consume message RabbitMQ 2", time.Now())
				zlog.L.Debug("acknowledged message  successfully", zlog.String("queueName", queueName))
			}
		}
		zlog.L.Info("rabbit consumer  channel closed", zlog.String("queueName", queueName))
		close(messageChan)
	}()

	return messageChan, nil
}

func (a *Adapter) Ack(deliveryTag uint64) error {
	return a.channel.Ack(deliveryTag, false)
}

func (a *Adapter) Nack(tag uint64, requeue bool) error {
	return a.channel.Nack(tag, false, requeue)
}

func (a *Adapter) ConsumeDelivery(queueName string) (<-chan amqp.Delivery, error) {
	messages, err := a.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		zlog.L.Error("can't consume", zlog.String("queueName", queueName), zlog.Any("error", err))
		return nil, err
	}

	return messages, nil
}

// ConsumeWithManualAckAndSinglePrefetch consumes messages with manual acknowledgment and prefetch count of 1
func (a *Adapter) ConsumeWithManualAckAndSinglePrefetch(queueName string) (<-chan broker.Message, error) {
	// Set prefetch count to 1
	err := a.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		zlog.L.Error("failed to set QoS", zlog.String("queueName", queueName), zlog.Any("error", err))
		return nil, err
	}

	// Use manual acknowledgment
	messages, err := a.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (false for manual ack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		zlog.L.Error("can't consume with single prefetch", zlog.String("queueName", queueName), zlog.Any("error", err))
		return nil, err
	}

	messageChan := make(chan broker.Message)
	go func() {
		defer func() {
			if rErr := recover(); rErr != nil {
				zlog.L.Debug("rabbit consumer panic", zlog.String("queueName", queueName), zlog.Any("panic", rErr))
			}
		}()
		for delivery := range messages {
			zlog.L.Debug("Consume message with single prefetch", zlog.String("queueName", queueName))
			message := broker.Message{
				ID:          delivery.MessageId,
				Body:        delivery.Body,
				MessageType: delivery.Type,
				QueueName:   delivery.RoutingKey,
				DeliveryTag: delivery.DeliveryTag,
			}

			messageChan <- message
		}
		zlog.L.Info("rabbit consumer channel closed", zlog.String("queueName", queueName))
		close(messageChan)
	}()

	return messageChan, nil
}
