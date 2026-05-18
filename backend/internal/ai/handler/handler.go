package handler

import (
	"net/http"

	"github.com/Tangyd893/WorkPal/backend/internal/ai/service"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *service.Service }

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.POST("/ai/search", h.SmartSearch)
	auth.POST("/ai/summarize", h.SummarizeTasks)
}

func (h *Handler) SmartSearch(c *gin.Context) {
	var input struct {
		Query     string `json:"query"`
		ProjectID *int64 `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	result, err := h.svc.SmartSearch(c.Request.Context(), input.Query, input.ProjectID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *Handler) SummarizeTasks(c *gin.Context) {
	var input service.TaskSummaryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	result, err := h.svc.SummarizeTasks(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}
