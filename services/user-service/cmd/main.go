package main

import (
	userv1 "backend-diplomski/proto/gen/user/v1"
	"backend-diplomski/services/user-service/db/sqlc"
	"backend-diplomski/services/user-service/internal/auth"
	"backend-diplomski/services/user-service/internal/config"
	"backend-diplomski/services/user-service/internal/db"
	servicegrpc "backend-diplomski/services/user-service/internal/grpc"
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
	pool, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	_ = queries

	lis, err := net.Listen("tcp", ":"+cfg.AppPort)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", cfg.AppPort, err)
	}
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		"user-service",
		time.Duration(cfg.AccessTokenTTLMinutes)*time.Minute,
	)
	grpcServer := grpc.NewServer()
	userServer := servicegrpc.NewUserServer(queries)
	authServer := servicegrpc.NewAuthServer(queries, jwtManager, cfg)
	userv1.RegisterUserServiceServer(grpcServer, userServer)
	userv1.RegisterAuthServiceServer(grpcServer, authServer)
	log.Printf("user-service connected to database and sqlc initialized")
	log.Printf("user-service gRPC listening on port %s", cfg.AppPort)
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
