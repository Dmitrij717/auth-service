package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Dmitrij717/auth-service/internal/apperrors"
	"github.com/Dmitrij717/auth-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByID(ctx context.Context, id int64) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}

type TokenManager interface {
	GenerateToken(email string) (string, error)
	RefreshToken(token string) (string, error)
}

type UserService struct {
	store        UserStore
	tokenManager TokenManager
}

func NewUserService(
	store UserStore,
	tokenManager TokenManager,
) *UserService {
	return &UserService{
		store:        store,
		tokenManager: tokenManager,
	}
}

func (u *UserService) Register(ctx context.Context, email, password string) (domain.User, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if email == "" {
		return domain.User{}, apperrors.ErrEmailIsEmpty
	}

	if password == "" {
		return domain.User{}, apperrors.ErrPasswordIsEmpty
	}

	passwordSlice := []byte(password)

	hashedPassword, err := bcrypt.GenerateFromPassword(passwordSlice, bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return domain.User{}, apperrors.ErrPasswordTooLong
		}

		return domain.User{}, fmt.Errorf("error while hashing password - %w", err)
	}

	user := domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	created, err := u.store.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user - %w", err)
	}

	return created, nil
}

func (u *UserService) ValidateCredentials(
	ctx context.Context,
	email, password string,
) (bool, error) {
	user, err := u.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, fmt.Errorf("compare password hash: %w", err)
	}

	return true, nil
}

func (u *UserService) GenerateToken(ctx context.Context, email string) (string, error) {
	token, err := u.tokenManager.GenerateToken(email)
	if err != nil {
		return "", fmt.Errorf("generation failed - %w", err)
	}

	return token, nil
}

func (u *UserService) RefreshToken(ctx context.Context, token string) (string, error) {
	refreshToken, err := u.tokenManager.RefreshToken(token)
	if err != nil {
		return "", fmt.Errorf("refresh failed - %w", err)
	}

	return refreshToken, nil
}
