package handler

import (
	"net/http"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/gin-gonic/gin"
)

type MeetingRoom struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	EventID  *int64  `json:"event_id"`
	HostID   int64   `json:"host_id"`
}

type CreateRoomInput struct {
	Name    string `json:"name"`
	EventID *int64 `json:"event_id"`
}

var rooms = make(map[string]*MeetingRoom)

func RegisterMeetingRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.POST("/meetings/rooms", createMeetingRoom)
	auth.GET("/meetings/rooms/:id", getMeetingRoom)
	auth.DELETE("/meetings/rooms/:id", deleteMeetingRoom)
}

func createMeetingRoom(c *gin.Context) {
	var input CreateRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	roomID := "room-" + c.GetString("requestID")
	room := &MeetingRoom{ID: roomID, Name: input.Name, EventID: input.EventID, HostID: c.GetInt64("userID")}
	rooms[roomID] = room
	response.Success(c, room)
}

func getMeetingRoom(c *gin.Context) {
	room, ok := rooms[c.Param("id")]
	if !ok {
		response.FailWithMessage(c, http.StatusNotFound, "room not found")
		return
	}
	response.Success(c, room)
}

func deleteMeetingRoom(c *gin.Context) {
	delete(rooms, c.Param("id"))
	response.Success(c, nil)
}
