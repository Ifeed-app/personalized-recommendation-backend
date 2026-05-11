package grpc

import (
	userv1 "backend-diplomski/proto/gen/user/v1"
	"backend-diplomski/services/user-service/db/sqlc"
	"backend-diplomski/services/user-service/internal/auth"
	"backend-diplomski/services/user-service/internal/config"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	userv1.UnimplementedAuthServiceServer
	queries    *sqlc.Queries
	jwtManager *auth.JWTManager
	cgf        *config.Config
}

func NewAuthServer(queries *sqlc.Queries, jwtManager *auth.JWTManager, cfg *config.Config) *AuthServer {
	return &AuthServer{
		queries:    queries,
		jwtManager: jwtManager,
		cgf:        cfg,
	}
}

func (s *AuthServer) SignUp(ctx context.Context, req *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.GetEmail()))
	username := strings.TrimSpace(req.GetUsername())
	password := req.GetPassword()

	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is requeired")
	}

	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "Username is requeired")
	}

	if password == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is requeired")
	}

	if len(password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	_, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "User already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password : %v", err)
	}
	NewUserID := uuid.New()

	dbUser, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID: pgtype.UUID{
			Bytes: NewUserID,
			Valid: true,
		},
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	userId := uuid.UUID(dbUser.ID.Bytes).String()
	accessToken, err := s.jwtManager.GenerateToken(userId, dbUser.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to geberate access token: %v", err)
	}
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}
	refreshTokenID := uuid.New()
	refreshExpiresAt := time.Now().Add(time.Duration(s.cgf.RefreshTokenTTLHours) * time.Hour)
	_, err = s.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		ID: pgtype.UUID{
			Bytes: [16]byte(refreshTokenID),
			Valid: true,
		},
		UserID: dbUser.ID,
		Token:  refreshToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  refreshExpiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store refresh token: %v", err)
	}
	return &userv1.SignUpResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserInfo: &userv1.UserInfo{
			Id:       uuid.UUID(dbUser.ID.Bytes).String(),
			Email:    dbUser.Email,
			Username: dbUser.Username,
		},
	}, nil
}

func (s *AuthServer) SignIn(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.GetEmail()))
	password := req.GetPassword()

	if email == "" || password == "" {
		return nil, status.Error(codes.InvalidArgument, "Password or email are required")
	}
	dbUser, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}
	accessToken, err := s.jwtManager.GenerateToken(uuid.UUID(dbUser.ID.Bytes).String(), dbUser.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate access token %v", err)
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate refresh token %v ", err)
	}

	refreshTokenID := uuid.New()
	refreshExpiresAt := time.Now().Add(time.Duration(s.cgf.RefreshTokenTTLHours) * time.Hour)
	_, err = s.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		ID: pgtype.UUID{
			Bytes: [16]byte(refreshTokenID),
			Valid: true,
		},
		UserID: dbUser.ID,
		Token:  refreshToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  refreshExpiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store refresh token: %v", err)
	}

	return &userv1.SignInResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserInfo: &userv1.UserInfo{
			Id:       uuid.UUID(dbUser.ID.Bytes).String(),
			Email:    dbUser.Email,
			Username: dbUser.Username,
		},
	}, nil
}
