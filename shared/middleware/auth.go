package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIdContextKey = "authenticated-user-id"

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthInterceptor struct {
	jwtSecret     []byte
	publicMethods map[string]struct{}
}

func NewAuthInterceptor(
	jwtSecret string,
	publicMethods ...string) *AuthInterceptor {
	methods := make(map[string]struct{}, len(publicMethods))
	for _, method := range publicMethods {
		methods[method] = struct{}{}
	}
	return &AuthInterceptor{
		jwtSecret:     []byte(jwtSecret),
		publicMethods: methods,
	}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if _, public := a.publicMethods[info.FullMethod]; public {
			return handler(ctx, req)
		}
		userID, err := a.authenticate(ctx)
		if err != nil {
			return nil, err
		}
		authenticatedContext := context.WithValue(
			ctx,
			userIdContextKey,
			userID,
		)
		return handler(authenticatedContext, req)
	}

}

func (a *AuthInterceptor) authenticate(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(
			codes.Unauthenticated,
			"missing required metadata")
	}
	authorizationValues := md.Get("authorization")
	if len(authorizationValues) == 0 {
		return "", status.Error(
			codes.Unauthenticated,
			"missing authorization token")
	}
	const bearerPrefix = "Bearer "

	authorization := authorizationValues[0]
	if !strings.HasPrefix(authorization, bearerPrefix) {
		return "", status.Error(
			codes.Unauthenticated,
			"invalid authorization format")
	}
	rawToken := strings.TrimSpace(
		strings.TrimPrefix(authorization, bearerPrefix))
	if rawToken == "" {
		return "", status.Error(
			codes.Unauthenticated,
			"authorization token is empty")
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return a.jwtSecret, nil
		},
	)
	if err != nil || !token.Valid {
		return "", status.Error(
			codes.Unauthenticated,
			"invalid or expired access token")
	}
	if claims.UserID == "" {
		return "", status.Error(
			codes.Unauthenticated,
			"access token does not contain userID")
	}
	return claims.UserID, nil
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIdContextKey).(string)
	return userID, ok && userID != ""
}
