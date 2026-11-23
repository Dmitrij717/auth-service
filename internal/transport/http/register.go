package http_handlers

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"

	"github.com/Dmitrij717/auth-service/internal/apperrors"
	"github.com/Dmitrij717/auth-service/internal/domain"
)

type RegisterService interface {
	Register(ctx context.Context, email, password string) (domain.User, error)
}

type RegisterHandler struct {
	service RegisterService
}

func NewRegisterHandler(service RegisterService) *RegisterHandler {
	return &RegisterHandler{
		service: service,
	}
}

func (h *RegisterHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	email, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="auth"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.service.Register(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrEmailAlreadyExists):
			http.Error(w, "email already exists", http.StatusConflict)
		case errors.Is(err, apperrors.ErrEmailIsEmpty),
			errors.Is(err, apperrors.ErrPasswordIsEmpty),
			errors.Is(err, apperrors.ErrPasswordTooLong):
			http.Error(w, "invalid registration data", http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	res := &ResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	data, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "marshalling failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, "writing to response", http.StatusInternalServerError)
		return
	}
}
