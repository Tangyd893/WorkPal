package handler

import (
	"context"
	"net/http"
	"strconv"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/service"
	imWS "github.com/Tangyd893/WorkPal/backend/internal/im/ws"
	searchSvc "github.com/Tangyd893/WorkPal/backend/internal/search/service"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	msgSvc    *service.MessageService
	convSvc   *service.ConversationService
	hub       *imWS.Hub
	searchSvc *searchSvc.SearchService
	cluster   clusterUserBroadcaster
}

type clusterUserBroadcaster interface {
	BroadcastUsers(ctx context.Context, userIDs []int64, content []byte) error
}

type SendReq struct {
	Type     int8                   `json:"type"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	ReplyTo  int64                  `json:"reply_to"`
}

func NewMessageHandler(
	msgSvc *service.MessageService,
	convSvc *service.ConversationService,
	hub *imWS.Hub,
	searchSvc *searchSvc.SearchService,
	cluster clusterUserBroadcaster,
) *MessageHandler {
	return &MessageHandler{
		msgSvc:    msgSvc,
		convSvc:   convSvc,
		hub:       hub,
		searchSvc: searchSvc,
		cluster:   cluster,
	}
}

func (h *MessageHandler) GetHistory(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}

	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if !isMember {
		response.Fail(c, apperrors.ErrPermissionDenied)
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

func (h *MessageHandler) Send(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}

	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if !isMember {
		response.Fail(c, apperrors.ErrPermissionDenied)
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

	wsMsg := imWS.NewChatMsg(userID, convID, req.Content, 0)
	wsMsg.ID = msg.ID
	wsMsg.CreatedAt = msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	wsData, _ := wsMsg.Marshal()
	if h.hub != nil {
		h.broadcastToConversation(c.Request.Context(), convID, wsData)
	}

	if h.searchSvc != nil {
		_ = h.searchSvc.IndexMessage(msg)
	}

	response.Success(c, msg)
}

func (h *MessageHandler) Edit(c *gin.Context) {
	userID := c.GetInt64("userID")
	msgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的消息 ID")
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
	if h.searchSvc != nil {
		_ = h.searchSvc.IndexMessage(msg)
	}
	response.Success(c, msg)
}

func (h *MessageHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("userID")
	msgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的消息 ID")
		return
	}

	if err := h.msgSvc.Recall(c.Request.Context(), msgID, userID); err != nil {
		handleServiceErr(c, err)
		return
	}
	if h.searchSvc != nil {
		_ = h.searchSvc.DeleteMessage(msgID)
	}
	response.Success(c, nil)
}

func (h *MessageHandler) MarkRead(c *gin.Context) {
	userID := c.GetInt64("userID")
	msgID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的消息 ID")
		return
	}

	msg, err := h.msgSvc.GetByID(c.Request.Context(), msgID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	isMember, err := h.convSvc.IsMember(c.Request.Context(), msg.ConvID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if !isMember {
		response.Fail(c, apperrors.ErrPermissionDenied)
		return
	}

	if err := h.msgSvc.MarkRead(c.Request.Context(), userID, msg.ConvID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *MessageHandler) MarkReadAll(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}

	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if !isMember {
		response.Fail(c, apperrors.ErrPermissionDenied)
		return
	}

	if err := h.msgSvc.MarkRead(c.Request.Context(), userID, convID); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, nil)
}

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

func (h *MessageHandler) broadcastToConversation(ctx context.Context, convID int64, wsData []byte) {
	members, err := h.convSvc.GetMembers(ctx, convID)
	if err != nil {
		if h.hub != nil {
			h.hub.BroadcastToRoom(convID, 0, wsData, nil)
		}
		return
	}
	if h.cluster != nil {
		if err := h.cluster.BroadcastUsers(ctx, members, wsData); err == nil {
			return
		}
	}
	if h.hub == nil {
		return
	}
	for _, userID := range members {
		h.hub.SendToUser(userID, wsData)
	}
}
