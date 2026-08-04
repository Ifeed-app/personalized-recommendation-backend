package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UnaryAuthInterceptor struct {
	jwtManager    *JWTManager
	publicMethods map[string]struct{}
}

func NewUnaryAuthInterceptor(
	jwtManager *JWTManager,
	publicMethods ...string,
) *UnaryAuthInterceptor {
	publicMethodSet := make(
		map[string]struct{},
		len(publicMethods),
	)

	for _, method := range publicMethods {
		publicMethodSet[method] = struct{}{}
	}

	return &UnaryAuthInterceptor{
		jwtManager:    jwtManager,
		publicMethods: publicMethodSet,
	}
}

func (i *UnaryAuthInterceptor) Intercept(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if _, public := i.publicMethods[info.FullMethod]; public {
		return handler(ctx, req)
	}

	rawToken, err := extractBearerToken(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := i.jwtManager.ValidateToken(rawToken)
	if err != nil {
		return nil, status.Error(
			codes.Unauthenticated,
			"invalid or expired access token",
		)
	}

	authenticatedContext := ContextWithClaims(
		ctx,
		claims,
	)

	return handler(authenticatedContext, req)
}

func extractBearerToken(
	ctx context.Context,
) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(
			codes.Unauthenticated,
			"missing request metadata",
		)
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(
			codes.Unauthenticated,
			"missing authorization token",
		)
	}

	const prefix = "Bearer "

	authorization := values[0]

	if !strings.HasPrefix(authorization, prefix) {
		return "", status.Error(
			codes.Unauthenticated,
			"authorization must use Bearer scheme",
		)
	}

	rawToken := strings.TrimSpace(
		strings.TrimPrefix(authorization, prefix),
	)

	if rawToken == "" {
		return "", status.Error(
			codes.Unauthenticated,
			"authorization token is empty",
		)
	}

	return rawToken, nil
}
