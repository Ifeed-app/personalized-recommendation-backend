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

type Conusmer struct {
	rabbitURL string
	queueName string
}

func NewConusmer(rabbitURL string, queueName string) *Conusmer {
	return &Conusmer{
		rabbitURL: rabbitURL,
		queueName: queueName,
	}
}

func (c *Conusmer) Start(ctx context.Context) error {
	conn, err := c.connectWithRetry()
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open rabbit channel: %w", err)
	}
	defer ch.Close()
	queue, err := ch.QueueDeclare(
		c.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to delcare queue: %w", err)
	}
	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to set qos: %w", err)
	}
	deliveries, err := ch.Consume(
		queue.Name,
		"video-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming message: %w", err)
	}
	log.Printf("worker is listening to queue: %s", queue.Name)
	for {
		select {
		case <-ctx.Done():
			log.Println("video-worker consumer stopped")
			return nil

		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbit deliveries channel closed")
			}
			err := c.handleMessage(delivery)
			if err != nil {
				log.Printf("failed to handle message:%v", err)
				_ = delivery.Nack(false, false)
				continue
			}
			_ = delivery.Ack(false)
		}
	}

}

func (c *Conusmer) connectWithRetry() (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for attempt := 1; attempt <= 30; attempt++ {
		conn, err = amqp.Dial(c.rabbitURL)
		if err == nil {
			log.Println("Connected to rabbit")
			return conn, nil
		}
		log.Printf("Rabbit not ready, rettrying ... %d/30: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to rabbit: %w", err)
}

func (c *Conusmer) handleMessage(delivery amqp.Delivery) error {
	var msg message.ShortVideoFetchMessage
	log.Printf("raw msg recieved:%s", string(delivery.Body))
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return fmt.Errorf("invalid msg json:%w", err)
	}
	if msg.VideoURL == "" {
		return fmt.Errorf("videoURL is required")
	}
	log.Println("short video fetch message received")
	log.Printf("jobId: %s", msg.JobID)
	log.Printf("source: %s", msg.Source)
	log.Printf("videoUrl: %s", msg.VideoURL)
	log.Printf("title: %s", msg.Title)
	log.Printf("description: %s", msg.Description)

	return nil
}
