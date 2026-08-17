package main

import (
	"context"

	"backend-diplomski/services/email-worker/internal/config"
	"backend-diplomski/services/email-worker/internal/mail"
	"backend-diplomski/services/email-worker/internal/messaging"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf(
			"failed to load configuration: %v",
			err,
		)
	}

	mailSender := mail.NewSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPFromName,
		cfg.SMTPFromEmail,
	)

	consumer, err := messaging.NewConsumer(
		cfg.RabbitMQURL,
		cfg.EmailVerificationQueue,
		mailSender,
	)
	if err != nil {
		log.Fatalf(
			"failed to initialize email consumer: %v",
			err,
		)
	}
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf(
			"email consumer stopped with error: %v",
			err,
		)
	}
}
