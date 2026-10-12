package content

import (
	"context"
	"fmt"

	contentv1 "backend-diplomski/proto/gen/content/v1"
	"backend-diplomski/shared/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client contentv1.ContentServiceClient
}

func NewClient(
	address string,
	serviceToken string,
) (*Client, error) {

	serviceCredentials := auth.NewServiceCredentialsInterceptor(
		"recommendation-service",
		serviceToken,
	)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithUnaryInterceptor(
			serviceCredentials.Unary,
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create content-service client: %w",
			err,
		)
	}

	return &Client{
		conn:   conn,
		client: contentv1.NewContentServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetVideoMetadata(
	ctx context.Context,
	videoID string,
) (*contentv1.GetVideoMetadataResponse, error) {

	response, err := c.client.GetVideoMetadata(
		ctx,
		&contentv1.GetVideoMetadataRequest{
			VideoId: videoID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"content-service returned error: %w",
			err,
		)
	}

	return response, nil
}

func (c *Client) ListRecommendationCandidates(
	ctx context.Context,
	limit int32,
) (*contentv1.ListRecommendationCandidatesResponse, error) {

	response, err := c.client.ListRecommendationCandidates(
		ctx,
		&contentv1.ListRecommendationCandidatesRequest{
			Limit: limit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"content-service returned error: %w",
			err,
		)
	}

	return response, nil
}
