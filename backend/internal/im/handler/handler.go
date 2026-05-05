package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/audit"
	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/pagination"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/service"
	ws "github.com/Tangyd893/WorkPal/backend/internal/im/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ConversationHandler struct {
	convSvc *service.ConversationService
	audit   *audit.Recorder
}

func NewConversationHandler(convSvc *service.ConversationService, recorders ...*audit.Recorder) *ConversationHandler {
	var recorder *audit.Recorder
	if len(recorders) > 0 {
		recorder = recorders[0]
	}
	return &ConversationHandler{convSvc: convSvc, audit: recorder}
}

type CreateConvReq struct {
	Type      int8    `json:"type"`
	TargetUID int64   `json:"target_uid"`
	Name      string  `json:"name"`
	MemberIDs []int64 `json:"member_ids"`
}

func (h *ConversationHandler) Create(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req CreateConvReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if req.Type == model.ConversationTypePrivate || req.TargetUID > 0 {
		conv, err := h.convSvc.CreatePrivateConv(c.Request.Context(), userID, req.TargetUID)
		if err != nil {
			handleServiceErr(c, err)
			return
		}
		response.Success(c, conv)
		return
	}

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

func (h *ConversationHandler) Get(c *gin.Context) {
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

	conv, err := h.convSvc.GetByID(c.Request.Context(), convID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, conv)
}

func (h *ConversationHandler) Update(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}

	conv, err := h.convSvc.GetByID(c.Request.Context(), convID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	if conv.Type == model.ConversationTypePrivate {
		response.FailWithMessage(c, http.StatusBadRequest, "私聊无法修改")
		return
	}
	if conv.OwnerID != userID {
		response.Fail(c, apperrors.ErrPermissionDenied)
		return
	}

	var req struct {
		Name         *string `json:"name"`
		Announcement *string `json:"announcement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Name != nil {
		conv.Name = *req.Name
	}
	if req.Announcement != nil {
		now := time.Now()
		conv.Announcement = *req.Announcement
		conv.AnnouncementUpdatedAt = &now
	}

	if err := h.convSvc.Update(c.Request.Context(), conv); err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, conv)
}

func (h *ConversationHandler) UpdateAnnouncement(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}

	var req struct {
		Announcement string `json:"announcement"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误")
		return
	}

	conv, err := h.convSvc.UpdateAnnouncement(c.Request.Context(), convID, userID, req.Announcement)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, conv)
}

func (h *ConversationHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}

	if err := h.convSvc.Delete(c.Request.Context(), convID, userID); err != nil {
		handleServiceErr(c, err)
		return
	}
	h.audit.Record(c.Request.Context(), userID, "删除会话", "conversation", strconv.FormatInt(convID, 10), c.ClientIP())
	response.Success(c, nil)
}

func (h *ConversationHandler) AddMember(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}

	if !h.ensureGroupOwner(c, convID, userID) {
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

func (h *ConversationHandler) RemoveMember(c *gin.Context) {
	userID := c.GetInt64("userID")
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的会话 ID")
		return
	}
	targetUID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的用户 ID")
		return
	}

	if !h.ensureGroupOwner(c, convID, userID) {
		return
	}

	if err := h.convSvc.RemoveMember(c.Request.Context(), convID, targetUID); err != nil {
		handleServiceErr(c, err)
		return
	}
	h.audit.Record(c.Request.Context(), userID, "移出群成员", "conversation_member", strconv.FormatInt(targetUID, 10), c.ClientIP())
	response.Success(c, nil)
}

func (h *ConversationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.POST("/conversations", h.Create)
	auth.GET("/conversations", h.List)
	auth.GET("/conversations/:id", h.Get)
	auth.PUT("/conversations/:id", h.Update)
	auth.PUT("/conversations/:id/announcement", h.UpdateAnnouncement)
	auth.DELETE("/conversations/:id", h.Delete)
	auth.POST("/conversations/:id/members", h.AddMember)
	auth.DELETE("/conversations/:id/members/:uid", h.RemoveMember)
}

func (h *ConversationHandler) ensureGroupOwner(c *gin.Context, convID, userID int64) bool {
	conv, err := h.convSvc.GetByID(c.Request.Context(), convID)
	if err != nil {
		handleServiceErr(c, err)
		return false
	}
	if conv.Type == model.ConversationTypePrivate {
		response.FailWithMessage(c, http.StatusBadRequest, "私聊无法操作成员")
		return false
	}
	if conv.OwnerID != userID {
		response.Fail(c, apperrors.ErrPermissionDenied)
		return false
	}
	return true
}

func handleServiceErr(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		response.Fail(c, appErr)
		return
	}
	response.FailWithMessage(c, http.StatusInternalServerError, "内部错误: "+err.Error())
}

type WebSocketHandler struct {
	hub     *ws.Hub
	convSvc *service.ConversationService
}

func NewWebSocketHandler(hub *ws.Hub, convSvc *service.ConversationService) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, convSvc: convSvc}
}

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
			return ws.CheckOrigin(r)
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "升级失败"})
		return
	}

	hub := h.hub
	if hub == nil {
		hub = ws.GetHub()
	}
	if hub == nil {
		conn.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket hub not ready"})
		return
	}
	client := ws.NewClient(userID, conn, hub)
	if h.convSvc != nil {
		convs, _, err := h.convSvc.ListByUser(c.Request.Context(), userID, 0, 1000)
		if err == nil {
			for _, conv := range convs {
				hub.JoinRoom(client, conv.ID)
			}
		}
	}
	client.Run(c.Request.Context())
}
