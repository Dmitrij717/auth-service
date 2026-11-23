package grpc_server

import (
	"context"
	"errors"

	"github.com/Dmitrij717/auth-service/internal/apperrors"
	"github.com/Dmitrij717/auth-service/internal/domain"
	"github.com/Dmitrij717/auth-service/internal/transport/grpc/gen/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (domain.User, error)
	ValidateCredentials(ctx context.Context, email, password string) (bool, error)
	GenerateToken(ctx context.Context, email string) (string, error)
	RefreshToken(ctx context.Context, token string) (string, error)
}

type Server struct {
	auth.UnimplementedAuthServiceServer
	service AuthService
}

func NewServer(service AuthService) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) Register(
	ctx context.Context,
	req *auth.RegisterRequest,
) (*auth.RegisterResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	user, err := s.service.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrEmailAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		case errors.Is(err, apperrors.ErrEmailIsEmpty):
			return nil, status.Error(codes.InvalidArgument, "email is empty")

		case errors.Is(err, apperrors.ErrPasswordIsEmpty):
			return nil, status.Error(codes.InvalidArgument, "password is empty")

		case errors.Is(err, apperrors.ErrPasswordTooLong):
			return nil, status.Error(codes.InvalidArgument, "password too long")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	userRes := auth.User{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return &auth.RegisterResponse{
		User: &userRes,
	}, status.Error(codes.OK, "ok")
}

func (s *Server) Login(
	ctx context.Context,
	req *auth.LoginRequest,
) (*auth.LoginResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	isValid, err := s.service.ValidateCredentials(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if !isValid {
		return nil, status.Error(codes.Unauthenticated, "invalid email/password") // не знаю какую лучше тут использовать ошибку
	}

	token, err := s.service.GenerateToken(ctx, req.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	resp := &auth.LoginResponse{
		Jwt: token,
	}

	return resp, nil
}

func (s *Server) RefreshToken(
	ctx context.Context,
	req *auth.RefreshTokenRequest,
) (*auth.RefreshTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	refreshedToken, err := s.service.RefreshToken(ctx, req.GetJwt())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	resp := &auth.RefreshTokenResponse{
		Jwt: refreshedToken,
	}

	return resp, nil
}
