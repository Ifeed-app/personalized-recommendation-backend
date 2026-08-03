package main

import (
	interactionv1 "backend-diplomski/proto/gen/interaction/v1"
	"backend-diplomski/services/interaction-service/db"
	"backend-diplomski/services/interaction-service/db/sqlc"
	"backend-diplomski/services/interaction-service/internal/config"
	interactiongrpc "backend-diplomski/services/interaction-service/internal/grpc"
	_ "fmt"
	"log"
	"net"
	_ "net/http"
	_ "os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	dbPool, err := db.NewPool(&cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	queries := sqlc.New(dbPool)

	startGRPCServer(cfg.GRPCPort, queries)
}

func startGRPCServer(
	port string,
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

	grpcServer := grpc.NewServer()

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
