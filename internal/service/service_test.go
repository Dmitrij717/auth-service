package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Dmitrij717/auth-service/internal/apperrors"
	"github.com/Dmitrij717/auth-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type MockUserStore struct {
	CreatedUser domain.User
	CreatedErr  error
	CreateCalls int
}

func (m *MockUserStore) CreateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	m.CreateCalls++
	m.CreatedUser = user

	if m.CreatedErr != nil {
		return domain.User{}, m.CreatedErr
	}

	user.ID = 1
	return user, nil
}

func (m *MockUserStore) GetUserByID(
	ctx context.Context,
	id int64,
) (domain.User, error) {
	return domain.User{}, nil
}

func (m *MockUserStore) GetUserByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	return domain.User{}, nil
}

func TestService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		email           string
		password        string
		wantErr         error
		storeErr        error
		wantCreateCalls int
	}{
		{
			name:            "success",
			email:           "Dmitrijc@gmail.com",
			password:        "qwrptk34gnd",
			wantErr:         nil,
			wantCreateCalls: 1,
		},
		{
			name:            "empty email",
			email:           "",
			password:        "",
			wantErr:         apperrors.ErrEmailIsEmpty,
			wantCreateCalls: 0,
		},
		{
			name:            "email only whitespaces",
			email:           "            ",
			password:        "",
			wantErr:         apperrors.ErrEmailIsEmpty,
			wantCreateCalls: 0,
		},
		{
			name:            "empty password",
			email:           "dmitrij@mail.ru",
			password:        "",
			wantErr:         apperrors.ErrPasswordIsEmpty,
			wantCreateCalls: 0,
		},
		{
			name:            "password too long",
			email:           "Dmitrij@gmail.com",
			password:        strings.Repeat("q", 73),
			wantErr:         apperrors.ErrPasswordTooLong,
			wantCreateCalls: 0,
		},
		{
			name:            "email already exists",
			email:           "Dmitrij@gmail.com",
			password:        "qwrthth2342",
			wantErr:         apperrors.ErrEmailAlreadyExists,
			storeErr:        apperrors.ErrEmailAlreadyExists,
			wantCreateCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockUserStore{
				CreatedErr: tt.storeErr,
			}
			service := UserService{
				store: store,
			}

			actual, err := service.Register(ctx, tt.email, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}

			if store.CreateCalls != tt.wantCreateCalls {
				t.Fatalf(
					"expected CreateUser calls %v times, got %v",
					tt.wantCreateCalls,
					store.CreateCalls,
				)
			}

			if tt.wantErr != nil {
				return
			}

			if actual.ID != 1 {
				t.Fatalf("expected ID 1, got %v", actual.ID)
			}

			tt.email = strings.TrimSpace(tt.email)
			tt.email = strings.ToLower(tt.email)

			if actual.Email != tt.email {
				t.Fatalf("mismatch email")
			}

			if actual.PasswordHash == tt.password {
				t.Fatalf("password not hashed")
			}

			if store.CreatedUser.Email != tt.email {
				t.Fatalf("repository received unexpected email")
			}

			if store.CreatedUser.PasswordHash == tt.password {
				t.Fatalf("repository received unhashed password")
			}

			if err := bcrypt.CompareHashAndPassword([]byte(actual.PasswordHash), []byte(tt.password)); err != nil {
				t.Fatalf("mismatch password")
			}
		})
	}
}
