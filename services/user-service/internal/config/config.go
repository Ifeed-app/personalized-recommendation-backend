package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort                          string
	HTTPPort                         string
	DBHost                           string
	DBPort                           string
	DBUser                           string
	DBPassword                       string
	DBName                           string
	JWTSecret                        string
	AccessTokenTTLMinutes            int
	RefreshTokenTTLHours             int
	RabbitMQURL                      string
	EmailVerificationQueue           string
	EmailVerificationBaseURL         string
	EmailVerificationTokenTTLMinutes time.Duration
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("APP_PORT", "8060")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "video_platform")
	viper.SetDefault("ACCESS_TOKEN_TTL_MINUTES", 15)
	viper.SetDefault("REFRESH_TOKEN_TTL_HOURS", 720)
	viper.SetDefault("EMAIL_VERIFICATION_QUEUE", "email-verification")
	viper.SetDefault("EMAIL_VERIFICATION_BASE_URL", "http://localhost:8061/verify-email")
	viper.SetDefault("EMAIL_VERIFICATION_TOKEN_TTL_MINUTES", 30)
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("HTTP_PORT", "8061")

	cfg := &Config{
		AppPort:                          viper.GetString("APP_PORT"),
		DBHost:                           viper.GetString("DB_HOST"),
		DBPort:                           viper.GetString("DB_PORT"),
		DBUser:                           viper.GetString("DB_USER"),
		DBPassword:                       viper.GetString("DB_PASSWORD"),
		DBName:                           viper.GetString("DB_NAME"),
		JWTSecret:                        viper.GetString("JWT_SECRET"),
		AccessTokenTTLMinutes:            viper.GetInt("ACCESS_TOKEN_TTL_MINUTES"),
		RefreshTokenTTLHours:             viper.GetInt("REFRESH_TOKEN_TTL_HOURS"),
		RabbitMQURL:                      viper.GetString("RABBITMQ_URL"),
		EmailVerificationQueue:           viper.GetString("EMAIL_VERIFICATION_QUEUE"),
		EmailVerificationBaseURL:         viper.GetString("EMAIL_VERIFICATION_BASE_URL"),
		EmailVerificationTokenTTLMinutes: viper.GetDuration("EMAIL_VERIFICATION_TOKEN_TTL_MINUTES"),
		HTTPPort:                         viper.GetString("HTTP_PORT"),
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}
