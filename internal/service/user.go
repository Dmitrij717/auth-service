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

type UserService struct {
	store UserStore
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
