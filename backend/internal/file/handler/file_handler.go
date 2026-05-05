package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/audit"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/file/model"
	"github.com/Tangyd893/WorkPal/backend/internal/file/service"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileSvc *service.FileService
	convSvc conversationAuthorizer
	audit   *audit.Recorder
}

type conversationAuthorizer interface {
	IsMember(ctx context.Context, convID, userID int64) (bool, error)
}

func NewFileHandler(fileSvc *service.FileService, convSvc conversationAuthorizer, recorders ...*audit.Recorder) *FileHandler {
	var recorder *audit.Recorder
	if len(recorders) > 0 {
		recorder = recorders[0]
	}
	return &FileHandler{fileSvc: fileSvc, convSvc: convSvc, audit: recorder}
}

func (h *FileHandler) Upload(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, _ := strconv.ParseInt(c.PostForm("conv_id"), 10, 64)
	if convID > 0 && !h.ensureConversationMember(c, convID, userID) {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "请选择要上传的文件")
		return
	}

	uploadedFile, err := h.fileSvc.Upload(c.Request.Context(), userID, convID, file)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, h.serializeFile(c, uploadedFile))
}

func (h *FileHandler) Download(c *gin.Context) {
	userID := c.GetInt64("userID")
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	meta, err := h.fileSvc.GetByID(c.Request.Context(), fileID)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "文件不存在")
		return
	}
	if meta.UserID != userID {
		if meta.ConvID <= 0 || !h.ensureConversationMember(c, meta.ConvID, userID) {
			return
		}
	}

	reader, file, err := h.fileSvc.Download(c.Request.Context(), fileID)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "文件不存在")
		return
	}
	defer reader.Close()
	h.audit.Record(c.Request.Context(), userID, "下载文件", "file", strconv.FormatInt(fileID, 10), c.ClientIP())

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))
	c.Header("Content-Type", file.ContentType)
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))
	_, _ = io.Copy(c.Writer, reader)
}

func (h *FileHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("userID")
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	file, err := h.fileSvc.GetByID(c.Request.Context(), fileID)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "文件不存在")
		return
	}
	if file.UserID != userID {
		response.FailWithMessage(c, http.StatusForbidden, "无权删除该文件")
		return
	}

	deletedFile, err := h.fileSvc.Delete(c.Request.Context(), fileID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.audit.Record(c.Request.Context(), userID, "删除文件", "file", strconv.FormatInt(fileID, 10), c.ClientIP())
	response.Success(c, h.serializeFile(c, deletedFile))
}

func (h *FileHandler) Share(c *gin.Context) {
	userID := c.GetInt64("userID")
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	file, err := h.fileSvc.GetByID(c.Request.Context(), fileID)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "文件不存在")
		return
	}
	if file.UserID != userID && (file.ConvID <= 0 || !h.ensureConversationMember(c, file.ConvID, userID)) {
		return
	}

	response.Success(c, gin.H{
		"file_id":       file.ID,
		"name":          file.Name,
		"download_path": fmt.Sprintf("/api/v1/files/%d", file.ID),
		"share_text":    fmt.Sprintf("%s (%s)", file.Name, fmt.Sprintf("/api/v1/files/%d", file.ID)),
	})
}

func (h *FileHandler) ListByConv(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}
	if !h.ensureConversationMember(c, convID, userID) {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	files, err := h.fileSvc.ListByConv(c.Request.Context(), convID, offset, pageSize)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	payload := make([]gin.H, 0, len(files))
	for _, file := range files {
		payload = append(payload, h.serializeFile(c, file))
	}
	response.Success(c, payload)
}

func (h *FileHandler) ListByUser(c *gin.Context) {
	userID := c.GetInt64("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	files, err := h.fileSvc.ListByUser(c.Request.Context(), userID, offset, pageSize)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	payload := make([]gin.H, 0, len(files))
	for _, file := range files {
		payload = append(payload, h.serializeFile(c, file))
	}
	response.Success(c, payload)
}

func (h *FileHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.POST("/files/upload", h.Upload)
	auth.GET("/files", h.ListByUser)
	auth.GET("/files/:id", h.Download)
	auth.DELETE("/files/:id", h.Delete)
	auth.POST("/files/:id/share", h.Share)
	auth.GET("/conversations/:id/files", h.ListByConv)
}

func (h *FileHandler) ensureConversationMember(c *gin.Context, convID, userID int64) bool {
	if h.convSvc == nil {
		response.FailWithMessage(c, http.StatusForbidden, "无权访问该会话资源")
		return false
	}
	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return false
	}
	if !isMember {
		response.FailWithMessage(c, http.StatusForbidden, "无权访问该会话资源")
		return false
	}
	return true
}

func (h *FileHandler) serializeFile(c *gin.Context, file *model.File) gin.H {
	return gin.H{
		"id":            file.ID,
		"user_id":       file.UserID,
		"conv_id":       file.ConvID,
		"name":          file.Name,
		"size":          file.Size,
		"content_type":  file.ContentType,
		"mime_type":     file.MimeType,
		"created_at":    file.CreatedAt,
		"download_path": fmt.Sprintf("/api/v1/files/%d", file.ID),
		"share_path":    fmt.Sprintf("/api/v1/files/%d/share", file.ID),
		"download_url":  fmt.Sprintf("%s://%s/api/v1/files/%d", requestScheme(c), c.Request.Host, file.ID),
	}
}

func requestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		return forwarded
	}
	return "http"
}
