package client

import (
	"context"
	"fmt"

	contentv1 "backend-diplomski/proto/gen/content/v1"
	"backend-diplomski/shared/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ContentClient struct {
	conn   *grpc.ClientConn
	client contentv1.ContentServiceClient
}

func NewContentClient(
	address string,
	serviceToken string,
) (*ContentClient, error) {

	credentials := auth.NewServiceCredentialsInterceptor(
		"interaction-simulator",
		serviceToken,
	)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(credentials.Unary),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to content-service: %w",
			err,
		)
	}

	return &ContentClient{
		conn:   conn,
		client: contentv1.NewContentServiceClient(conn),
	}, nil
}

func (c *ContentClient) Close() error {
	return c.conn.Close()
}

func (c *ContentClient) ListRecommendationCandidates(
	ctx context.Context,
	limit int32,
) ([]*contentv1.RecommendationCandidate, error) {

	response, err := c.client.ListRecommendationCandidates(
		ctx,
		&contentv1.ListRecommendationCandidatesRequest{
			Limit: limit,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"ListRecommendationCandidates failed: %w",
			err,
		)
	}

	return response.GetVideos(), nil
}
