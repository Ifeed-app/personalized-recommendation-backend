package main

import (
	interactionv1 "backend-diplomski/proto/gen/interaction/v1"
	"backend-diplomski/services/interaction-service/db"
	"backend-diplomski/services/interaction-service/db/sqlc"
	"backend-diplomski/services/interaction-service/internal/config"
	interactiongrpc "backend-diplomski/services/interaction-service/internal/grpc"
	"backend-diplomski/shared/auth"
	"context"
	_ "fmt"
	"log"
	"net"
	_ "net/http"
	_ "os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	createInteractionMethod = "/interaction.v1.InteractionService/CreateInteraction"

	recordWatchSessionMethod = "/interaction.v1.InteractionService/RecordWatchSession"

	getVideoInteractionSummaryMethod = "/interaction.v1.InteractionService/GetVideoInteractionSummary"

	getUserInteractionHistoryMethod = "/interaction.v1.InteractionService/GetUserInteractionHistory"
)

func main() {
	cfg := config.Load()

	dbPool, err := db.NewPool(&cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	queries := sqlc.New(dbPool)

	startGRPCServer(cfg.GRPCPort, cfg.JWTSecret, cfg.RecommendationSecretToken, queries)
}

func startGRPCServer(
	port string,
	jwtSecret string,
	recommendationServiceToken string,
	queries *sqlc.Queries,
) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf(
			"failed to listen on gRPC port %s: %v",
			port,
			err,
		)
	}

	jwtManager := auth.NewJWTManager(
		jwtSecret,
		"user-service",
		time.Minute)

	authInterceptor := auth.NewUnaryAuthInterceptor(
		jwtManager,
	)

	serviceAuth := auth.NewServiceAuthInterceptor(
		recommendationServiceToken,
		"recommendation-service")
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			interactionServiceUnaryInterceptor(
				authInterceptor,
				serviceAuth)),
	)

	interactionServer := interactiongrpc.NewServer(queries)

	interactionv1.RegisterInteractionServiceServer(
		grpcServer,
		interactionServer,
	)

	reflection.Register(grpcServer)

	log.Printf(
		"interaction-service gRPC listening on port %s",
		port,
	)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf(
			"failed to serve interaction-service gRPC: %v",
			err,
		)
	}
}

func interactionServiceUnaryInterceptor(
	userAuth *auth.UnaryAuthInterceptor,
	serviceAuth *auth.ServiceAuthInterceptor,
) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		switch info.FullMethod {

		case createInteractionMethod,
			recordWatchSessionMethod,
			getVideoInteractionSummaryMethod:

			return userAuth.Intercept(
				ctx,
				req,
				info,
				handler,
			)

		case getUserInteractionHistoryMethod:

			return handler(ctx, req)

		default:
			return nil, status.Error(
				codes.PermissionDenied,
				"method access policy is not configured",
			)
		}
	}
}
