package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/im/service"
	searchSvc "github.com/Tangyd893/WorkPal/backend/internal/search/service"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchSvc *searchSvc.SearchService
	convSvc  *service.ConversationService
}

func NewSearchHandler(searchSvc *searchSvc.SearchService, convSvc *service.ConversationService) *SearchHandler {
	return &SearchHandler{
		searchSvc: searchSvc,
		convSvc:   convSvc,
	}
}

// Search 搜索消息
// GET /api/v1/search/messages?q=keyword
func (h *SearchHandler) Search(c *gin.Context) {
	userID := c.GetInt64("userID")
	query := c.Query("q")
	if query == "" {
		response.FailWithMessage(c, http.StatusBadRequest, "搜索关键词不能为空")
		return
	}

	convID, _ := strconv.ParseInt(c.DefaultQuery("conv_id", "0"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 如果指定了会话，检查权限
	if convID > 0 {
		isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
		if err != nil || !isMember {
			response.FailWithMessage(c, http.StatusForbidden, "无权限搜索该会话")
			return
		}
		result, err := h.searchSvc.SearchInConv(c.Request.Context(), convID, query, page, pageSize)
		if err != nil {
			response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Success(c, result)
		return
	}

	// 全局搜索
	result, err := h.searchSvc.GlobalSearch(c.Request.Context(), query, page, pageSize)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

// RegisterRoutes 注册路由
func (h *SearchHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/search/messages", h.Search)
}
