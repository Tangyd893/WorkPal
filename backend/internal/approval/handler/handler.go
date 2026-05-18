package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/approval/service"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *service.Service }
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/approvals/templates", h.ListTemplates)
	auth.POST("/approvals/templates", h.CreateTemplate)
	auth.GET("/approvals/instances", h.ListInstances)
	auth.POST("/approvals/instances", h.CreateInstance)
	auth.GET("/approvals/instances/:id", h.GetInstance)
	auth.POST("/approvals/instances/:id/action", h.ProcessAction)
}

func (h *Handler) ListTemplates(c *gin.Context) {
	var pid *int64
	if p := c.Query("project_id"); p != "" { id, _ := strconv.ParseInt(p, 10, 64); pid = &id }
	ts, err := h.svc.ListTemplates(c.Request.Context(), pid)
	if err != nil { response.FailWithMessage(c, http.StatusInternalServerError, err.Error()); return }
	response.Success(c, ts)
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var input service.TemplateInput
	if err := c.ShouldBindJSON(&input); err != nil { response.FailWithMessage(c, http.StatusBadRequest, "invalid payload"); return }
	t, err := h.svc.CreateTemplate(c.Request.Context(), input)
	if err != nil { response.FailWithMessage(c, http.StatusInternalServerError, err.Error()); return }
	response.Success(c, t)
}

func (h *Handler) ListInstances(c *gin.Context) {
	var sid *int64
	if s := c.Query("submitter_id"); s != "" { id, _ := strconv.ParseInt(s, 10, 64); sid = &id }
	var status *string
	if st := c.Query("status"); st != "" { status = &st }
	insts, err := h.svc.ListInstances(c.Request.Context(), sid, status)
	if err != nil { response.FailWithMessage(c, http.StatusInternalServerError, err.Error()); return }
	response.Success(c, insts)
}

func (h *Handler) CreateInstance(c *gin.Context) {
	var input service.InstanceInput
	if err := c.ShouldBindJSON(&input); err != nil { response.FailWithMessage(c, http.StatusBadRequest, "invalid payload"); return }
	inst, err := h.svc.CreateInstance(c.Request.Context(), c.GetInt64("userID"), input)
	if err != nil { response.FailWithMessage(c, http.StatusInternalServerError, err.Error()); return }
	response.Success(c, inst)
}

func (h *Handler) GetInstance(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	inst, err := h.svc.GetInstance(c.Request.Context(), id)
	if err != nil { response.FailWithMessage(c, http.StatusNotFound, "instance not found"); return }
	response.Success(c, inst)
}

func (h *Handler) ProcessAction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input service.ActionInput
	if err := c.ShouldBindJSON(&input); err != nil { response.FailWithMessage(c, http.StatusBadRequest, "invalid payload"); return }
	inst, err := h.svc.ProcessAction(c.Request.Context(), id, c.GetInt64("userID"), input)
	if err != nil { response.FailWithMessage(c, http.StatusInternalServerError, err.Error()); return }
	response.Success(c, inst)
}
