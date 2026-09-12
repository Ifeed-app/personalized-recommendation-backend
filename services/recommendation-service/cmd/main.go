package main

import (
	recommendationv1 "backend-diplomski/proto/gen/recommendation/v1"
	"backend-diplomski/services/recommendation-service/internal/client/content"
	"backend-diplomski/services/recommendation-service/internal/client/interaction"
	"backend-diplomski/services/recommendation-service/internal/config"
	recommendationgrpc "backend-diplomski/services/recommendation-service/internal/grpc"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	interactionClient, err := interaction.NewClient(
		cfg.InteractionServiceGRPCURL,
		cfg.RecommendationServiceToken,
	)
	if err != nil {
		log.Fatalf(
			"failed to initialize interaction client: %v",
			err,
		)
	}
	defer interactionClient.Close()

	contentClient, err := content.NewClient(
		cfg.ContentServiceGRPCURL,
		cfg.RecommendationServiceToken,
	)
	if err != nil {
		log.Fatalf(
			"failed to initialize content client: %v",
			err,
		)
	}
	defer contentClient.Close()

	listener, err := net.Listen(
		"tcp",
		":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf(
			"failed to listen on gRPC port %s: %v",
			cfg.GRPCPort,
			err)
	}
	grpcServer := grpc.NewServer()
	recommendationServer := recommendationgrpc.NewServer(
		interactionClient,
		contentClient,
	)

	recommendationv1.RegisterRecommendationServiceServer(
		grpcServer,
		recommendationServer,
	)
	reflection.Register(grpcServer)

	log.Printf(
		"recommendation-service gRPC listening on port %s",
		cfg.GRPCPort,
	)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf(
			"failed to serve recommendation-service: %v",
			err,
		)
	}
}
