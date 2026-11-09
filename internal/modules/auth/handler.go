package auth

import (
	"errors"
	"net/http"

	"go-up-monitoring/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.RefreshToken)
		authGroup.GET("/me", authMiddleware, h.GetProfile)
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid registration payload", err)
		return
	}

	authResp, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			response.Error(c, http.StatusConflict, "Email already registered", err)
			return
		}
		response.InternalServerError(c, "Failed to register user", err)
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", authResp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid login payload", err)
		return
	}

	authResp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		response.InternalServerError(c, "Failed to login", err)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", authResp)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid refresh payload", err)
		return
	}

	authResp, err := h.svc.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", authResp)
}

func (h *Handler) GetProfile(c *gin.Context) {
	rawUserID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	userID, ok := rawUserID.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	profile, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "User profile not found")
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", profile)
}
