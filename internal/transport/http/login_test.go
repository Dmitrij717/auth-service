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

func TestLoginHandler_Handle_Success(t *testing.T) {
	mockService := mocks.NewMockLoginService(t)
	mockService.EXPECT().
		ValidateCredentials(mock.Anything, "test@example.com", "password123").
		Return(true, nil).
		Once()

	mockService.EXPECT().
		GenerateToken(mock.Anything, "test@example.com").
		Return("test-token", nil).
		Once()

	handler := http_handlers.NewLoginHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.SetBasicAuth("test@example.com", "password123")
	recorder := httptest.NewRecorder()
	handler.Handle(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "Bearer test-token", recorder.Header().Get("Authorization"))
}

func TestLoginHandler_Handle_InvalidCredentials(t *testing.T) {
	mockService := mocks.NewMockLoginService(t)
	mockService.EXPECT().
		ValidateCredentials(mock.Anything, "test@example.com", "password123").
		Return(false, nil).
		Once()

	handler := http_handlers.NewLoginHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.SetBasicAuth("test@example.com", "password123")
	recorder := httptest.NewRecorder()

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Equal(t, `Basic realm="auth"`, recorder.Header().Get("WWW-Authenticate"))
}

func TestLoginHandler_Handle_MissingAuthorization(t *testing.T) {
	mockService := mocks.NewMockLoginService(t)

	handler := http_handlers.NewLoginHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)

	recorder := httptest.NewRecorder()

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Equal(t, `Basic realm="auth"`, recorder.Header().Get("WWW-Authenticate"))
}

func TestLoginHandler_Handle_InvalidAuthorizationScheme(t *testing.T) {
	mockService := mocks.NewMockLoginService(t)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Header.Set("Authorization", "Bearer some-token")

	recorder := httptest.NewRecorder()

	handler := http_handlers.NewLoginHandler(mockService)

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Equal(t, `Basic realm="auth"`, recorder.Header().Get("WWW-Authenticate"))
}

func TestLoginHandler_Handle_ValidateCredentialsError(t *testing.T) {
	mockService := mocks.NewMockLoginService(t)

	mockService.EXPECT().
		ValidateCredentials(mock.Anything, "test@example.com", "password123").
		Return(false, errors.New("database connection failed")).
		Once()

	handler := http_handlers.NewLoginHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.SetBasicAuth("test@example.com", "password123")

	recorder := httptest.NewRecorder()

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestLoginHandler_Handle_GenerateTokenError(t *testing.T) {
	mockService := mocks.NewMockLoginService(t)

	mockService.EXPECT().
		ValidateCredentials(mock.Anything, "test@example.com", "password123").
		Return(true, nil).
		Once()

	mockService.EXPECT().
		GenerateToken(mock.Anything, "test@example.com").
		Return("", errors.New("token generation failed")).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.SetBasicAuth("test@example.com", "password123")

	handler := http_handlers.NewLoginHandler(mockService)
	recorder := httptest.NewRecorder()

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestLoginHandler_Handle_MethodNotAllowed(t *testing.T) {
	mockService := mocks.NewMockLoginService(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)

	recorder := httptest.NewRecorder()

	handler := http_handlers.NewLoginHandler(mockService)

	handler.Handle(recorder, req)

	require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	require.Equal(t, http.MethodPost, recorder.Header().Get("Allow"))
}
