package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgauth "go-up-monitoring/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-middleware-secret-key"

	r := gin.New()
	r.Use(AuthMiddleware(secret))
	r.GET("/protected", func(c *gin.Context) {
		userID := c.MustGet(ContextUserID).(uuid.UUID)
		c.JSON(http.StatusOK, gin.H{"user_id": userID.String()})
	})

	// 1. Missing header -> 401
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 2. Malformed header -> 401
	req, _ = http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormatToken")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 3. Expired token -> 401
	expiredToken, _ := pkgauth.GenerateToken(uuid.New(), "a@b.com", "user", "access", secret, -1*time.Minute)
	req, _ = http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 4. Refresh token instead of access token -> 401
	refreshToken, _ := pkgauth.GenerateToken(uuid.New(), "a@b.com", "user", "refresh", secret, 1*time.Hour)
	req, _ = http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 5. Valid access token -> 200
	userID := uuid.New()
	validToken, _ := pkgauth.GenerateToken(userID, "a@b.com", "user", "access", secret, 10*time.Minute)
	req, _ = http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
