package interaction

import (
	"context"
	"fmt"

	interactionv1 "backend-diplomski/proto/gen/interaction/v1"
	"backend-diplomski/shared/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client interactionv1.InteractionServiceClient
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
			"failed to create interaction-service client: %w",
			err,
		)
	}

	return &Client{
		conn:   conn,
		client: interactionv1.NewInteractionServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetUserInteractionHistory(
	ctx context.Context,
	userID string,
) (*interactionv1.GetUserInteractionHistoryResponse, error) {

	response, err := c.client.GetUserInteractionHistory(
		ctx,
		&interactionv1.GetUserInteractionHistoryRequest{
			UserId: userID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"interaction-service returned error: %w",
			err,
		)
	}

	return response, nil
}
