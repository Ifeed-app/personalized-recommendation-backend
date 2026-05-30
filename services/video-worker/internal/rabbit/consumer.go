package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"video-worker/internal/downloader"
	"video-worker/internal/message"
	"video-worker/internal/providers/pexels"
	"video-worker/internal/upload"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Conusmer struct {
	rabbitURL    string
	queueName    string
	pexelsClient *pexels.Client
	downloader   *downloader.Downloader
	uploadClient *upload.Client
}

func NewConusmer(rabbitURL string, queueName string, pexelsClient *pexels.Client, downloader *downloader.Downloader, uploadClient *upload.Client) *Conusmer {
	return &Conusmer{
		rabbitURL:    rabbitURL,
		queueName:    queueName,
		pexelsClient: pexelsClient,
		downloader:   downloader,
		uploadClient: uploadClient,
	}
}

func (c *Conusmer) Start(ctx context.Context) error {
	conn, err := c.connectWithRetry()
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open rabbit channel: %w", err)
	}
	defer ch.Close()
	queue, err := ch.QueueDeclare(
		c.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to delcare queue: %w", err)
	}
	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to set qos: %w", err)
	}
	deliveries, err := ch.Consume(
		queue.Name,
		"video-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming message: %w", err)
	}
	log.Printf("worker is listening to queue: %s", queue.Name)
	for {
		select {
		case <-ctx.Done():
			log.Println("video-worker consumer stopped")
			return nil

		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbit deliveries channel closed")
			}
			err := c.handleMessage(delivery)
			if err != nil {
				log.Printf("failed to handle message:%v", err)
				_ = delivery.Nack(false, false)
				continue
			}
			_ = delivery.Ack(false)
		}
	}

}

func (c *Conusmer) connectWithRetry() (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for attempt := 1; attempt <= 30; attempt++ {
		conn, err = amqp.Dial(c.rabbitURL)
		if err == nil {
			log.Println("Connected to rabbit")
			return conn, nil
		}
		log.Printf("Rabbit not ready, rettrying ... %d/30: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to rabbit: %w", err)
}

func (c *Conusmer) handleMessage(delivery amqp.Delivery) error {
	log.Printf("raw msg received: %s", string(delivery.Body))

	var msg message.ShortVideoFetchMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return fmt.Errorf("invalid message JSON: %w", err)
	}

	if msg.Provider == "" {
		msg.Provider = "pexels"
	}

	if msg.Provider != "pexels" {
		return fmt.Errorf("unsupported provider: %s", msg.Provider)
	}

	if msg.Query == "" {
		return fmt.Errorf("query is required")
	}

	if msg.Orientation == "" {
		msg.Orientation = "portrait"
	}

	if msg.Limit <= 0 {
		msg.Limit = 10
	}

	if msg.Page <= 0 {
		msg.Page = 1
	}

	if msg.MaxDurationSeconds <= 0 {
		msg.MaxDurationSeconds = 120
	}

	if msg.MinDurationSeconds <= 0 {
		msg.MinDurationSeconds = 3
	}

	log.Printf(
		"searching Pexels videos: jobId=%s query=%s orientation=%s limit=%d page=%d duration=%d-%d",
		msg.JobID,
		msg.Query,
		msg.Orientation,
		msg.Limit,
		msg.Page,
		msg.MinDurationSeconds,
		msg.MaxDurationSeconds,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	videos, err := c.pexelsClient.SearchVideos(ctx, pexels.SearchParams{
		Query:              msg.Query,
		Orientation:        msg.Orientation,
		Limit:              msg.Limit,
		Page:               msg.Page,
		MinDurationSeconds: msg.MinDurationSeconds,
		MaxDurationSeconds: msg.MaxDurationSeconds,
	})
	if err != nil {
		return fmt.Errorf("failed to search Pexels videos: %w", err)
	}

	log.Printf("Pexels returned %d usable videos", len(videos))

	for _, video := range videos {
		log.Printf(
			"video externalId=%s duration=%ds size=%dx%d author=%s url=%s",
			video.ExternalID,
			video.DurationSec,
			video.Width,
			video.Height,
			video.Author,
			video.VideoURL,
		)
		downloadedVideo, err := c.downloader.Download(ctx, video.VideoURL)
		if err != nil {
			log.Printf("failed o download video externalID=%v", err)
			continue
		}
		uploadResult, err := c.uploadClient.UploadVideo(ctx, upload.UploadVideoInput{
			Source:      video.Source,
			ExternalID:  video.ExternalID,
			FileName:    video.ExternalID + ".mp4",
			ContentType: downloadedVideo.ContentType,
			ObjectKey:   fmt.Sprintf("videos/imported/%s/%s/original.mp4", video.Source, video.ExternalID),
			Data:        downloadedVideo.Data,
		})
		if err != nil {
			log.Printf("failed to upload video externalID=%s : %v", video.ExternalID, err)
			continue
		}
		log.Printf(
			"downloaded video externalId=%s bucket=%s objectKey=%s sizeBytes=%d",
			video.ExternalID,
			uploadResult.Bucket,
			uploadResult.ObjectKey,
			uploadResult.SizeBytes,
		)
	}

	return nil
}
