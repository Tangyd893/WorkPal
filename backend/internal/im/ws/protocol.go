package ws

import (
	"encoding/json"
	"time"
)

// MessageType 消息类型
const (
	TypeChat     = "chat"      // 聊天消息
	TypePresence = "presence"  // 在线状态
	TypeAck      = "ack"       // 消息确认
	TypePing     = "ping"      // 心跳
	TypePong     = "pong"      // 心跳响应
	TypeError    = "error"     // 错误
	TypeRead     = "read"      // 已读回执
	TypeReadAll  = "read_all"  // 全部已读
)

// WSMessage WebSocket 消息格式
type WSMessage struct {
	Type      string      `json:"type"`
	From     int64       `json:"from"`
	To       string      `json:"to"`        // user_id 或 conv_id
	ConvID   int64       `json:"conv_id"`   // 会话ID
	Content  interface{} `json:"content"`   // 消息内容
	Seq      int64       `json:"seq"`       // 客户端本地序列号
	CreatedAt string     `json:"created_at"`// 服务器时间戳
	UserID   int64       `json:"user_id"`   // 用于 presence 类型
	Status   string      `json:"status"`    // online/offline/busy
}

// Marshal JSON 序列化
func (m *WSMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal JSON 反序列化
func Unmarshal(data []byte) (*WSMessage, error) {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewChatMsg 创建聊天消息
func NewChatMsg(from int64, convID int64, content interface{}, seq int64) *WSMessage {
	return &WSMessage{
		Type:      TypeChat,
		From:     from,
		ConvID:   convID,
		Content:  content,
		Seq:      seq,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
}

// NewPresenceMsg 创建在线状态消息
func NewPresenceMsg(userID int64, status string) *WSMessage {
	return &WSMessage{
		Type:    TypePresence,
		UserID:  userID,
		Status:  status,
	}
}

// NewAckMsg 创建确认消息
func NewAckMsg(seq int64) *WSMessage {
	return &WSMessage{
		Type: TypeAck,
		Seq:  seq,
	}
}

// NewErrorMsg 创建错误消息
func NewErrorMsg(msg string) *WSMessage {
	return &WSMessage{
		Type:    TypeError,
		Content: msg,
	}
}

// NewReadMsg 创建已读回执
func NewReadMsg(convID int64, lastReadMsgID int64) *WSMessage {
	return &WSMessage{
		Type:    TypeRead,
		ConvID:  convID,
		Seq:     lastReadMsgID,
	}
}

// NewReadAllMsg 创建全部已读回执
func NewReadAllMsg(convID int64) *WSMessage {
	return &WSMessage{
		Type:   TypeReadAll,
		ConvID: convID,
	}
}
