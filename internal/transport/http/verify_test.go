package http_handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	http_handlers "github.com/Dmitrij717/auth-service/internal/transport/http"
	"github.com/Dmitrij717/auth-service/internal/transport/http/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVerifyHander_Handle_Success(t *testing.T) {
	mockService := mocks.NewMockVerifyService(t)

	mockService.EXPECT().
		RefreshToken(mock.Anything, "some-token").
		Return("new-token", nil).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/verify", nil)
	req.Header.Set("Authorization", "Bearer some-token")

	recorder := httptest.NewRecorder()

	header := http_handlers.NewVerifyHandler(mockService)

	header.Handle(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "Bearer new-token", recorder.Header().Get("Authorization"))
}

func TestVerifyHandler_Handle_ExpiredToken(t *testing.T) {
	mockService := mocks.NewMockVerifyService(t)

	mockService.EXPECT().
		RefreshToken(mock.Anything, "expired-token").
		Return("", errors.New("expired token")).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/verify", nil)
	req.Header.Set("Authorization", "Bearer expired-token")

	recorder := httptest.NewRecorder()

	handler := http_handlers.NewVerifyHandler(mockService)

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Equal(t, "Bearer", recorder.Header().Get("WWW-Authenticate"))
}

func TestVerifyHandler_Handle_InvalidToken(t *testing.T) {
	mockService := mocks.NewMockVerifyService(t)

	mockService.EXPECT().
		RefreshToken(mock.Anything, "forged-token").
		Return("", errors.New("invalid token")).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/verify", nil)
	req.Header.Set("Authorization", "Bearer forged-token")

	recorder := httptest.NewRecorder()

	handler := http_handlers.NewVerifyHandler(mockService)

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Equal(t, "Bearer", recorder.Header().Get("WWW-Authenticate"))
}
