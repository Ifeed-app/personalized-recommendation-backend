package messaging

import (
	"backend-diplomski/services/email-worker/internal/mail"
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EmailVerificationMessage struct {
	UserID            string `json:"userId"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	VerificationToken string `json:"verificationToken"`
	VerificationURL   string `json:"verificationUrl"`
}

type Consumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queueName  string
	mailSender *mail.Sender
}

func NewConsumer(
	rabbitMQURL string,
	queueName string,
	mailSender *mail.Sender,
) (*Consumer, error) {
	connection, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf(
			"connect to RabbitMQ: %w",
			err,
		)
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()

		return nil, fmt.Errorf(
			"open RabbitMQ channel: %w",
			err,
		)
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()

		return nil, fmt.Errorf(
			"declare queue %q: %w",
			queueName,
			err,
		)
	}

	err = channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()

		return nil, fmt.Errorf(
			"configure consumer QoS: %w",
			err,
		)
	}

	return &Consumer{
		connection: connection,
		channel:    channel,
		queueName:  queueName,
		mailSender: mailSender,
	}, nil
}

func (consumer *Consumer) Start(
	ctx context.Context,
) error {
	deliveries, err := consumer.channel.Consume(
		consumer.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"start queue consumer: %w",
			err,
		)
	}

	log.Printf(
		"email worker is listening on queue %q",
		consumer.queueName,
	)

	for {
		select {
		case <-ctx.Done():
			log.Println(
				"email worker stopping",
			)

			return nil

		case delivery, open := <-deliveries:
			if !open {
				return fmt.Errorf(
					"RabbitMQ delivery channel closed",
				)
			}

			consumer.handleDelivery(delivery)
		}
	}
}

func (consumer *Consumer) handleDelivery(
	delivery amqp.Delivery,
) {
	var message EmailVerificationMessage

	if err := json.Unmarshal(
		delivery.Body,
		&message,
	); err != nil {
		log.Printf(
			"invalid email verification message: %v",
			err,
		)

		if err := delivery.Reject(false); err != nil {
			log.Printf(
				"failed to reject malformed message: %v",
				err,
			)
		}

		return
	}

	if message.Email == "" ||
		message.Username == "" ||
		message.VerificationURL == "" {
		log.Printf(
			"incomplete email verification message for user %q",
			message.UserID,
		)

		if err := delivery.Reject(false); err != nil {
			log.Printf(
				"failed to reject incomplete message: %v",
				err,
			)
		}

		return
	}

	err := consumer.mailSender.SendVerificationEmail(
		message.Email,
		message.Username,
		message.VerificationURL,
	)
	if err != nil {
		log.Printf(
			"failed to send verification email to %s: %v",
			message.Email,
			err,
		)

		if nackErr := delivery.Nack(
			false,
			true,
		); nackErr != nil {
			log.Printf(
				"failed to requeue email message: %v",
				nackErr,
			)
		}

		return
	}

	if err := delivery.Ack(false); err != nil {
		log.Printf(
			"failed to acknowledge email message: %v",
			err,
		)

		return
	}

	log.Printf(
		"verification email sent to %s for user %s",
		message.Email,
		message.UserID,
	)
}

func (consumer *Consumer) Close() error {
	if consumer.channel != nil {
		_ = consumer.channel.Close()
	}

	if consumer.connection != nil {
		return consumer.connection.Close()
	}

	return nil
}
