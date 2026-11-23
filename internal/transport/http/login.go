package http_handlers

import (
	"context"
	"net/http"
)

type LoginService interface {
	ValidateCredentials(ctx context.Context, email, pasword string) (bool, error)
	GenerateToken(ctx context.Context, email string) (string, error)
}

type LoginHandler struct {
	service LoginService
}

func NewLoginHandler(service LoginService) *LoginHandler {
	return &LoginHandler{
		service: service,
	}
}

func (h *LoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="auth"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	isValid, err := h.service.ValidateCredentials(r.Context(), username, password)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !isValid {
		w.Header().Set("WWW-Authenticate", `Basic realm="auth"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := h.service.GenerateToken(r.Context(), username)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
