package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"video-worker/internal/config"
	"video-worker/internal/rabbit"
)

func main() {
	log.Println("video-worker starting...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("video-worker started")
	log.Printf("RabbitMQ URL: %s", cfg.RabbitMQURL)
	log.Printf("Queue: %s", cfg.RabbitMQShortVideoQueue)
	log.Printf("Upload service URL: %s", cfg.UploadServiceURL)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	consumer := rabbit.NewConusmer(
		cfg.RabbitMQURL,
		cfg.RabbitMQShortVideoQueue,
	)
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("consumer stopped with error:%v", err)
	}

}
