package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/repo"
	"github.com/gin-gonic/gin"
)

type ChannelHandler struct {
	channelRepo *repo.ChannelRepo
}

func NewChannelHandler(channelRepo *repo.ChannelRepo) *ChannelHandler {
	return &ChannelHandler{channelRepo: channelRepo}
}

func (ch *ChannelHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/channels", ch.ListChannels)
	auth.POST("/channels", ch.CreateChannel)
	auth.GET("/channels/:id", ch.GetChannel)
	auth.DELETE("/channels/:id", ch.DeleteChannel)
	auth.POST("/channels/:id/members", ch.AddMember)
	auth.DELETE("/channels/:id/members/:userId", ch.RemoveMember)
	auth.GET("/channels/:id/threads", ch.ListThreads)
	auth.POST("/channels/:id/threads", ch.CreateThread)
}

func (ch *ChannelHandler) ListChannels(c *gin.Context) {
	channels, err := ch.channelRepo.List(c.Request.Context())
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, "查询频道列表失败")
		return
	}
	if channels == nil {
		channels = []*model.Channel{}
	}
	response.Success(c, channels)
}

func (ch *ChannelHandler) CreateChannel(c *gin.Context) {
	var input struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ChannelType string  `json:"channel_type"`
		ProjectID   *int64  `json:"project_id"`
		MemberIDs   []int64 `json:"member_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid channel payload")
		return
	}
	if input.Name == "" {
		response.FailWithMessage(c, http.StatusBadRequest, "频道名称不能为空")
		return
	}
	userID := c.GetInt64("userID")
	channelType := input.ChannelType
	if channelType == "" {
		channelType = "public"
	}
	channel := &model.Channel{
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		ChannelType: channelType,
		CreatedBy:   userID,
	}
	if err := ch.channelRepo.Create(c.Request.Context(), channel); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, "创建频道失败")
		return
	}
	if err := ch.channelRepo.AddMember(c.Request.Context(), &model.ChannelMember{
		ChannelID: channel.ID,
		UserID:    userID,
		Role:      "owner",
	}); err != nil {
		_ = ch.channelRepo.Delete(c.Request.Context(), channel.ID)
		response.FailWithMessage(c, http.StatusInternalServerError, "添加频道创建者失败")
		return
	}
	for _, memberID := range input.MemberIDs {
		if memberID == userID {
			continue
		}
		_ = ch.channelRepo.AddMember(c.Request.Context(), &model.ChannelMember{
			ChannelID: channel.ID,
			UserID:    memberID,
			Role:      "member",
		})
	}
	response.Success(c, channel)
}

func (ch *ChannelHandler) GetChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的频道ID")
		return
	}
	channel, err := ch.channelRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "频道不存在")
		return
	}
	members, _ := ch.channelRepo.ListMembers(c.Request.Context(), id)
	result := gin.H{
		"id":           channel.ID,
		"project_id":   channel.ProjectID,
		"name":         channel.Name,
		"description":  channel.Description,
		"channel_type": channel.ChannelType,
		"created_by":   channel.CreatedBy,
		"is_archived":  channel.IsArchived,
		"created_at":   channel.CreatedAt,
		"updated_at":   channel.UpdatedAt,
		"members":      members,
	}
	response.Success(c, result)
}

func (ch *ChannelHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的频道ID")
		return
	}
	channel, err := ch.channelRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage(c, http.StatusNotFound, "频道不存在")
		return
	}
	userID := c.GetInt64("userID")
	if channel.CreatedBy != userID {
		response.FailWithMessage(c, http.StatusForbidden, "只有频道创建者可以删除频道")
		return
	}
	if err := ch.channelRepo.Delete(c.Request.Context(), id); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, "删除频道失败")
		return
	}
	response.Success(c, nil)
}

func (ch *ChannelHandler) AddMember(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的频道ID")
		return
	}
	var input struct {
		UserID int64 `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if input.UserID == 0 {
		response.FailWithMessage(c, http.StatusBadRequest, "userId不能为空")
		return
	}
	if isMember, _ := ch.channelRepo.IsMember(c.Request.Context(), channelID, input.UserID); isMember {
		response.FailWithMessage(c, http.StatusConflict, "用户已是频道成员")
		return
	}
	if err := ch.channelRepo.AddMember(c.Request.Context(), &model.ChannelMember{
		ChannelID: channelID,
		UserID:    input.UserID,
		Role:      "member",
	}); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, "添加成员失败")
		return
	}
	response.Success(c, nil)
}

func (ch *ChannelHandler) RemoveMember(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的频道ID")
		return
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的用户ID")
		return
	}
	if err := ch.channelRepo.RemoveMember(c.Request.Context(), channelID, userID); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, "移除成员失败")
		return
	}
	response.Success(c, nil)
}

func (ch *ChannelHandler) ListThreads(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的频道ID")
		return
	}
	threads, err := ch.channelRepo.ListThreads(c.Request.Context(), channelID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, "查询话题列表失败")
		return
	}
	if threads == nil {
		threads = []*model.Thread{}
	}
	response.Success(c, threads)
}

func (ch *ChannelHandler) CreateThread(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "无效的频道ID")
		return
	}
	var input struct {
		ParentMsgID int64  `json:"parent_msg_id"`
		Title       string `json:"title"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid thread payload")
		return
	}
	if input.ParentMsgID == 0 {
		response.FailWithMessage(c, http.StatusBadRequest, "parentMsgId不能为空")
		return
	}
	thread := &model.Thread{
		ChannelID:   channelID,
		ParentMsgID: input.ParentMsgID,
		Title:       input.Title,
	}
	if err := ch.channelRepo.CreateThread(c.Request.Context(), thread); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, "创建话题失败")
		return
	}
	response.Success(c, thread)
}
