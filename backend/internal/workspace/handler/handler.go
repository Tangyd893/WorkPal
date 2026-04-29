package handler

import (
	"errors"
	"net/http"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/workspace/service"
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
	auth.GET("/tasks", h.ListTasks)
	auth.POST("/tasks", h.CreateTask)
	auth.PUT("/tasks/:id/status", h.UpdateTaskStatus)
	auth.POST("/tasks/:id/share", h.ShareTask)
	auth.DELETE("/tasks/:id", h.DeleteTask)
	auth.GET("/schedule", h.ListEvents)
	auth.POST("/schedule", h.CreateEvent)
	auth.POST("/schedule/:id/share", h.ShareEvent)
	auth.DELETE("/schedule/:id", h.DeleteEvent)
}

func (h *Handler) ListTasks(c *gin.Context) {
	tasks, err := h.svc.ListTasks(c.Request.Context(), c.GetInt64("userID"))
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, tasks)
}

func (h *Handler) CreateTask(c *gin.Context) {
	var input service.TaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid task payload")
		return
	}
	task, err := h.svc.CreateTask(c.Request.Context(), c.GetInt64("userID"), input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, task)
}

func (h *Handler) UpdateTaskStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid task status payload")
		return
	}
	task, err := h.svc.UpdateTaskStatus(c.Request.Context(), c.GetInt64("userID"), id, input.Status)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, task)
}

func (h *Handler) ShareTask(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	task, err := h.svc.ShareTask(c.Request.Context(), c.GetInt64("userID"), id)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, task)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteTask(c.Request.Context(), c.GetInt64("userID"), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListEvents(c *gin.Context) {
	events, err := h.svc.ListEvents(c.Request.Context(), c.GetInt64("userID"))
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, events)
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var input service.EventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid schedule payload")
		return
	}
	event, err := h.svc.CreateEvent(c.Request.Context(), c.GetInt64("userID"), input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, event)
}

func (h *Handler) ShareEvent(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	event, err := h.svc.ShareEvent(c.Request.Context(), c.GetInt64("userID"), id)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, event)
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteEvent(c.Request.Context(), c.GetInt64("userID"), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := service.ParseID(c.Param("id"))
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func handleErr(c *gin.Context, err error) {
	if errors.Is(err, service.ErrNotFound) {
		response.FailWithMessage(c, http.StatusNotFound, "item not found")
		return
	}
	response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
}
