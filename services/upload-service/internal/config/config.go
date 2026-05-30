package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3Region    string
	S3UseSSL    bool

	GRPCPort string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("APP_PORT", "8050")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "video_platform")

	viper.SetDefault("S3_ENDPOINT", "localhost:9000")
	viper.SetDefault("S3_ACCESS_KEY", "minioadmin")
	viper.SetDefault("S3_SECRET_KEY", "minioadmin123")
	viper.SetDefault("S3_BUCKET", "feeder-media")
	viper.SetDefault("S3_REGION", "us-east-1")
	viper.SetDefault("S3_USE_SSL", false)
	viper.SetDefault("GRPC_PORT", "9050")

	cfg := &Config{
		AppPort: viper.GetString("APP_PORT"),

		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),

		S3Endpoint:  viper.GetString("S3_ENDPOINT"),
		S3AccessKey: viper.GetString("S3_ACCESS_KEY"),
		S3SecretKey: viper.GetString("S3_SECRET_KEY"),
		S3Bucket:    viper.GetString("S3_BUCKET"),
		S3Region:    viper.GetString("S3_REGION"),
		S3UseSSL:    viper.GetBool("S3_USE_SSL"),
		GRPCPort:    viper.GetString("GRPC_PORT"),
	}

	if cfg.S3Endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT is required")
	}

	if cfg.S3AccessKey == "" {
		return nil, fmt.Errorf("S3_ACCESS_KEY is required")
	}

	if cfg.S3SecretKey == "" {
		return nil, fmt.Errorf("S3_SECRET_KEY is required")
	}

	if cfg.S3Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	return cfg, nil
}
