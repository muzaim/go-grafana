package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-up-monitoring/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() (*gin.Engine, Service) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := newMockRepository()
	cfg := config.JWTConfig{
		Secret:                "test-secret",
		AccessTokenExpiryMin:  15 * time.Minute,
		RefreshTokenExpiryDay: 7 * time.Hour,
	}
	svc := NewService(repo, cfg)
	handler := NewHandler(svc)

	v1 := r.Group("/api/v1")
	fakeAuthMiddleware := func(c *gin.Context) {
		// Mock inject user_id from header or use default for test
		c.Next()
	}
	handler.RegisterRoutes(v1, fakeAuthMiddleware)

	return r, svc
}

func TestHandler_Register_Success(t *testing.T) {
	router, _ := setupTestRouter()

	body, _ := json.Marshal(RegisterRequest{
		Email:    "handler_test@example.com",
		Password: "Password123!",
		FullName: "Handler Test",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Register_ValidationFailure(t *testing.T) {
	router, _ := setupTestRouter()

	// Short password (< 8 chars)
	body, _ := json.Marshal(RegisterRequest{
		Email:    "handler_test@example.com",
		Password: "short",
		FullName: "Handler Test",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Login_Success(t *testing.T) {
	router, svc := setupTestRouter()

	_, _ = svc.Register(t.Context(), RegisterRequest{
		Email:    "login_test@example.com",
		Password: "Password123!",
		FullName: "Login User",
	})

	body, _ := json.Marshal(LoginRequest{
		Email:    "login_test@example.com",
		Password: "Password123!",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
