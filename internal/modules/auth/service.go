package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-up-monitoring/internal/config"
	pkgauth "go-up-monitoring/pkg/auth"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidTokenType   = errors.New("invalid token type for refresh")
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*AuthResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error)
}

type service struct {
	repo Repository
	cfg  config.JWTConfig
}

func NewService(repo Repository, cfg config.JWTConfig) Service {
	return &service{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := pkgauth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         "user",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(user)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !pkgauth.CheckPassword(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	return s.generateAuthResponse(user)
}

func (s *service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*AuthResponse, error) {
	claims, err := pkgauth.ValidateToken(req.RefreshToken, s.cfg.Secret)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return s.generateAuthResponse(user)
}

func (s *service) GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := ToUserResponse(user)
	return &resp, nil
}

func (s *service) generateAuthResponse(user *User) (*AuthResponse, error) {
	accessDuration := time.Duration(s.cfg.AccessTokenExpiryMin) * time.Minute
	if accessDuration <= 0 {
		accessDuration = 15 * time.Minute
	}

	refreshDuration := time.Duration(s.cfg.RefreshTokenExpiryDay) * 24 * time.Hour
	if refreshDuration <= 0 {
		refreshDuration = 7 * 24 * time.Hour
	}

	accessToken, err := pkgauth.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
		"access",
		s.cfg.Secret,
		accessDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := pkgauth.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
		"refresh",
		s.cfg.Secret,
		refreshDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessDuration.Seconds()),
		User:         ToUserResponse(user),
	}, nil
}
