package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"video-worker/internal/content"
	"video-worker/internal/downloader"
	"video-worker/internal/message"
	"video-worker/internal/providers/pexels"
	"video-worker/internal/upload"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Conusmer struct {
	rabbitURL     string
	queueName     string
	pexelsClient  *pexels.Client
	downloader    *downloader.Downloader
	uploadClient  *upload.Client
	contentClient *content.Client
}

func NewConusmer(rabbitURL string, queueName string, pexelsClient *pexels.Client, downloader *downloader.Downloader, uploadClient *upload.Client, contentClient *content.Client) *Conusmer {
	return &Conusmer{
		rabbitURL:     rabbitURL,
		queueName:     queueName,
		pexelsClient:  pexelsClient,
		downloader:    downloader,
		uploadClient:  uploadClient,
		contentClient: contentClient,
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

	searchCtx, searchCancel := context.WithTimeout(
		context.Background(),
		30*time.Second)
	videos, err := c.pexelsClient.SearchVideos(searchCtx, pexels.SearchParams{
		Query:              msg.Query,
		Orientation:        msg.Orientation,
		Limit:              msg.Limit,
		Page:               msg.Page,
		MinDurationSeconds: msg.MinDurationSeconds,
		MaxDurationSeconds: msg.MaxDurationSeconds,
	})
	searchCancel()
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
		videoCtx, videoCancel := context.WithTimeout(
			context.Background(),
			4*time.Minute)
		downloadedVideo, err := c.downloader.Download(videoCtx, video.VideoURL)
		if err != nil {
			videoCancel()
			log.Printf("failed to download video externalID=%v", err)
			continue
		}
		uploadResult, err := c.uploadClient.UploadVideo(videoCtx, upload.UploadVideoInput{
			Source:      video.Source,
			ExternalID:  video.ExternalID,
			FileName:    video.ExternalID + ".mp4",
			ContentType: downloadedVideo.ContentType,
			ObjectKey:   fmt.Sprintf("videos/imported/%s/%s/original.mp4", video.Source, video.ExternalID),
			Data:        downloadedVideo.Data,
		})
		if err != nil {
			videoCancel()
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
		log.Printf(
			"CREATE VIDEO REQUEST externalID=%s categories=%v tags=%v",
			video.ExternalID,
			[]string{msg.Query},
			[]string{},
		)
		createdVideo, err := c.contentClient.CreateVideo(
			videoCtx,
			content.CreateVideoInput{
				Title: msg.Query,
				Description: fmt.Sprintf(
					"Imported from Pexels. Author: %s. External ID: %s",
					video.Author,
					video.ExternalID,
				),
				StorageKey:      uploadResult.ObjectKey,
				ThumbnailKey:    video.ThumbnailURL,
				MimeType:        uploadResult.MimeType,
				DurationSeconds: int32(video.DurationSec),
				Status:          "READY",

				Categories: []string{
					msg.Query,
				},
				Tags: []string{},
			},
		)
		videoCancel()
		if err != nil {
			log.Printf("failed to create content video externalID=%s: %v", video.ExternalID, err)
			continue
		}
		log.Printf(
			"created content video externalId=%s videoId=%s storageKey=%s status=%s",
			video.ExternalID,
			createdVideo.ID,
			createdVideo.StorageKey,
			createdVideo.Status,
		)
	}

	return nil
}
