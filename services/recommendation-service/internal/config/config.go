package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	GRPCPort string

	InteractionServiceGRPCURL string
	ContentServiceGRPCURL     string

	RecommendationServiceToken string

	MLInferenceURL string
}

func Load() Config {
	viper.AutomaticEnv()

	viper.SetDefault(
		"RECOMMENDATION_SERVICE_GRPC_PORT",
		"9090",
	)

	viper.SetDefault(
		"INTERACTION_SERVICE_GRPC_URL",
		"interaction-service:9070",
	)

	viper.SetDefault(
		"CONTENT_SERVICE_GRPC_URL",
		"content-service:9080",
	)

	viper.SetDefault(
		"ML_INFERENCE_URL",
		"http://ml-inference:8090",
	)

	return Config{
		GRPCPort: viper.GetString(
			"RECOMMENDATION_SERVICE_GRPC_PORT",
		),

		InteractionServiceGRPCURL: viper.GetString(
			"INTERACTION_SERVICE_GRPC_URL",
		),

		ContentServiceGRPCURL: viper.GetString(
			"CONTENT_SERVICE_GRPC_URL",
		),

		RecommendationServiceToken: viper.GetString(
			"RECOMMENDATION_SERVICE_TOKEN",
		),

		MLInferenceURL: viper.GetString(
			"ML_INFERENCE_URL",
		),
	}
}
