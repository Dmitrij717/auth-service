package http_handlers

import (
	"context"
	"net/http"
	"strings"
)

type VerifyService interface {
	RefreshToken(ctx context.Context, token string) (string, error)
}

type VerifyHandler struct {
	service VerifyService
}

func NewVerifyHandler(service VerifyService) *VerifyHandler {
	return &VerifyHandler{
		service: service,
	}
}

func (h *VerifyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")

	slicedHeader := strings.Fields(authHeader)
	if len(slicedHeader) != 2 || !strings.EqualFold(slicedHeader[0], "Bearer") {
		w.Header().Set("WWW-Authenticate", `Bearer`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshToken, err := h.service.RefreshToken(r.Context(), slicedHeader[1])
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer`)
		http.Error(w, "user unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Authorization", "Bearer "+refreshToken)
	w.WriteHeader(http.StatusOK)
}
