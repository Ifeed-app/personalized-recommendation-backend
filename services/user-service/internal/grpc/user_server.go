package grpc

import (
	userv1 "backend-diplomski/proto/gen/user/v1"
	"backend-diplomski/services/user-service/db/sqlc"
	auth "backend-diplomski/shared/auth"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "strings"
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

func (s *UserServer) GetCurrentUser(ctx context.Context, req *userv1.GetCurrentUserRequest) (*userv1.GetCurrentUserResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(
			codes.Unauthenticated,
			"authenticated user is missing",
		)
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

	return &userv1.GetCurrentUserResponse{
		User: &userv1.User{
			Id:        uuid.UUID(dbUser.ID.Bytes).String(),
			Email:     dbUser.Email,
			Username:  dbUser.Username,
			CreatedAt: dbUser.CreatedAt.Time.String(),
			UpdatedAt: dbUser.UpdatedAt.Time.String(),
		},
	}, nil
}
