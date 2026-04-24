package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/service"
	"github.com/Tangyd893/WorkPal/backend/internal/im/ws"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	msgSvc *service.MessageService
	convSvc *service.ConversationService
	hub    *ws.Hub
}

func NewMessageHandler(msgSvc *service.MessageService, convSvc *service.ConversationService, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{
		msgSvc:  msgSvc,
		convSvc: convSvc,
		hub:     hub,
	}
}

// SendReq 发送消息请求
type SendReq struct {
	Type      int8                  `json:"type"`      // 消息类型
	Content   string                `json:"content"`    // 消息内容
	Metadata  map[string]interface{} `json:"metadata"`   // 扩展字段
	ReplyTo   int64                 `json:"reply_to"`   // 回复的消息ID
}

// GetHistory 获取历史消息（分页）
// GET /api/v1/conversations/:id/messages
func (h *MessageHandler) GetHistory(c *gin.Context) {
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

	beforeID, _ := strconv.ParseInt(c.DefaultQuery("before_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	msgs, err := h.msgSvc.GetHistory(c.Request.Context(), convID, beforeID, limit)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, msgs)
}

// Send 发送消息（HTTP 备用）
// POST /api/v1/conversations/:id/messages
func (h *MessageHandler) Send(c *gin.Context) {
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

	var req SendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	msgType := req.Type
	if msgType == 0 {
		msgType = model.MessageTypeText
	}

	msg, err := h.msgSvc.Send(c.Request.Context(), convID, userID, msgType, req.Content, req.Metadata, req.ReplyTo)
	if err != nil {
		handleServiceErr(c, err)
		return
	}

	// 通过 WebSocket 广播到房间
	wsMsg := ws.NewChatMsg(userID, convID, req.Content, 0)
	wsMsg.CreatedAt = msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	wsData, _ := wsMsg.Marshal()
	if h.hub != nil {
		h.hub.BroadcastToRoom(convID, userID, wsData, nil)
	}

	response.Success(c, msg)
}

// Edit 编辑消息
// PUT /api/v1/messages/:id
func (h *MessageHandler) Edit(c *gin.Context) {
	userID := c.GetInt64("userID")
	msgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的消息ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误")
		return
	}

	msg, err := h.msgSvc.Edit(c.Request.Context(), msgID, userID, req.Content)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, msg)
}

// Delete 撤回消息
// DELETE /api/v1/messages/:id
func (h *MessageHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("userID")
	msgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的消息ID")
		return
	}

	if err := h.msgSvc.Recall(c.Request.Context(), msgID, userID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

// MarkRead 标记已读
// POST /api/v1/messages/:id/read
func (h *MessageHandler) MarkRead(c *gin.Context) {
	userID := c.GetInt64("userID")
	msgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的消息ID")
		return
	}

	// 获取消息对应的会话
	msg, err := h.msgSvc.GetByID(c.Request.Context(), msgID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}

	if err := h.msgSvc.MarkRead(c.Request.Context(), userID, msg.ConvID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

// MarkReadAll 标记会话全部已读
// POST /api/v1/conversations/:id/read-all
func (h *MessageHandler) MarkReadAll(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话ID")
		return
	}

	if err := h.msgSvc.MarkRead(c.Request.Context(), userID, convID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

// RegisterRoutes 注册路由
func (h *MessageHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/conversations/:id/messages", h.GetHistory)
	auth.POST("/conversations/:id/messages", h.Send)
	auth.PUT("/messages/:id", h.Edit)
	auth.DELETE("/messages/:id", h.Delete)
	auth.POST("/messages/:id/read", h.MarkRead)
	auth.POST("/conversations/:id/read-all", h.MarkReadAll)
}
