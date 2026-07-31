package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	RabbitMQURL                  string
	RabbitMQShortVideoQueue      string
	UploadServiceURL             string
	PexelsAPIKey                 string
	PexelsAPIURL                 string
	VideoFetchAutoStart          bool
	VideoFetchIntervalSeconds    int
	VideoFetchQueries            []string
	VideoFetchLimit              int
	VideoFetchOrientation        string
	VideoFetchMinDurationSeconds int
	VideoFetchMaxDurationSeconds int
	UploadServiceGRPCURL         string
	ContentServiceGRPCURL        string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("RABBITMQ_SHORT_VIDEO_QUEUE", "short-video-fetch")
	viper.SetDefault("UPLOAD_SERVICE_URL", "http://localhost:8050")
	viper.SetDefault("PEXELS_API_URL", "https://api.pexels.com/v1/videos")
	viper.SetDefault("VIDEO_FETCH_AUTO_START", false)
	viper.SetDefault("VIDEO_FETCH_INTERVAL_SECONDS", 300)
	viper.SetDefault("VIDEO_FETCH_QUERIES", "fitness,food,travel,nature")
	viper.SetDefault("VIDEO_FETCH_LIMIT", 10)
	viper.SetDefault("VIDEO_FETCH_ORIENTATION", "portrait")
	viper.SetDefault("VIDEO_FETCH_MIN_DURATION_SECONDS", 3)
	viper.SetDefault("VIDEO_FETCH_MAX_DURATION_SECONDS", 120)
	viper.SetDefault("UPLOAD_SERVICE_GRPC_URL", "localhost:9050")
	viper.SetDefault("CONTENT_SERVICE_GRPC_URL", "localhost:9080")

	cfg := &Config{
		RabbitMQURL:                  viper.GetString("RABBITMQ_URL"),
		RabbitMQShortVideoQueue:      viper.GetString("RABBITMQ_SHORT_VIDEO_QUEUE"),
		UploadServiceURL:             viper.GetString("UPLOAD_SERVICE_URL"),
		PexelsAPIKey:                 viper.GetString("PEXELS_API_KEY"),
		PexelsAPIURL:                 viper.GetString("PEXELS_API_URL"),
		VideoFetchAutoStart:          viper.GetBool("VIDEO_FETCH_AUTO_START"),
		VideoFetchIntervalSeconds:    viper.GetInt("VIDEO_FETCH_INTERVAL_SECONDS"),
		VideoFetchQueries:            splitAndTrim(viper.GetString("VIDEO_FETCH_QUERIES")),
		VideoFetchLimit:              viper.GetInt("VIDEO_FETCH_LIMIT"),
		VideoFetchOrientation:        viper.GetString("VIDEO_FETCH_ORIENTATION"),
		VideoFetchMinDurationSeconds: viper.GetInt("VIDEO_FETCH_MIN_DURATION_SECONDS"),
		VideoFetchMaxDurationSeconds: viper.GetInt("VIDEO_FETCH_MAX_DURATION_SECONDS"),
		UploadServiceGRPCURL:         viper.GetString("UPLOAD_SERVICE_GRPC_URL"),
		ContentServiceGRPCURL:        viper.GetString("CONTENT_SERVICE_GRPC_URL"),
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

	if cfg.PexelsAPIKey == "" {
		return nil, fmt.Errorf("PEXELS_API_KEY is required")
	}
	if cfg.PexelsAPIURL == "" {
		return nil, fmt.Errorf("PEXELS_API_URL is required")
	}

	return cfg, nil
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
