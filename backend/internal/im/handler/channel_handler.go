package handler

import (
	"net/http"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/gin-gonic/gin"
)

type ChannelHandler struct {
}

func NewChannelHandler() *ChannelHandler {
	return &ChannelHandler{}
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
	response.Success(c, []interface{}{})
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
	response.Success(c, gin.H{
		"id":           "chn-1",
		"name":         input.Name,
		"channel_type": input.ChannelType,
	})
}

func (ch *ChannelHandler) GetChannel(c *gin.Context) {
	response.Success(c, gin.H{"id": c.Param("id")})
}

func (ch *ChannelHandler) DeleteChannel(c *gin.Context) {
	response.Success(c, nil)
}

func (ch *ChannelHandler) AddMember(c *gin.Context) {
	response.Success(c, nil)
}

func (ch *ChannelHandler) RemoveMember(c *gin.Context) {
	response.Success(c, nil)
}

func (ch *ChannelHandler) ListThreads(c *gin.Context) {
	response.Success(c, []interface{}{})
}

func (ch *ChannelHandler) CreateThread(c *gin.Context) {
	var input struct {
		ParentMsgID int64  `json:"parent_msg_id"`
		Title       string `json:"title"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid thread payload")
		return
	}
	response.Success(c, gin.H{"id": 1, "title": input.Title})
}
