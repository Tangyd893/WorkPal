package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/notification/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/notifications", h.List)
	auth.GET("/notifications/unread-count", h.UnreadCount)
	auth.PUT("/notifications/:id/read", h.MarkRead)
	auth.PUT("/notifications/read-all", h.MarkAllRead)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), c.GetInt64("userID"))
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, items)
}

func (h *Handler) UnreadCount(c *gin.Context) {
	count, err := h.svc.CountUnread(c.Request.Context(), c.GetInt64("userID"))
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"count": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), c.GetInt64("userID"), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	if err := h.svc.MarkAllRead(c.Request.Context(), c.GetInt64("userID")); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}
