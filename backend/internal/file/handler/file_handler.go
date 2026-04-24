package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/file/service"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileSvc *service.FileService
}

func NewFileHandler(fileSvc *service.FileService) *FileHandler {
	return &FileHandler{fileSvc: fileSvc}
}

// Upload 上传文件
// POST /api/v1/files/upload
func (h *FileHandler) Upload(c *gin.Context) {
	userID := c.GetInt64("userID")

	// 支持表单参数指定会话 ID
	convID, _ := strconv.ParseInt(c.PostForm("conv_id"), 10, 64)

	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "请选择要上传的文件")
		return
	}

	f, err := h.fileSvc.Upload(c.Request.Context(), userID, convID, file)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":          f.ID,
		"name":        f.Name,
		"size":        f.Size,
		"content_type": f.ContentType,
		"created_at":  f.CreatedAt,
	})
}

// Download 下载文件
// GET /api/v1/files/:id
func (h *FileHandler) Download(c *gin.Context) {
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的文件ID")
		return
	}

	rc, f, err := h.fileSvc.Download(c.Request.Context(), fileID)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "文件不存在")
		return
	}
	defer rc.Close()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, f.Name))
	c.Header("Content-Type", f.ContentType)
	c.Header("Content-Length", strconv.FormatInt(f.Size, 10))
	io.Copy(c.Writer, rc)
}

// ListByConv 获取会话文件列表
// GET /api/v1/conversations/:id/files
func (h *FileHandler) ListByConv(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话ID")
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

	response.Success(c, files)
}

// ListByUser 获取用户文件列表
// GET /api/v1/files
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

	response.Success(c, files)
}

// RegisterRoutes 注册路由
func (h *FileHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.POST("/files/upload", h.Upload)
	auth.GET("/files", h.ListByUser)
	auth.GET("/files/:id", h.Download)
	auth.GET("/conversations/:id/files", h.ListByConv)
}
