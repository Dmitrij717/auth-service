package service

import (
	"context"

	"github.com/Dmitrij717/auth-service/internal/domain"
)

type UserStore interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByID(ctx context.Context, id int64) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}
