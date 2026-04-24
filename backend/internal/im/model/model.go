package model

import (
	"time"

	"gorm.io/gorm"
)

// ConversationType 会话类型
const (
	ConversationTypePrivate = 1 // 私聊
	ConversationTypeGroup  = 2 // 群聊
)

// Conversation 会话
type Conversation struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      int8           `gorm:"not null;default:1" json:"type"` // 1=私聊 2=群聊
	Name      string         `gorm:"size:256" json:"name"`           // 群名（私聊为空）
	AvatarURL string          `gorm:"size:512" json:"avatar_url"`
	OwnerID   int64           `gorm:"default:0" json:"owner_id"`    // 群主（私聊为空）
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// ConversationMember 会话成员
type ConversationMember struct {
	ID       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConvID   int64     `gorm:"not null;index" json:"conv_id"`
	UserID   int64     `gorm:"not null;index" json:"user_id"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

func (ConversationMember) TableName() string {
	return "conversation_members"
}

// MessageType 消息类型
const (
	MessageTypeText   = 1 // 文字
	MessageTypeImage  = 2 // 图片
	MessageTypeFile  = 3 // 文件
	MessageTypeCode  = 4 // 代码
	MessageTypeCard  = 5 // 卡片
)

// Message 消息
type Message struct {
	ID        int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ConvID    int64           `gorm:"not null;index" json:"conv_id"`
	SenderID  int64           `gorm:"not null;index" json:"sender_id"`
	Type      int8            `gorm:"not null;default:1" json:"type"` // 1=文字 2=图片 3=文件 4=代码 5=卡片
	Content   string          `gorm:"type:text" json:"content"`
	Metadata  string          `gorm:"type:jsonb" json:"metadata"` // JSON存储扩展字段
	ReplyTo   int64           `gorm:"default:0" json:"reply_to"`  // 回复的消息ID
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Message) TableName() string {
	return "messages"
}

// MessageRead 已读状态
type MessageRead struct {
	UserID   int64     `gorm:"primaryKey;autoIncrement" json:"user_id"`
	ConvID   int64     `gorm:"primaryKey;autoIncrement" json:"conv_id"`
	ReadAt   time.Time `gorm:"autoUpdateTime" json:"read_at"`
}

func (MessageRead) TableName() string {
	return "message_reads"
}

// ConvMember 视图用（JOIN查询时携带用户信息）
type ConvMember struct {
	ConvID   int64  `json:"conv_id"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}
