package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/audit"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/project/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc   *service.Service
	audit *audit.Recorder
}

func NewHandler(svc *service.Service, recorders ...*audit.Recorder) *Handler {
	var recorder *audit.Recorder
	if len(recorders) > 0 {
		recorder = recorders[0]
	}
	return &Handler{svc: svc, audit: recorder}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/projects", h.ListProjects)
	auth.POST("/projects", h.CreateProject)
	auth.GET("/projects/:id", h.GetProject)
	auth.DELETE("/projects/:id", h.DeleteProject)
	auth.GET("/projects/:id/issues", h.ListIssues)
	auth.POST("/projects/:id/issues", h.CreateIssue)
	auth.GET("/projects/:id/issue-types", h.ListIssueTypes)
	auth.GET("/issues/:id", h.GetIssue)
	auth.PUT("/issues/:id", h.UpdateIssue)
	auth.PUT("/issues/:id/status", h.UpdateIssueStatus)
	auth.DELETE("/issues/:id", h.DeleteIssue)
	auth.GET("/issues/:id/changelogs", h.ListChangelogs)
	auth.POST("/issues/:id/associations", h.CreateAssociation)
	auth.GET("/issues/:id/associations", h.ListAssociations)
}

func (h *Handler) ListProjects(c *gin.Context) {
	projects, err := h.svc.ListProjects(c.Request.Context())
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, projects)
}

func (h *Handler) CreateProject(c *gin.Context) {
	var input service.ProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid project payload")
		return
	}
	project, err := h.svc.CreateProject(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, project)
}

func (h *Handler) GetProject(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	project, err := h.svc.GetProject(c.Request.Context(), id)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, project)
}

func (h *Handler) DeleteProject(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteProject(c.Request.Context(), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.audit.Record(c.Request.Context(), c.GetInt64("userID"), "删除项目", "project", strconv.FormatInt(id, 10), c.ClientIP())
	response.Success(c, nil)
}

func (h *Handler) ListIssues(c *gin.Context) {
	projectID, ok := parseIDParam(c)
	if !ok {
		return
	}
	issues, err := h.svc.ListIssues(c.Request.Context(), projectID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, issues)
}

func (h *Handler) CreateIssue(c *gin.Context) {
	projectID, ok := parseIDParam(c)
	if !ok {
		return
	}
	var input service.IssueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid issue payload")
		return
	}
	input.ProjectID = projectID
	if input.ReporterID == 0 {
		input.ReporterID = c.GetInt64("userID")
	}
	issue, err := h.svc.CreateIssue(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, issue)
}

func (h *Handler) GetIssue(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	issue, err := h.svc.GetIssue(c.Request.Context(), id)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, issue)
}

func (h *Handler) UpdateIssue(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var input service.IssueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid issue payload")
		return
	}
	issue, err := h.svc.UpdateIssue(c.Request.Context(), id, c.GetInt64("userID"), input)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, issue)
}

func (h *Handler) UpdateIssueStatus(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var input service.IssueStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid status payload")
		return
	}
	issue, err := h.svc.UpdateIssueStatus(c.Request.Context(), id, c.GetInt64("userID"), input.Status)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, issue)
}

func (h *Handler) DeleteIssue(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteIssue(c.Request.Context(), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.audit.Record(c.Request.Context(), c.GetInt64("userID"), "删除事项", "issue", strconv.FormatInt(id, 10), c.ClientIP())
	response.Success(c, nil)
}

func (h *Handler) ListChangelogs(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	logs, err := h.svc.ListChangelogs(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, logs)
}

func (h *Handler) CreateAssociation(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var input struct {
		TargetType string `json:"target_type"`
		TargetID   int64  `json:"target_id"`
		LinkType   string `json:"link_type"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid association payload")
		return
	}
	if err := h.svc.CreateAssociation(c.Request.Context(), "issue", id, input.TargetType, input.TargetID, input.LinkType, c.GetInt64("userID")); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListAssociations(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	assocs, err := h.svc.ListAssociations(c.Request.Context(), "issue", id)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, assocs)
}

func (h *Handler) ListIssueTypes(c *gin.Context) {
	projectID, ok := parseIDParam(c)
	if !ok {
		return
	}
	types, err := h.svc.ListIssueTypes(c.Request.Context(), projectID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, types)
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
