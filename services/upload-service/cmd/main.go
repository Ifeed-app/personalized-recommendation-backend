package main

import (
	uploadv1 "backend-diplomski/proto/gen/upload/v1"
	"backend-diplomski/services/upload-service/db/sqlc"
	"backend-diplomski/services/upload-service/internal/config"
	"backend-diplomski/services/upload-service/internal/db"
	"backend-diplomski/services/upload-service/internal/storage"
	"backend-diplomski/services/upload-service/internal/upload_grpc"
	"backend-diplomski/shared/auth"
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	minioStorage, err := storage.NewMinioStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize bucket")
	}
	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	_ = queries
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := minioStorage.EnsureBucket(ctx); err != nil {
		log.Fatalf("Failed to ensure bucket: %v", err)
	}
	startGRPCServer(cfg.GRPCPort, cfg.ServiceToken, minioStorage, queries)
	log.Printf("upload-service listening on port %s", cfg.AppPort)

}

func startGRPCServer(port string, serviceToken string, minioStorage *storage.MinioStorage, queries *sqlc.Queries) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port %s: %v", port, err)
	}
	serviceAuth := auth.NewServiceAuthInterceptor(serviceToken, "video-worker")
	server := grpc.NewServer(
		grpc.UnaryInterceptor(serviceAuth.Unary),
		grpc.StreamInterceptor(serviceAuth.Stream))

	uploadv1.RegisterUploadServiceServer(
		server,
		upload_grpc.NewUploadServer(minioStorage, queries),
	)
	reflection.Register(server)

	log.Printf("upload-service gRPC listening on port %s", port)

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
