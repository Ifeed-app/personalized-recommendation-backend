package grpc

import (
	interactionv1 "backend-diplomski/proto/gen/interaction/v1"
	db "backend-diplomski/services/interaction-service/db/sqlc"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	interactionv1.UnimplementedInteractionServiceServer
	queries *db.Queries
}

func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl any) {
	//TODO implement me
	panic("implement me")
}

func NewServer(queries *db.Queries) *Server {
	return &Server{
		queries: queries,
	}

}

func (s *Server) CreateInteraction(
	ctx context.Context,
	req *interactionv1.CreateInteractionRequest) (*interactionv1.CreateInteractionResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	videoID, err := parseUUID(req.GetVideoId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid video_id")
	}

	interactionType, err := mapInteractionType(req.GetInteractionType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if interactionType == "LIKE" || interactionType == "UNLIKE" {
		latestType, err :=
			s.queries.GetLatestInteractionTypeByUserAndVideo(
				ctx,
				db.GetLatestInteractionTypeByUserAndVideoParams{
					UserID:  userID,
					VideoID: videoID,
				},
			)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(
				codes.Internal,
				"failed to check latest interaction: %v",
				err,
			)
		}

		if latestType == interactionType {
			return nil, status.Errorf(
				codes.AlreadyExists,
				"interaction %s already exists for this video",
				interactionType,
			)
		}
	}

	interactionID, err := parseUUID(uuid.NewString())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate interaction id")
	}

	created, err := s.queries.CreateVideoInteraction(
		ctx,
		db.CreateVideoInteractionParams{
			ID:              interactionID,
			UserID:          userID,
			VideoID:         videoID,
			InteractionType: interactionType,
		},
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create interaction: %v",
			err,
		)
	}

	return &interactionv1.CreateInteractionResponse{
		Id:              uuidToString(created.ID),
		UserId:          uuidToString(created.UserID),
		VideoId:         uuidToString(created.VideoID),
		InteractionType: req.GetInteractionType(),
		CreatedAt:       timestampToString(created.CreatedAt),
	}, nil

}

func (s *Server) RecordWatchSession(
	ctx context.Context,
	req *interactionv1.RecordWatchSessionRequest,
) (*interactionv1.RecordWatchSessionResponse, error) {
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	videoID, err := parseUUID(req.GetVideoId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid video_id")
	}

	if req.GetWatchedSeconds() < 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"watched_seconds cannot be negative",
		)
	}

	startedAt, err := parseTime(req.GetStartedAt())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid started_at")
	}

	endedAt, err := parseTime(req.GetEndedAt())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ended_at")
	}

	if endedAt.Before(startedAt) {
		return nil, status.Error(
			codes.InvalidArgument,
			"ended_at cannot be before started_at",
		)
	}

	sessionID, err := parseUUID(uuid.NewString())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate session id")
	}

	created, err := s.queries.CreateVideoWatchSession(
		ctx,
		db.CreateVideoWatchSessionParams{
			ID:             sessionID,
			UserID:         userID,
			VideoID:        videoID,
			WatchedSeconds: req.GetWatchedSeconds(),
			Completed:      req.GetCompleted(),
			StartedAt: pgtype.Timestamptz{
				Time:  startedAt,
				Valid: true,
			},
			EndedAt: pgtype.Timestamptz{
				Time:  endedAt,
				Valid: true,
			},
		},
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to record watch session: %v",
			err,
		)
	}

	return &interactionv1.RecordWatchSessionResponse{
		Id:             uuidToString(created.ID),
		UserId:         uuidToString(created.UserID),
		VideoId:        uuidToString(created.VideoID),
		WatchedSeconds: created.WatchedSeconds,
		Completed:      created.Completed,
		StartedAt:      timestampToString(created.StartedAt),
		EndedAt:        timestampToString(created.EndedAt),
	}, nil
}

func (s *Server) GetVideoInteractionSummary(
	ctx context.Context,
	req *interactionv1.GetVideoInteractionSummaryRequest) (
	*interactionv1.GetVideoInteractionSummaryResponse, error) {
	videoID, err := parseUUID(req.GetVideoId())
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid video_id")
	}
	userID, err := parseUUID(req.GetUserId())
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid user_id")
	}
	likeCount, err := s.queries.CountCurrentLikesByVideoID(
		ctx, videoID)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to count video likes: %v",
			err)
	}
	likedByUser, err := s.queries.IsVideoCurrentlyLikedByUser(
		ctx,
		db.IsVideoCurrentlyLikedByUserParams{
			userID,
			videoID,
		},
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to check user like status: %v",
			err)
	}
	return &interactionv1.GetVideoInteractionSummaryResponse{
		LikeCount:   likeCount,
		LikedByUser: likedByUser,
	}, nil
}

func mapInteractionType(
	value interactionv1.InteractionType,
) (string, error) {
	switch value {
	case interactionv1.InteractionType_INTERACTION_TYPE_LIKE:
		return "LIKE", nil

	case interactionv1.InteractionType_INTERACTION_TYPE_UNLIKE:
		return "UNLIKE", nil

	case interactionv1.InteractionType_INTERACTION_TYPE_SKIP:
		return "SKIP", nil

	default:
		return "", fmt.Errorf("interaction_type is required")
	}
}

func parseUUID(value string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{
		Bytes: parsed,
		Valid: true,
	}, nil
}

func uuidToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}

	id, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}

	return id.String()
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}

func timestampToString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}

	return value.Time.UTC().Format(time.RFC3339Nano)
}
