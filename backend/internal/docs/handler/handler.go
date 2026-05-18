package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/docs/service"
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
	auth.GET("/documents", h.ListDocuments)
	auth.POST("/documents", h.CreateDocument)
	auth.GET("/documents/:id", h.GetDocument)
	auth.PUT("/documents/:id", h.UpdateDocument)
	auth.DELETE("/documents/:id", h.DeleteDocument)
	auth.GET("/documents/:id/revisions", h.ListRevisions)
}

func (h *Handler) ListDocuments(c *gin.Context) {
	var projectID *int64
	if pid := c.Query("project_id"); pid != "" {
		id, err := strconv.ParseInt(pid, 10, 64)
		if err != nil {
			response.FailWithMessage(c, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &id
	}
	var parentID *int64
	if pid := c.Query("parent_id"); pid != "" {
		id, err := strconv.ParseInt(pid, 10, 64)
		if err != nil {
			response.FailWithMessage(c, http.StatusBadRequest, "invalid parent_id")
			return
		}
		parentID = &id
	}
	docs, err := h.svc.ListDocuments(c.Request.Context(), projectID, parentID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, docs)
}

func (h *Handler) CreateDocument(c *gin.Context) {
	var input service.DocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid document payload")
		return
	}
	doc, err := h.svc.CreateDocument(c.Request.Context(), c.GetInt64("userID"), input)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, doc)
}

func (h *Handler) GetDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid id")
		return
	}
	doc, err := h.svc.GetDocument(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.FailWithMessage(c, http.StatusNotFound, "document not found")
			return
		}
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, doc)
}

func (h *Handler) UpdateDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid id")
		return
	}
	var input service.DocumentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid document payload")
		return
	}
	doc, err := h.svc.UpdateDocument(c.Request.Context(), id, c.GetInt64("userID"), input)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.FailWithMessage(c, http.StatusNotFound, "document not found")
			return
		}
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, doc)
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteDocument(c.Request.Context(), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListRevisions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid id")
		return
	}
	revs, err := h.svc.ListRevisions(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, revs)
}
