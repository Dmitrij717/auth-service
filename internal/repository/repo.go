package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dmitrij717/auth-service/internal/apperrors"
	"github.com/Dmitrij717/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func (p PostgresRepo) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	row := p.pool.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at, updated_at;", user.Email, user.PasswordHash)

	var usr domain.User
	err := row.Scan(
		&usr.ID,
		&usr.Email,
		&usr.PasswordHash,
		&usr.CreatedAt,
		&usr.UpdatedAt)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == "23505" {
				return domain.User{}, apperrors.ErrEmailAlreadyExists
			}
		}

		return domain.User{}, fmt.Errorf("create user - %w", err)
	}

	return usr, nil
}

func (p PostgresRepo) GetUserByID(
	ctx context.Context,
	id int64,
) (domain.User, error) {
	row := p.pool.QueryRow(ctx, "SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id=$1", id)

	var user domain.User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, apperrors.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("create user - %w", err)
	}

	return user, nil
}

func (p PostgresRepo) GetUserByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	row := p.pool.QueryRow(ctx, "SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email=$1", email)

	var user domain.User
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, apperrors.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("create user - %w", err)
	}

	return user, nil
}
