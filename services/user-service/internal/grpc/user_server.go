package grpc

import (
	userv1 "backend-diplomski/proto/gen/user/v1"
	"backend-diplomski/services/user-service/db/sqlc"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	queries *sqlc.Queries
}

func NewUserServer(queries *sqlc.Queries) *UserServer {
	return &UserServer{
		queries: queries,
	}
}

func (s *UserServer) GetUserById(ctx context.Context, req *userv1.GetUserByIdRequest) (*userv1.GetUserByIdResponse, error) {
	userID := strings.TrimSpace(req.GetId())

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	dbUser, err := s.queries.GetUserByID(ctx, pgtype.UUID{
		Bytes: parsedUserID,
		Valid: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &userv1.GetUserByIdResponse{
		User: &userv1.User{
			Id:        uuid.UUID(dbUser.ID.Bytes).String(),
			Email:     dbUser.Email,
			Username:  dbUser.Username,
			CreatedAt: dbUser.CreatedAt.Time.String(),
			UpdatedAt: dbUser.UpdatedAt.Time.String(),
		},
	}, nil
}
