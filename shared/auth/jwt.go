package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey string
	issuer    string
	ttl       time.Duration
}

func NewJWTManager(
	secretKey string,
	issuer string,
	ttl time.Duration,
) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		issuer:    issuer,
		ttl:       ttl,
	}
}

func (m *JWTManager) GenerateToken(
	userID string,
	email string,
) (string, error) {
	now := time.Now()

	claims := UserClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(
		[]byte(m.secretKey),
	)
	if err != nil {
		return "", fmt.Errorf(
			"failed to sign JWT token: %w",
			err,
		)
	}

	return signedToken, nil
}

func (m *JWTManager) ValidateToken(
	rawToken string,
) (*UserClaims, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf(
					"unexpected signing method: %s",
					token.Method.Alg(),
				)
			}

			return []byte(m.secretKey), nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to validate JWT token: %w",
			err,
		)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	if claims.UserID == "" {
		return nil, fmt.Errorf(
			"JWT token does not contain user_id",
		)
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf(
			"JWT token does not contain subject",
		)
	}

	if claims.Subject != claims.UserID {
		return nil, fmt.Errorf(
			"JWT subject does not match user_id",
		)
	}

	return claims, nil
}
