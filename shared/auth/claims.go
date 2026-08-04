package auth

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
