package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	RabbitMQURL             string
	RabbitMQShortVideoQueue string
	UploadServiceURL        string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("RABBITMQ_SHORT_VIDEO_QUEUE", "short-video-fetch")
	viper.SetDefault("UPLOAD_SERVICE_URL", "http://localhost:8050")

	cfg := &Config{
		RabbitMQURL:             viper.GetString("RABBITMQ_URL"),
		RabbitMQShortVideoQueue: viper.GetString("RABBITMQ_SHORT_VIDEO_QUEUE"),
		UploadServiceURL:        viper.GetString("UPLOAD_SERVICE_URL"),
	}

	if cfg.RabbitMQURL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL is required")
	}

	if cfg.RabbitMQShortVideoQueue == "" {
		return nil, fmt.Errorf("RABBITMQ_SHORT_VIDEO_QUEUE is required")
	}

	if cfg.UploadServiceURL == "" {
		return nil, fmt.Errorf("UPLOAD_SERVICE_URL is required")
	}

	return cfg, nil
}
