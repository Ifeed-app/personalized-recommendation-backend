package main

import (
	contentv1 "backend-diplomski/proto/gen/content/v1"
	"backend-diplomski/services/content-service/db"
	"backend-diplomski/services/content-service/db/sqlc"
	"backend-diplomski/services/content-service/internal/config"
	contentgrpc "backend-diplomski/services/content-service/internal/grpc"
	auth "backend-diplomski/shared/auth"
	"context"

	"backend-diplomski/services/content-service/internal/storage"
	_ "fmt"
	"log"
	"net"
	_ "net/http"
	_ "os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	createVideoMethod              = "/content.v1.ContentService/CreateVideo"
	listFeedMethod                 = "/content.v1.ContentService/ListFeedVideos"
	metaData                       = "/content.v1.ContentService/GetVideoMetadata"
	recommendationCandidatesMethod = "/content.v1.ContentService/ListRecommendationCandidates"
)

func main() {
	cfg := config.Load()
	dbPool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	quries := sqlc.New(dbPool)
	minioStorage, err := storage.NewMinioStorage(cfg)
	if err != nil {
		log.Fatalf("failed to create minio storage: %v", err)
	}
	startGRPCServer(cfg.GRPCPort, cfg.VideoWorkerServiceToken, cfg.RecommendationSecretToken, cfg.InteractionSimulatorToken, quries, minioStorage)
}

func startGRPCServer(
	port string,
	videoWorkerServiceToken string,
	recommendationServiceToken string,
	interactionSimulatorToken string,
	queries *sqlc.Queries,
	minioStorage *storage.MinioStorage,
) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port %s: %v", port, err)
	}

	videoWorkerAuth := auth.NewServiceAuthInterceptor(
		videoWorkerServiceToken,
		"video-worker",
	)

	recommendationAuth := auth.NewServiceAuthInterceptor(
		recommendationServiceToken,
		"recommendation-service",
	)

	simulatorAuth := auth.NewServiceAuthInterceptor(
		interactionSimulatorToken,
		"interaction-simulator")

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			contentServiceUnaryInterceptor(
				videoWorkerAuth,
				recommendationAuth,
				simulatorAuth,
			),
		),
	)

	contentv1.RegisterContentServiceServer(
		grpcServer,
		contentgrpc.NewServer(queries, minioStorage),
	)

	reflection.Register(grpcServer)

	log.Printf("content-service gRPC listening on port %s", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve content-service gRPC: %v", err)
	}
}

func contentServiceUnaryInterceptor(
	videoWorkerAuth *auth.ServiceAuthInterceptor,
	recommendationAuth *auth.ServiceAuthInterceptor,
	simulatorAuth *auth.ServiceAuthInterceptor,
) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		switch info.FullMethod {

		case listFeedMethod:
			return handler(ctx, req)

		case createVideoMethod:
			return videoWorkerAuth.Unary(
				ctx,
				req,
				info,
				handler,
			)

		case metaData:
			return recommendationAuth.Unary(
				ctx,
				req,
				info,
				handler,
			)

		case recommendationCandidatesMethod:
			md, _ := metadata.FromIncomingContext(ctx)
			serviceNames := md.Get("x-service-name")

			if len(serviceNames) > 0 &&
				serviceNames[0] == "interaction-simulator" {

				return simulatorAuth.Unary(
					ctx,
					req,
					info,
					handler,
				)
			}

			return recommendationAuth.Unary(
				ctx,
				req,
				info,
				handler,
			)

		default:
			return nil, status.Error(
				codes.PermissionDenied,
				"method access policy is not configured",
			)
		}
	}
}
