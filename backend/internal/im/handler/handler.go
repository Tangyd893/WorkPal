package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/pagination"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/service"
	"github.com/gin-gonic/gin"
	ws "github.com/Tangyd893/WorkPal/backend/internal/im/ws"
	"github.com/gorilla/websocket"
)

type ConversationHandler struct {
	convSvc *service.ConversationService
}

func NewConversationHandler(convSvc *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{convSvc: convSvc}
}

// CreateConvReq 创建会话请求
type CreateConvReq struct {
	Type      int8    `json:"type"` // 1=私聊 2=群聊，默认私聊
	TargetUID int64   `json:"target_uid"` // 私聊目标用户ID
	Name      string  `json:"name"`        // 群名（群聊时）
	MemberIDs []int64 `json:"member_ids"`  // 群聊成员ID列表
}

// Create 创建会话
// POST /api/v1/conversations
func (h *ConversationHandler) Create(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req CreateConvReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 私聊
	if req.Type == model.ConversationTypePrivate || req.TargetUID > 0 {
		conv, err := h.convSvc.CreatePrivateConv(c.Request.Context(), userID, req.TargetUID)
		if err != nil {
			handleServiceErr(c, err)
			return
		}
		response.Success(c, conv)
		return
	}

	// 群聊
	if req.Type == model.ConversationTypeGroup {
		conv, err := h.convSvc.CreateGroup(c.Request.Context(), req.Name, userID, req.MemberIDs)
		if err != nil {
			handleServiceErr(c, err)
			return
		}
		response.Success(c, conv)
		return
	}

	response.FailWithMessage(c, http.StatusBadRequest, "无效的会话类型")
}

// List 获取会话列表
// GET /api/v1/conversations
func (h *ConversationHandler) List(c *gin.Context) {
	userID := c.GetInt64("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = pagination.GetParams(page, pageSize)

	convs, total, err := h.convSvc.ListByUser(c.Request.Context(), userID, pagination.GetOffset(page, pageSize), pageSize)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.SuccessPage(c, convs, total, page, pageSize)
}

// Get 获取会话详情
// GET /api/v1/conversations/:id
func (h *ConversationHandler) Get(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话ID")
		return
	}

	// 检查是否是成员
	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if !isMember {
		response.Fail(c, errors.ErrForbidden)
		return
	}

	conv, err := h.convSvc.GetByID(c.Request.Context(), convID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, conv)
}

// Update 更新会话（群名）
// PUT /api/v1/conversations/:id
func (h *ConversationHandler) Update(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话ID")
		return
	}

	conv, err := h.convSvc.GetByID(c.Request.Context(), convID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}

	// 私聊不能改名，群聊只有群主能改
	if conv.Type == model.ConversationTypePrivate {
		response.FailWithMessage(c, http.StatusBadRequest, "私聊无法修改")
		return
	}
	if conv.OwnerID != userID {
		response.Fail(c, errors.ErrForbidden)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误")
		return
	}

	conv.Name = req.Name
	if err := h.convSvc.Update(c.Request.Context(), conv); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, conv)
}

// Delete 解散会话
// DELETE /api/v1/conversations/:id
func (h *ConversationHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话ID")
		return
	}

	if err := h.convSvc.Delete(c.Request.Context(), convID, userID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

// AddMember 添加成员
// POST /api/v1/conversations/:id/members
func (h *ConversationHandler) AddMember(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话ID")
		return
	}

	// 检查操作者是否是成员
	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if !isMember {
		response.Fail(c, errors.ErrForbidden)
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.convSvc.AddMember(c.Request.Context(), convID, req.UserID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

// RemoveMember 移除成员
// DELETE /api/v1/conversations/:id/members/:uid
func (h *ConversationHandler) RemoveMember(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话ID")
		return
	}
	targetUID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	// 检查操作者是否是成员
	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if !isMember {
		response.Fail(c, errors.ErrForbidden)
		return
	}

	if err := h.convSvc.RemoveMember(c.Request.Context(), convID, targetUID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

// RegisterRoutes 注册路由
func (h *ConversationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.POST("/conversations", h.Create)
	auth.GET("/conversations", h.List)
	auth.GET("/conversations/:id", h.Get)
	auth.PUT("/conversations/:id", h.Update)
	auth.DELETE("/conversations/:id", h.Delete)
	auth.POST("/conversations/:id/members", h.AddMember)
	auth.DELETE("/conversations/:id/members/:uid", h.RemoveMember)
}

func handleServiceErr(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		response.Fail(c, appErr)
		return
	}
	response.FailWithMessage(c, http.StatusInternalServerError, "内部错误: "+err.Error())
}

// WebSocketHandler WebSocket 处理
type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

// Handle WebSocket 升级处理
// WSS /ws
func (h *WebSocketHandler) Handle(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "升级失败"})
		return
	}

	hub := ws.GetHub()
	client := ws.NewClient(userID, conn, hub)
	client.Run(c.Request.Context())
}
