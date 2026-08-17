package grpc

import (
	userv1 "backend-diplomski/proto/gen/user/v1"
	"backend-diplomski/services/user-service/db/sqlc"
	refreshToken "backend-diplomski/services/user-service/internal/auth"
	"backend-diplomski/services/user-service/internal/config"
	"backend-diplomski/services/user-service/internal/messaging"
	auth "backend-diplomski/shared/auth"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	userv1.UnimplementedAuthServiceServer

	pool           *pgxpool.Pool
	queries        *sqlc.Queries
	jwtManager     *auth.JWTManager
	cfg            *config.Config
	emailPublisher *messaging.EmailPublisher
}

func NewAuthServer(pool *pgxpool.Pool, queries *sqlc.Queries, jwtManager *auth.JWTManager, cfg *config.Config, emailPublisher *messaging.EmailPublisher) *AuthServer {
	return &AuthServer{
		pool:           pool,
		queries:        queries,
		jwtManager:     jwtManager,
		cfg:            cfg,
		emailPublisher: emailPublisher,
	}
}

func (s *AuthServer) SignUp(ctx context.Context, req *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.GetEmail()))
	username := strings.TrimSpace(req.GetUsername())
	password := req.GetPassword()
	gender := strings.ToUpper(strings.TrimSpace(req.GetGender()))
	countryCode := strings.TrimSpace(req.GetCountryCode())
	preferredLanguage := strings.ToLower(
		strings.TrimSpace(req.GetPreferredLanguage()),
	)
	interests := req.GetInterests()
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

	if req.GetBirthYear() == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"birth year is required",
		)
	}

	currentYear := int32(time.Now().Year())
	if req.GetBirthYear() < 1900 || req.GetBirthYear() > currentYear {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid birth year",
		)
	}

	if countryCode == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"country code is required",
		)
	}

	if len(countryCode) != 2 {
		return nil, status.Error(
			codes.InvalidArgument,
			"country code must contain two letters",
		)
	}

	if preferredLanguage == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"preferred language is required",
		)
	}

	if len(interests) < 3 {
		return nil, status.Error(
			codes.InvalidArgument,
			"at least three interests are required",
		)
	}

	_, err := s.queries.GetUserByEmail(ctx, email)
	switch {
	case err == nil:
		return nil, status.Error(
			codes.AlreadyExists,
			"user with that email already exists",
		)
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, status.Errorf(
			codes.Internal,
			"failed to check email availability: %v",
			err,
		)
	}

	_, err = s.queries.GetUserByUsername(ctx, username)
	switch {
	case err == nil:
		return nil, status.Error(
			codes.AlreadyExists,
			"user with that username already exists",
		)
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, status.Errorf(
			codes.Internal,
			"failed to check username availability: %v",
			err,
		)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password : %v", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to start registration transaction: %v",
			err,
		)
	}
	defer tx.Rollback(ctx)

	txQueries := s.queries.WithTx(tx)
	NewUserID := uuid.New()

	dbUser, err := txQueries.CreateUser(ctx, sqlc.CreateUserParams{
		ID: pgtype.UUID{
			Bytes: NewUserID,
			Valid: true,
		},
		Email:             email,
		Username:          username,
		PasswordHash:      string(hashedPassword),
		BirthYear:         nullableBirthYear(req.GetBirthYear()),
		Gender:            nullableText(gender),
		CountryCode:       nullableText(countryCode),
		PreferredLanguage: nullableText(preferredLanguage),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	uniqueInterests := make(map[string]struct{})
	for _, rawInterest := range interests {
		interest := strings.ToLower(strings.TrimSpace(rawInterest))
		if interest == "" {
			continue
		}

		if _, exists := uniqueInterests[interest]; exists {
			continue
		}

		uniqueInterests[interest] = struct{}{}

		_, err = txQueries.CreateUserInterest(
			ctx,
			sqlc.CreateUserInterestParams{
				UserID:   dbUser.ID,
				Interest: interest,
				Weight:   1.0,
			},
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"failed to save user interest: %v",
				err,
			)
		}
	}

	if len(uniqueInterests) < 3 {
		return nil, status.Error(
			codes.InvalidArgument,
			"at least three unique interests are required",
		)
	}

	_, err = txQueries.CompleteUserOnboarding(ctx, dbUser.ID)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to complete onboarding: %v",
			err,
		)
	}
	verificationTokenID := uuid.New()

	rawVerificationToken, tokenHash, err := generateVerificationToken()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to generate verification token: %v",
			err)
	}
	_, err = txQueries.CreateEmailVerificationToken(
		ctx,
		sqlc.CreateEmailVerificationTokenParams{
			ID: pgtype.UUID{
				Bytes: verificationTokenID,
				Valid: true,
			},
			UserID:    dbUser.ID,
			TokenHash: tokenHash,
			ExpiresAt: pgtype.Timestamptz{
				Time: time.Now().Add(
					s.cfg.EmailVerificationTokenTTLMinutes * time.Minute,
				),
				Valid: true,
			},
		},
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create email verification token: %v",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to complete registration: %v",
			err,
		)
	}
	userID := uuid.UUID(dbUser.ID.Bytes).String()
	verificationURL := fmt.Sprintf(
		"%s?token=%s",
		s.cfg.EmailVerificationBaseURL,
		rawVerificationToken)
	if s.emailPublisher == nil {
		log.Printf(
			"email publisher is not configured; verification email not sent for user %s",
			userID,
		)
	} else {
		err = s.emailPublisher.PublishVerificationEmail(
			ctx,
			messaging.EmailVerificationMessage{
				UserID:            userID,
				Email:             dbUser.Email,
				Username:          dbUser.Username,
				VerificationToken: rawVerificationToken,
				VerificationURL:   verificationURL,
			})
		if err != nil {
			log.Printf(
				"failed to publish verification email for user %s: %v",
				userID,
				err,
			)
		}
	}
	return &userv1.SignUpResponse{
		UserInfo: &userv1.UserInfo{
			Id:       userID,
			Email:    dbUser.Email,
			Username: dbUser.Username,
		},
	}, nil
}

