package auth

import (
	"context"
	"testing"
	"time"

	"go-up-monitoring/internal/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	users map[string]*User
	byID  map[uuid.UUID]*User
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users: make(map[string]*User),
		byID:  make(map[uuid.UUID]*User),
	}
}

func (m *mockRepository) Create(ctx context.Context, user *User) error {
	if _, exists := m.users[user.Email]; exists {
		return ErrUserAlreadyExists
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.Email] = user
	m.byID[user.ID] = user
	return nil
}

func (m *mockRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, exists := m.byID[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockRepository()
	cfg := config.JWTConfig{
		Secret:                "super-jwt-secret-12345",
		AccessTokenExpiryMin:  15 * time.Minute,
		RefreshTokenExpiryDay: 7 * time.Hour,
	}

	svc := NewService(repo, cfg)
	ctx := context.Background()

	// 1. Test Register
	regReq := RegisterRequest{
		Email:    "test@example.com",
		Password: "Password123!",
		FullName: "Test User",
	}

	authResp, err := svc.Register(ctx, regReq)
	assert.NoError(t, err)
	assert.NotNil(t, authResp)
	assert.NotEmpty(t, authResp.AccessToken)
	assert.NotEmpty(t, authResp.RefreshToken)
	assert.Equal(t, "test@example.com", authResp.User.Email)

	// 2. Duplicate Register
	_, err = svc.Register(ctx, regReq)
	assert.ErrorIs(t, err, ErrUserAlreadyExists)

	// 3. Test Login Success
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}
	loginResp, err := svc.Login(ctx, loginReq)
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)

	// 4. Test Login Wrong Password
	badLogin := LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword!",
	}
	_, err = svc.Login(ctx, badLogin)
	assert.ErrorIs(t, err, ErrInvalidCredentials)

	// 5. Test Refresh Token
	refreshResp, err := svc.RefreshToken(ctx, RefreshTokenRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	assert.NoError(t, err)
	assert.NotNil(t, refreshResp)
	assert.NotEmpty(t, refreshResp.AccessToken)

	// 6. Test Refresh Token with Access Token (Should fail)
	_, err = svc.RefreshToken(ctx, RefreshTokenRequest{
		RefreshToken: loginResp.AccessToken,
	})
	assert.ErrorIs(t, err, ErrInvalidTokenType)

	// 7. Test Get Profile
	profile, err := svc.GetProfile(ctx, authResp.User.ID)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", profile.Email)
}
