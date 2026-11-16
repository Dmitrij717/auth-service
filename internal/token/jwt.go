package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret string
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret: secret,
	}
}

func (j *JWTManager) GenerateToken(email string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", fmt.Errorf("signing failed - %w", err)
	}

	return jwtToken, nil
}

func (j *JWTManager) RefreshToken(token string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			return []byte(j.secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("parse token - %w", err)
	}

	if !parsedToken.Valid {
		return "", fmt.Errorf("invalid parsed token")
	}

	if claims.Subject == "" {
		return "", fmt.Errorf("subject is empty")
	}

	return j.GenerateToken(claims.Subject)
}
