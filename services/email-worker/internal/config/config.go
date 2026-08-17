package config

import (
	"fmt"
	_ "strings"
	_ "time"

	"github.com/spf13/viper"
)

type Config struct {
	RabbitMQURL            string
	EmailVerificationQueue string

	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromName  string
	SMTPFromEmail string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault(
		"RABBITMQ_URL",
		"amqp://guest:guest@localhost:5672/",
	)

	viper.SetDefault(
		"EMAIL_VERIFICATION_QUEUE",
		"email-verification",
	)

	viper.SetDefault(
		"SMTP_HOST",
		"localhost",
	)

	viper.SetDefault(
		"SMTP_PORT",
		1025,
	)

	viper.SetDefault(
		"SMTP_FROM_NAME",
		"Feeder",
	)

	viper.SetDefault(
		"SMTP_FROM_EMAIL",
		"no-reply@feeder.local",
	)

	cfg := &Config{
		RabbitMQURL: viper.GetString(
			"RABBITMQ_URL",
		),
		EmailVerificationQueue: viper.GetString(
			"EMAIL_VERIFICATION_QUEUE",
		),
		SMTPHost: viper.GetString(
			"SMTP_HOST",
		),
		SMTPPort: viper.GetInt(
			"SMTP_PORT",
		),
		SMTPUsername: viper.GetString(
			"SMTP_USERNAME",
		),
		SMTPPassword: viper.GetString(
			"SMTP_PASSWORD",
		),
		SMTPFromName: viper.GetString(
			"SMTP_FROM_NAME",
		),
		SMTPFromEmail: viper.GetString(
			"SMTP_FROM_EMAIL",
		),
	}

	if cfg.RabbitMQURL == "" {
		return nil, fmt.Errorf(
			"RABBITMQ_URL is required",
		)
	}

	if cfg.EmailVerificationQueue == "" {
		return nil, fmt.Errorf(
			"EMAIL_VERIFICATION_QUEUE is required",
		)
	}

	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf(
			"SMTP_HOST is required",
		)
	}

	if cfg.SMTPPort <= 0 {
		return nil, fmt.Errorf(
			"SMTP_PORT must be greater than zero",
		)
	}

	if cfg.SMTPFromEmail == "" {
		return nil, fmt.Errorf(
			"SMTP_FROM_EMAIL is required",
		)
	}

	return cfg, nil
}
