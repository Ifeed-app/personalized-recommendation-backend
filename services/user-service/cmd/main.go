package main

import (
	userv1 "backend-diplomski/proto/gen/user/v1"
	"backend-diplomski/services/user-service/db/sqlc"
	_ "backend-diplomski/services/user-service/internal/auth"
	"backend-diplomski/services/user-service/internal/config"
	"backend-diplomski/services/user-service/internal/db"
	servicegrpc "backend-diplomski/services/user-service/internal/grpc"
	"backend-diplomski/services/user-service/internal/messaging"
	auth "backend-diplomski/shared/auth"
	"log"
	"net"
	"net/http"
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

	authInterceptor := auth.NewUnaryAuthInterceptor(
		jwtManager,

		"/user.v1.AuthService/SignUp",
		"/user.v1.AuthService/SignIn",
		"/user.v1.AuthService/RefreshToken",
		"/user.v1.AuthService/ForgotPassword",
	)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			authInterceptor.Intercept,
		),
	)
	emailPublisher, err := messaging.NewEmailPublisher(
		cfg.RabbitMQURL,
		cfg.EmailVerificationQueue)
	if err != nil {
		log.Fatalf(
			"failed to initialize email publisher: %t",
			err)

	}
	defer emailPublisher.Close()
	log.Printf(
		"email publisher configured: %t",
		emailPublisher != nil)
	userServer := servicegrpc.NewUserServer(queries)
	authServer := servicegrpc.NewAuthServer(pool, queries, jwtManager, cfg, emailPublisher)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/verify-email", authServer.VerifyEmailHTTP)

	go func() {
		log.Printf("user-service HTTP listening on port %s", cfg.HTTPPort)

		if err := http.ListenAndServe(":"+cfg.HTTPPort, httpMux); err != nil {
			log.Fatalf("failed to serve HTTP server: %v", err)
		}
	}()
	userv1.RegisterUserServiceServer(grpcServer, userServer)
	userv1.RegisterAuthServiceServer(grpcServer, authServer)
	log.Printf("user-service connected to database and sqlc initialized")
	log.Printf("user-service gRPC listening on port %s", cfg.AppPort)
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
}
