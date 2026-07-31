package upload

import (
	uploadv1 "backend-diplomski/proto/gen/upload/v1"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const chunkSize = 1024 * 1024

type Client struct {
	conn   *grpc.ClientConn
	client uploadv1.UploadServiceClient
}

type UploadVideoInput struct {
	Source      string
	ExternalID  string
	FileName    string
	ContentType string
	ObjectKey   string
	Data        []byte
}

type UploadVideoResult struct {
	Bucket           string
	ObjectKey        string
	MimeType         string
	SizeBytes        int64
	OriginalFileName string
}

func NewCLient(target string) (*Client, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload serrvice grpc clinet: %w", err)
	}
	return &Client{
		conn:   conn,
		client: uploadv1.NewUploadServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) UploadVideo(ctx context.Context, input UploadVideoInput) (*UploadVideoResult, error) {
	if input.ExternalID == "" {
		return nil, fmt.Errorf("external ID is req")
	}

	if len(input.Data) == 0 {
		return nil, fmt.Errorf("video data is empty")
	}

	if input.Source == "" {
		input.Source = "pexels"
	}

	if input.FileName == "" {
		input.FileName = input.ExternalID + ".mp4"
	}

	if input.ContentType == "" {
		input.ContentType = "video/mp4"
	}

	if input.ObjectKey == "" {
		input.ObjectKey = fmt.Sprintf("videos/imported/%s/%s/original.mp4", input.Source, input.ExternalID)
	}
	stream, err := c.client.UploadVideo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open upload stream : %w", err)
	}
	err = stream.Send(&uploadv1.UploadVideoRequest{
		Payload: &uploadv1.UploadVideoRequest_Metadata{
			Metadata: &uploadv1.UploadVideoMetaData{
				Source:      input.Source,
				ExternalId:  input.ExternalID,
				FileName:    input.FileName,
				ContentType: input.ContentType,
				ObjectKey:   input.ObjectKey,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send upload metada: %w", err)
	}
	for start := 0; start < len(input.Data); start += chunkSize {
		end := start + chunkSize
		if end > len(input.Data) {
			end = len(input.Data)
		}
		err = stream.Send(&uploadv1.UploadVideoRequest{
			Payload: &uploadv1.UploadVideoRequest_Chunk{
				Chunk: input.Data[start:end],
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send upload chunk : %w", err)
		}
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("upload service returned error: %w", err)
	}
	return &UploadVideoResult{
		Bucket:           response.Bucket,
		ObjectKey:        response.ObjectKey,
		MimeType:         response.MimeType,
		SizeBytes:        response.SizeBytes,
		OriginalFileName: response.OriginalFileName,
	}, nil
}
