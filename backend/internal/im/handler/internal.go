package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/gin-gonic/gin"
)

func (h *ConversationHandler) RegisterInternalRoutes(rg *gin.RouterGroup) {
	internal := rg.Group("/internal")
	internal.GET("/conversations/:id/members/:uid", h.InternalIsMember)
	internal.GET("/users/:uid/conversations", h.InternalListUserConversationIDs)
}

func (h *ConversationHandler) InternalIsMember(c *gin.Context) {
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid conversation id")
		return
	}
	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}

	isMember, err := h.convSvc.IsMember(c.Request.Context(), convID, userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, gin.H{"is_member": isMember})
}

func (h *ConversationHandler) InternalListUserConversationIDs(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}

	ids, err := h.convSvc.ListUserConversationIDs(c.Request.Context(), userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, gin.H{"conv_ids": ids})
}
