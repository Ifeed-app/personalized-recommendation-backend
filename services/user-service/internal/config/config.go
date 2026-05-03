package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort               string
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLHours  int
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("APP_PORT", "8060")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "video_platform")
	viper.SetDefault("ACCESS_TOKEN_TTL_MINUTES", 15)
	viper.SetDefault("REFRESH_TOKEN_TTL_HOURS", 720)
	cfg := &Config{
		AppPort:               viper.GetString("APP_PORT"),
		DBHost:                viper.GetString("DB_HOST"),
		DBPort:                viper.GetString("DB_PORT"),
		DBUser:                viper.GetString("DB_USER"),
		DBPassword:            viper.GetString("DB_PASSWORD"),
		DBName:                viper.GetString("DB_NAME"),
		JWTSecret:             viper.GetString("JWT_SECRET"),
		AccessTokenTTLMinutes: viper.GetInt("ACCESS_TOKEN_TTL_MINUTES"),
		RefreshTokenTTLHours:  viper.GetInt("REFRESH_TOKEN_TTL_HOURS"),
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}
