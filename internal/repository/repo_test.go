package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Dmitrij717/auth-service/internal/apperrors"
	"github.com/Dmitrij717/auth-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *PostgresRepo {
	t.Helper()

	ctx := context.Background()
	ptgrs, err := testcontainers.Run(ctx,
		"postgres:18.6-bookworm",
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "auth_test",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("error when run container - %v", err)
	}

	testcontainers.CleanupContainer(t, ptgrs)

	host, err := ptgrs.Host(ctx)
	if err != nil {
		t.Fatalf("error of getting host string - %v", err)
	}

	port, err := ptgrs.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("error of getting port - %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://test:test@%s:%s/auth_test?sslmode=disable",
		host,
		port.Port(),
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("can't to create new pool - %v", err)
	}

	t.Cleanup(pool.Close)

	err = pool.Ping(ctx)
	if err != nil {
		t.Fatalf("database is failed - %v", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("failed to close db - %v", err)
		}
	})

	err = goose.SetDialect("postgres")
	if err != nil {
		t.Fatalf("set goose dialect - %v", err)
	}

	migrationDir := "../../migrations"

	err = goose.UpContext(ctx, sqlDB, migrationDir)
	if err != nil {
		t.Fatalf("apply migrations - %v", err)
	}

	return NewPostgresRepo(pool)
}

func TestPostgresConnection(t *testing.T) {
	repo := setupTestDB(t)
	if repo == nil {
		t.Fatalf("expected repository")
	}
}

func TestPostgresRepo_CreateUser(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	tests := []struct {
		name                string
		email, passwordHash string
		wantErr             error
		prepare             func(t *testing.T, email, passwordHash string)
	}{
		{
			name:         "success user create",
			email:        "Dmitrij@gmail.com",
			passwordHash: "w1rdm124idqwQ",
			wantErr:      nil,
			prepare:      nil,
		},

		{
			name:         "duplicate email user",
			email:        "duplicate@gmail.com",
			passwordHash: "qwori2nfas-24",
			wantErr:      apperrors.ErrEmailAlreadyExists,
			prepare: func(t *testing.T, email, passwordHash string) {

				_, err := repo.CreateUser(ctx, domain.User{
					Email:        email,
					PasswordHash: passwordHash,
				})

				if err != nil {
					t.Fatalf("prepare user: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(t, tt.email, tt.passwordHash)
			}

			user := domain.User{
				Email:        tt.email,
				PasswordHash: tt.passwordHash,
			}

			user, err := repo.CreateUser(ctx, user)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("test failed - %v", err)
			}

			if err != nil {
				return
			}

			if user.ID <= 0 ||
				user.Email != tt.email ||
				user.PasswordHash != tt.passwordHash ||
				user.CreatedAt.IsZero() ||
				user.UpdatedAt.IsZero() {
				t.Fatalf("data mismatch")
			}

		})
	}

}

func TestPostgresRepo_GetUserByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := domain.User{
			Email:        "Dmitrij@gmail.com",
			PasswordHash: "Wr1k2mt0q240ok",
		}

		created, err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("created user - %v", err)
		}

		actual, err := repo.GetUserByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("get user - %v", err)
		}

		if actual.ID != created.ID ||
			actual.Email != created.Email ||
			actual.PasswordHash != created.PasswordHash ||
			!actual.CreatedAt.Equal(created.CreatedAt) ||
			!actual.UpdatedAt.Equal(created.UpdatedAt) {
			t.Fatalf("users don't match:\nexpected: %v\nactual: %v", created, actual)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, 12471269)

		if !errors.Is(err, apperrors.ErrUserNotFound) {
			t.Fatalf("expected %v, got %v", apperrors.ErrUserNotFound, err)
		}
	})
}

func TestPostgresRepo_GetUserByEmail(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := domain.User{
			Email:        "Dmitrij@gmail.com",
			PasswordHash: "214otgmeqw2Wq",
		}

		created, err := repo.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("create user - %v", err)
		}

		actual, err := repo.GetUserByEmail(ctx, created.Email)
		if err != nil {
			t.Fatalf("get user by email - %v", err)
		}

		if actual.ID != created.ID ||
			actual.Email != created.Email ||
			actual.PasswordHash != created.PasswordHash ||
			!actual.CreatedAt.Equal(created.CreatedAt) ||
			!actual.UpdatedAt.Equal(created.UpdatedAt) {
			t.Fatalf("users don't match\nexpected: %v\nactual: %v", created, actual)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := repo.GetUserByEmail(ctx, "user@mail.ru")
		if !errors.Is(err, apperrors.ErrUserNotFound) {
			t.Fatalf("expected %v, got %v", apperrors.ErrUserNotFound, err)
		}
	})
}
