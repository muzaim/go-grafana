package monitor

import (
	"errors"
	"net/http"
	"strconv"

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
	monitors := rg.Group("/monitors")
	monitors.Use(authMiddleware)
	{
		monitors.POST("", h.Create)
		monitors.GET("", h.List)
		monitors.GET("/:id", h.GetByID)
		monitors.PUT("/:id", h.Update)
		monitors.DELETE("/:id", h.Delete)
		monitors.PATCH("/:id/toggle", h.Toggle)
	}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := val.(uuid.UUID)
	return userID, ok
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	var req CreateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid monitor configuration payload", err)
		return
	}

	res, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidURL) || errors.Is(err, ErrProtocolMismatch) || errors.Is(err, ErrInvalidTimeout) {
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), err)
			return
		}
		response.InternalServerError(c, "Failed to create monitor", err)
		return
	}

	response.Success(c, http.StatusCreated, "Monitor created successfully", res)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	res, err := h.svc.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch monitors", err)
		return
	}

	response.Success(c, http.StatusOK, "Monitors retrieved successfully", res)
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid monitor ID", err)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrMonitorNotFound) {
			response.NotFound(c, "Monitor not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch monitor", err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor retrieved successfully", res)
}

func (h *Handler) Update(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid monitor ID", err)
		return
	}

	var req UpdateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid monitor configuration payload", err)
		return
	}

	res, err := h.svc.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrMonitorNotFound) {
			response.NotFound(c, "Monitor not found")
			return
		}
		if errors.Is(err, ErrInvalidURL) || errors.Is(err, ErrProtocolMismatch) || errors.Is(err, ErrInvalidTimeout) {
			response.Error(c, http.StatusUnprocessableEntity, err.Error(), err)
			return
		}
		response.InternalServerError(c, "Failed to update monitor", err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor updated successfully", res)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid monitor ID", err)
		return
	}

	err = h.svc.Delete(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrMonitorNotFound) {
			response.NotFound(c, "Monitor not found")
			return
		}
		response.InternalServerError(c, "Failed to delete monitor", err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor deleted successfully", nil)
}

type ToggleRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *Handler) Toggle(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid monitor ID", err)
		return
	}

	var req ToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid toggle request payload", err)
		return
	}

	res, err := h.svc.Toggle(c.Request.Context(), id, userID, req.IsActive)
	if err != nil {
		if errors.Is(err, ErrMonitorNotFound) {
			response.NotFound(c, "Monitor not found")
			return
		}
		response.InternalServerError(c, "Failed to toggle monitor", err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor status updated successfully", res)
}
