package heartbeat

import (
	"errors"
	"net/http"
	"strconv"

	"go-up-monitoring/internal/modules/monitor"
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
		monitors.GET("/:id/heartbeats", h.GetHeartbeats)
		monitors.GET("/:id/stats", h.GetStats)
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

func (h *Handler) GetHeartbeats(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	monitorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid monitor ID", err)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	heartbeats, total, err := h.svc.GetHeartbeats(c.Request.Context(), monitorID, userID, limit, offset)
	if err != nil {
		if errors.Is(err, monitor.ErrMonitorNotFound) {
			response.NotFound(c, "Monitor not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch heartbeats", err)
		return
	}

	response.Success(c, http.StatusOK, "Heartbeats retrieved successfully", gin.H{
		"heartbeats": heartbeats,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

func (h *Handler) GetStats(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized request")
		return
	}

	monitorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid monitor ID", err)
		return
	}

	period := c.DefaultQuery("period", "24h")
	stats, err := h.svc.GetStats(c.Request.Context(), monitorID, userID, period)
	if err != nil {
		if errors.Is(err, monitor.ErrMonitorNotFound) {
			response.NotFound(c, "Monitor not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Statistics retrieved successfully", stats)
}
