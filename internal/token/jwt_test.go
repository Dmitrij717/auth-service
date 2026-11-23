package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	secret := "test-secret-key-which-is-long-enough"
	email := "example@gmail.com"

	manager := NewJWTManager(secret)

	now := time.Now()

	token, err := manager.GenerateToken(email)
	if err != nil {
		t.Fatalf("generate failed - %v", err)
	}

	if token == "" {
		t.Fatalf("token is empty")
	}

	claims := &jwt.RegisteredClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("parse token - %v", err)
	}

	if !parsedToken.Valid {
		t.Fatalf("invalid token - %v", err)
	}

	if parsedToken.Method.Alg() != "HS256" {
		t.Fatalf("expected HS256, got %s", parsedToken.Method.Alg())
	}

	if claims.Subject != email {
		t.Fatalf("expected subject %v, got %v", email, claims.Subject)
	}

	if claims.ExpiresAt == nil {
		t.Fatalf("expiration time is nil")
	}

	expectedExpiration := now.Add(60 * time.Minute)
	actualExpiration := claims.ExpiresAt.Time

	if diff := actualExpiration.Sub(expectedExpiration); diff < -time.Second || diff > time.Second {
		t.Fatalf("expected expiration around %v, got %v", expectedExpiration, actualExpiration)
	}

}

func TestJWTManager_RefreshToken(t *testing.T) {
	secret := "test-secret-key-which-is-long-enough"
	email := "example@gmail.com"
	targetMethod := jwt.SigningMethodHS256.Alg()

	manager := NewJWTManager(secret)

	token, err := manager.GenerateToken(email)
	if err != nil {
		t.Fatalf("generation failed - %v", err)
	}

	now := time.Now()

	refreshedToken, err := manager.RefreshToken(token)
	if err != nil {
		t.Fatalf("refresh token failed - %v", err)
	}

	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(refreshedToken, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("parse failed - %v", err)
	}

	if !parsedToken.Valid {
		t.Fatalf("parsed token is invalid")
	}

	if claims.ExpiresAt == nil {
		t.Fatalf("expiration time is nil")
	}

	expectedExpiration := now.Add(60 * time.Minute)
	actualExpiration := claims.ExpiresAt.Time

	if diff := actualExpiration.Sub(expectedExpiration); diff < -time.Second || diff > time.Second {
		t.Fatalf("unexpected expiration: %v", actualExpiration)
	}

	if claims.Subject != email {
		t.Fatalf("expected subject %v, got %v", email, claims.Subject)
	}

	if parsedToken.Method.Alg() != targetMethod {
		t.Fatalf("expected method %v, got %v", targetMethod, parsedToken.Method.Alg())
	}
}

func TestJWTManager_RefreshToken_ExpiredToken(t *testing.T) {
	secret := "test-secret-key-which-is-long-enough"
	email := "example@gmail.com"

	manager := NewJWTManager(secret)

	expiredClaims := jwt.RegisteredClaims{
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)

	tokenString, err := expiredToken.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	_, err = manager.RefreshToken(tokenString)
	if err == nil {
		t.Fatalf("expected error for expired token")
	}
}

func TestJWTManager_RefreshToken_InvalidSignature(t *testing.T) {
	secret := "test-secret-key-which-is-long-enough"
	email := "example@gmail.com"

	manager := NewJWTManager(secret)

	fakeManager := NewJWTManager("another-secret-key")

	fakeToken, err := fakeManager.GenerateToken(email)
	if err != nil {
		t.Fatalf("generate fake token - %v", err)
	}

	_, err = manager.RefreshToken(fakeToken)
	if err == nil {
		t.Fatalf("expected error for invalid signature")
	}
}
