package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort          string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	GRPCPort         string
	S3Endpoint       string
	S3AccessKey      string
	S3Bucket         string
	S3SecretKey      string
	S3Region         string
	S3UseSSL         bool
	S3PublicEndpoint string
}

func Load() *Config {
	viper.AutomaticEnv()
	viper.SetDefault("APP_PORT", "8082")
	viper.SetDefault("GRPC_PORT", "9080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "video_platform")
	viper.SetDefault("S3_ENDPOINT", "localhost:9000")
	viper.SetDefault("S3_ACCESS_KEY", "minioadmin")
	viper.SetDefault("S3_SECRET_KEY", "minioadmin")
	viper.SetDefault("S3_BUCKET", "feederbucket")
	viper.SetDefault("S3_REGION", "us-east-1")
	viper.SetDefault("S3_USE_SSL", false)
	viper.SetDefault("S3_PUBLIC_ENDPOINT", "localhost:9000")

	cfg := &Config{
		AppPort:          viper.GetString("APP_PORT"),
		GRPCPort:         viper.GetString("GRPC_PORT"),
		DBHost:           viper.GetString("DB_HOST"),
		DBPort:           viper.GetString("DB_PORT"),
		DBUser:           viper.GetString("DB_USER"),
		DBPassword:       viper.GetString("DB_PASSWORD"),
		DBName:           viper.GetString("DB_NAME"),
		S3Endpoint:       viper.GetString("S3_ENDPOINT"),
		S3AccessKey:      viper.GetString("S3_ACCESS_KEY"),
		S3SecretKey:      viper.GetString("S3_SECRET_KEY"),
		S3Bucket:         viper.GetString("S3_BUCKET"),
		S3Region:         viper.GetString("S3_REGION"),
		S3UseSSL:         viper.GetBool("S3_USE_SSL"),
		S3PublicEndpoint: viper.GetString("S3_PUBLIC_ENDPOINT"),
	}

	log.Printf(
		"content-service config loaded appPort=%s grpcPort=%s dbHost=%s dbPort=%s dbName=%s",
		cfg.AppPort,
		cfg.GRPCPort,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	return cfg
}
