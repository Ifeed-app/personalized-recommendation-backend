package content

import (
	contentv1 "backend-diplomski/proto/gen/content/v1"
	"backend-diplomski/shared/auth"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client contentv1.ContentServiceClient
}

type CreateVideoInput struct {
	Title           string
	Description     string
	StorageKey      string
	ThumbnailKey    string
	MimeType        string
	DurationSeconds int32
	Status          string
}

type CreateVideoResult struct {
	ID         string
	Title      string
	StorageKey string
	Status     string
}

func NewClient(
	address string,
	serviceToken string,
) (*Client, error) {
	credentials := auth.NewServiceCredentialsInterceptor(
		"video-worker",
		serviceToken,
	)

	connection, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithUnaryInterceptor(
			credentials.Unary,
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"connect to content service: %w",
			err,
		)
	}

	return &Client{
		conn:   connection,
		client: contentv1.NewContentServiceClient(connection),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateVideo(ctx context.Context, input CreateVideoInput) (*CreateVideoResult, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if input.StorageKey == "" {
		return nil, fmt.Errorf("storageKey is required")
	}

	if input.MimeType == "" {
		input.MimeType = "video/mp4"
	}

	if input.Status == "" {
		input.Status = "READY"
	}

	response, err := c.client.CreateVideo(ctx, &contentv1.CreateVideoRequest{
		Title:          input.Title,
		Description:    input.Description,
		StroageKey:     input.StorageKey,
		ThumbnailKey:   input.ThumbnailKey,
		MimeType:       input.MimeType,
		DurationSecond: input.DurationSeconds,
		Status:         input.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("content-service returned error:%w", err)
	}
	return &CreateVideoResult{
		ID:         response.Id,
		Title:      response.Title,
		StorageKey: response.StroageKey,
		Status:     response.Status,
	}, nil
}
