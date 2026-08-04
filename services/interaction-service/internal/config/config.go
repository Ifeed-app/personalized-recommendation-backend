package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	GRPCPort   string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
}

func Load() Config {
	viper.AutomaticEnv()

	viper.SetDefault("INTERACTION_SERVICE_GRPC_PORT", "9070")

	viper.SetDefault("DB_HOST", "postgres")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "video_platform")
	viper.SetDefault("DB_SSLMODE", "disable")

	cfg := Config{
		GRPCPort:   viper.GetString("INTERACTION_SERVICE_GRPC_PORT"),
		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		DBSSLMode:  viper.GetString("DB_SSLMODE"),
		JWTSecret:  viper.GetString("JWT_SECRET"),
	}

	if cfg.GRPCPort == "" {
		log.Fatal("INTERACTION_SERVICE_GRPC_PORT is required")
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	return cfg
}
