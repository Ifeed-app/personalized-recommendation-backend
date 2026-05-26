package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"video-worker/internal/message"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	rabbitURL string
	queueName string
}

func NewPublisher(rabbitURL string, queueName string) *Publisher {
	return &Publisher{
		rabbitURL: rabbitURL,
		queueName: queueName,
	}
}

func (p *Publisher) PublishFetchJob(ctx context.Context, msg message.ShortVideoFetchMessage) error {
	conn, err := p.connectWithRetry()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		p.queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal fetch job: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish fetch job: %w", err)
	}

	log.Printf("published fetch job: jobId=%s provider=%s query=%s page=%d", msg.JobID, msg.Provider, msg.Query, msg.Page)

	return nil
}

func (p *Publisher) connectWithRetry() (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for attempt := 1; attempt <= 30; attempt++ {
		conn, err = amqp.Dial(p.rabbitURL)
		if err == nil {
			return conn, nil
		}

		log.Printf("RabbitMQ publisher connection failed, retrying %d/30: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect publisher to RabbitMQ after retries: %w", err)
}
