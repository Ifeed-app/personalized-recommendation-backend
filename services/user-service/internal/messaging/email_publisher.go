package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EmailVerificationMessage struct {
	UserID            string `json:"userId"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	VerificationToken string `json:"verificationToken"`
	VerificationURL   string `json:"verificationUrl"`
}

type EmailPublisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queueName  string
}

func NewEmailPublisher(rabbitMQURL string, queueName string) (*EmailPublisher, error) {
	connection, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("connect to RabbitMQ:%w", err)
	}
	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open RabbitMQ channel: %w", err)
	}
	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()

		return nil, fmt.Errorf(
			"declare email verification queueu: %w", err)
	}
	return &EmailPublisher{
		connection: connection,
		channel:    channel,
		queueName:  queueName,
	}, nil
}

func (publisher *EmailPublisher) PublishVerificationEmail(
	ctx context.Context, message EmailVerificationMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf(
			"marshal email verification message: %w", err)
	}
	err = publisher.channel.PublishWithContext(
		ctx,
		"",
		publisher.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish email verification message: %w",
			err)
	}
	return nil
}

func (publisher *EmailPublisher) Close() error {
	if publisher.channel != nil {
		_ = publisher.channel.Close()
	}
	if publisher.connection != nil {
		return publisher.connection.Close()
	}
	return nil

}
