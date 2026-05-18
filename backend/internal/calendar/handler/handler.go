package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/calendar/service"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *service.Service }
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/calendar", h.ListEvents)
	auth.POST("/calendar", h.CreateEvent)
	auth.GET("/calendar/:id", h.GetEvent)
	auth.PUT("/calendar/:id", h.UpdateEvent)
	auth.DELETE("/calendar/:id", h.DeleteEvent)
}

func (h *Handler) ListEvents(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	var organizerID *int64
	if oid := c.Query("organizer_id"); oid != "" {
		id, _ := strconv.ParseInt(oid, 10, 64)
		organizerID = &id
	}
	var f, t *string
	if from != "" { f = &from }
	if to != "" { t = &to }
	events, err := h.svc.ListEvents(c.Request.Context(), f, t, organizerID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, events)
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var input service.EventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	ev, err := h.svc.CreateEvent(c.Request.Context(), c.GetInt64("userID"), input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, ev)
}

func (h *Handler) GetEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ev, err := h.svc.GetEvent(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "event not found")
		return
	}
	response.Success(c, ev)
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input service.EventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	ev, err := h.svc.UpdateEvent(c.Request.Context(), id, input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, ev)
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteEvent(c.Request.Context(), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}
