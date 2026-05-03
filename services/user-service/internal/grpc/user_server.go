package grpc

import (
	userv1 "backend-diplomski/proto/gen/user/v1"
	"backend-diplomski/services/user-service/db/sqlc"
	"context"
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
	return &userv1.GetUserByIdResponse{}, nil
}
