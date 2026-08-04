package auth

import "context"

type contextKey string

const claimsContextKey contextKey = "authenticated-user-claims"

func ContextWithClaims(
	ctx context.Context,
	claims *UserClaims,
) context.Context {
	return context.WithValue(
		ctx,
		claimsContextKey,
		claims,
	)
}

func ClaimsFromContext(
	ctx context.Context,
) (*UserClaims, bool) {
	claims, ok := ctx.Value(
		claimsContextKey,
	).(*UserClaims)

	return claims, ok && claims != nil
}

func UserIDFromContext(
	ctx context.Context,
) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims.UserID == "" {
		return "", false
	}

	return claims.UserID, true
}
