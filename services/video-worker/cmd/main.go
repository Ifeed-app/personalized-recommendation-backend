package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"video-worker/internal/config"
	"video-worker/internal/content"
	"video-worker/internal/downloader"
	"video-worker/internal/providers/pexels"
	"video-worker/internal/rabbit"
	"video-worker/internal/scheduler"
	"video-worker/internal/upload"
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
	pexelsClient := pexels.NewClient(cfg.PexelsAPIKey, cfg.PexelsAPIURL)
	videoDownloader := downloader.NewDownloader()
	uploadClient, err := upload.NewCLient(cfg.UploadServiceGRPCURL)
	if err != nil {
		log.Fatalf("failed to create uploade service client : %w", err)
	}
	defer uploadClient.Close()

	contentClient, err := content.NewClient(cfg.ContentServiceGRPCURL)
	if err != nil {
		log.Fatalf("failed to create content-service client : %v", err)
	}
	defer contentClient.Close()

	consumer := rabbit.NewConusmer(
		cfg.RabbitMQURL,
		cfg.RabbitMQShortVideoQueue,
		pexelsClient,
		videoDownloader,
		uploadClient,
		contentClient,
	)
	if cfg.VideoFetchAutoStart {
		publisher := rabbit.NewPublisher(
			cfg.RabbitMQURL,
			cfg.RabbitMQShortVideoQueue,
		)

		fetchScheduler := scheduler.NewFetchScheduler(
			publisher,
			cfg.VideoFetchIntervalSeconds,
			cfg.VideoFetchQueries,
			cfg.VideoFetchLimit,
			cfg.VideoFetchOrientation,
			cfg.VideoFetchMinDurationSeconds,
			cfg.VideoFetchMaxDurationSeconds,
		)

		log.Println("video fetcher scheduler enabled")
		go fetchScheduler.Start(ctx)
	} else {
		log.Println("video fetcher scheduler disabled")
	}
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("consumer stopped with error:%v", err)
	}

}
