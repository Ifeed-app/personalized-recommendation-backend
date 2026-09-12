package client

import (
	"context"
	"fmt"

	interactionv1 "backend-diplomski/proto/gen/interaction/v1"
	"backend-diplomski/shared/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InteractionClient struct {
	conn   *grpc.ClientConn
	client interactionv1.InteractionServiceClient
}

func NewInteractionClient(
	address string,
	serviceToken string,
) (*InteractionClient, error) {

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
			"failed to connect to interaction-service: %w",
			err,
		)
	}

	return &InteractionClient{
		conn:   conn,
		client: interactionv1.NewInteractionServiceClient(conn),
	}, nil
}

func (c *InteractionClient) Close() error {
	return c.conn.Close()
}

func (c *InteractionClient) CreateInteraction(
	ctx context.Context,
	userID string,
	videoID string,
	interactionType interactionv1.InteractionType,
) error {

	_, err := c.client.CreateInteraction(
		ctx,
		&interactionv1.CreateInteractionRequest{
			UserId:          userID,
			VideoId:         videoID,
			InteractionType: interactionType,
		},
	)

	if err != nil {
		return fmt.Errorf("CreateInteraction failed: %w", err)
	}

	return nil
}

func (c *InteractionClient) RecordWatchSession(
	ctx context.Context,
	userID string,
	videoID string,
	watchedSeconds int32,
	completed bool,
	startedAt string,
	endedAt string,
) error {

	_, err := c.client.RecordWatchSession(
		ctx,
		&interactionv1.RecordWatchSessionRequest{
			UserId:         userID,
			VideoId:        videoID,
			WatchedSeconds: watchedSeconds,
			Completed:      completed,
			StartedAt:      startedAt,
			EndedAt:        endedAt,
		},
	)

	if err != nil {
		return fmt.Errorf("RecordWatchSession failed: %w", err)
	}

	return nil
}
