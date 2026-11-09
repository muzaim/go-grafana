package monitor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestMonitorRouter(userID uuid.UUID) (*gin.Engine, Service) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := newMockMonitorRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	v1 := r.Group("/api/v1")
	fakeAuthMiddleware := func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
	handler.RegisterRoutes(v1, fakeAuthMiddleware)

	return r, svc
}

func TestHandler_MonitorCRUD(t *testing.T) {
	userID := uuid.New()
	router, _ := setupTestMonitorRouter(userID)

	// 1. Create Monitor
	createBody, _ := json.Marshal(CreateMonitorRequest{
		Name:        "API Test",
		URL:         "https://api.github.com",
		Protocol:    ProtocolHTTPS,
		IntervalSec: 60,
		TimeoutSec:  10,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/monitors", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	data := createResp["data"].(map[string]interface{})
	monitorID := data["id"].(string)

	// 2. Get Monitor By ID
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/monitors/"+monitorID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. List Monitors
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/monitors", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Toggle Monitor
	toggleBody, _ := json.Marshal(ToggleRequest{IsActive: false})
	req, _ = http.NewRequest(http.MethodPatch, "/api/v1/monitors/"+monitorID+"/toggle", bytes.NewBuffer(toggleBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Delete Monitor
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/monitors/"+monitorID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