func (s *AuthServer) SignIn(
	ctx context.Context,
	req *userv1.SignInRequest,
) (*userv1.SignInResponse, error) {
	identifier := strings.TrimSpace(req.GetIdentifier())
	password := req.GetPassword()

	if identifier == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"email or username is required",
		)
	}

	if password == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"password is required",
		)
	}

	dbUser, err := s.queries.GetUserByLoginIdentifier(
		ctx,
		identifier,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid credentials")
		}
		return nil, status.Errorf(
			codes.Internal,
			"failed to get user: %v",
			err,
		)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(dbUser.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, status.Error(
			codes.Unauthenticated,
			"invalid credentials",
		)
	}

	if !dbUser.EmailVerified {
		return nil, status.Error(
			codes.PermissionDenied,
			"email address is not verified",
		)
	}

	userID := uuid.UUID(dbUser.ID.Bytes).String()

	accessToken, err := s.jwtManager.GenerateToken(
		userID,
		dbUser.Email,
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to generate access token: %v",
			err,
		)
	}

	newRefreshToken, err := refreshToken.GenerateRefreshToken()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to generate refresh token: %v",
			err,
		)
	}

	refreshTokenID := uuid.New()
	refreshExpiresAt := time.Now().Add(
		time.Duration(s.cfg.RefreshTokenTTLHours) * time.Hour,
	)

	_, err = s.queries.CreateRefreshToken(
		ctx,
		sqlc.CreateRefreshTokenParams{
			ID: pgtype.UUID{
				Bytes: refreshTokenID,
				Valid: true,
			},
			UserID: dbUser.ID,
			Token:  newRefreshToken,
			ExpiresAt: pgtype.Timestamptz{
				Time:  refreshExpiresAt,
				Valid: true,
			},
		},
	)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to store refresh token: %v",
			err,
		)
	}

	return &userv1.SignInResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserInfo: &userv1.UserInfo{
			Id:       userID,
			Email:    dbUser.Email,
			Username: dbUser.Username,
		},
	}, nil
}

func (s *AuthServer) VerifyEmail(
	ctx context.Context,
	req *userv1.VerifyEmailRequest,
) (*userv1.VerifyEmailResponse, error) {
	if err := s.verifyEmailToken(
		ctx,
		req.GetToken()); err != nil {
		return nil, err
	}

	return &userv1.VerifyEmailResponse{
		Verified: true,
		Message:  "Email verified successfully",
	}, nil
}

func (s *AuthServer) VerifyEmailHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	token := request.URL.Query().Get("token")

	if err := s.verifyEmailToken(
		request.Context(),
		token,
	); err != nil {
		http.Error(
			writer,
			status.Convert(err).Message(),
			http.StatusBadRequest,
		)
		return
	}

	writer.Header().Set(
		"Content-Type",
		"text/html; charset=utf-8",
	)

	_, _ = writer.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<body>
			<h1>Email verified</h1>
			<p>You can now sign in to Feeder.</p>
		</body>
		</html>
	`))
}

func (s *AuthServer) verifyEmailToken(
	ctx context.Context,
	rawToken string,
) error {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return status.Error(
			codes.InvalidArgument,
			"token is required",
		)
	}

	tokenHash := hashVerificationToken(rawToken)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return status.Errorf(
			codes.Internal,
			"failed to start verification transaction: %v",
			err,
		)
	}
	defer tx.Rollback(ctx)

	txQueries := s.queries.WithTx(tx)

	dbToken, err := txQueries.GetEmailVerificationTokenByHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(
				codes.NotFound,
				"verification token is invalid",
			)
		}

		return status.Errorf(
			codes.Internal,
			"failed to get verification token: %v",
			err,
		)
	}

	if dbToken.UsedAt.Valid {
		return status.Error(
			codes.FailedPrecondition,
			"token has already been used",
		)
	}

	if time.Now().After(dbToken.ExpiresAt.Time) {
		return status.Error(
			codes.FailedPrecondition,
			"verification token has expired",
		)
	}

	if err := txQueries.MarkUserEmailVerified(
		ctx,
		dbToken.UserID,
	); err != nil {
		return status.Errorf(
			codes.Internal,
			"failed to verify email: %v",
			err,
		)
	}

	_, err = txQueries.MarkEmailVerificationTokenUsed(
		ctx,
		dbToken.ID,
	)
	if err != nil {
		return status.Errorf(
			codes.Internal,
			"failed to mark verification token as used: %v",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return status.Errorf(
			codes.Internal,
			"failed to complete verification: %v",
			err,
		)
	}

	return nil
}

func nullableText(value string) pgtype.Text {
	value = strings.TrimSpace(value)

	return pgtype.Text{
		String: value,
		Valid:  value != "",
	}
}

func nullableBirthYear(value int32) pgtype.Int2 {
	if value == 0 {
		return pgtype.Int2{}
	}

	return pgtype.Int2{
		Int16: int16(value),
		Valid: true,
	}
}

func generateVerificationToken() (rawToken string, tokenHash string, err error) {
	tokenBytes := make([]byte, 32)

	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	rawToken = base64.RawURLEncoding.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(hash[:])

	return rawToken, tokenHash, nil
}

func hashVerificationToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])

}
