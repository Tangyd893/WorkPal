package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ConversationTypePrivate = 1
	ConversationTypeGroup   = 2
)

type Conversation struct {
	ID                    int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Type                  int8           `gorm:"not null;default:1" json:"type"`
	Name                  string         `gorm:"size:256" json:"name"`
	AvatarURL             string         `gorm:"size:512" json:"avatar_url"`
	OwnerID               int64          `gorm:"default:0" json:"owner_id"`
	Announcement          string         `gorm:"type:text" json:"announcement"`
	AnnouncementUpdatedAt *time.Time     `json:"announcement_updated_at,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Conversation) TableName() string {
	return "conversations"
}

type ConversationMember struct {
	ID       int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ConvID   int64     `gorm:"not null;index;uniqueIndex:uniq_conv_member" json:"conv_id"`
	UserID   int64     `gorm:"not null;index;uniqueIndex:uniq_conv_member" json:"user_id"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

func (ConversationMember) TableName() string {
	return "conversation_members"
}

const (
	MessageTypeText  = 1
	MessageTypeImage = 2
	MessageTypeFile  = 3
	MessageTypeCode  = 4
	MessageTypeCard  = 5
)

type Message struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ConvID    int64          `gorm:"not null;index" json:"conv_id"`
	SenderID  int64          `gorm:"not null;index" json:"sender_id"`
	Type      int8           `gorm:"not null;default:1" json:"type"`
	Content   string         `gorm:"type:text" json:"content"`
	Metadata  string         `gorm:"type:jsonb" json:"metadata"`
	ReplyTo   int64          `gorm:"default:0" json:"reply_to"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Message) TableName() string {
	return "messages"
}

const (
	OutboxStatusPending    = "pending"
	OutboxStatusPublishing = "publishing"
	OutboxStatusDelivered  = "delivered"
)

type MessageOutbox struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Topic         string     `gorm:"size:128;index" json:"topic"`
	Payload       string     `gorm:"type:jsonb" json:"payload"`
	Status        string     `gorm:"size:32;index" json:"status"`
	RetryCount    int        `gorm:"not null;default:0" json:"retry_count"`
	LastError     string     `gorm:"type:text" json:"last_error"`
	NextAttemptAt time.Time  `gorm:"index" json:"next_attempt_at"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (MessageOutbox) TableName() string {
	return "message_outbox"
}

type MessageRead struct {
	UserID int64     `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	ConvID int64     `gorm:"primaryKey;autoIncrement:false" json:"conv_id"`
	ReadAt time.Time `gorm:"autoUpdateTime" json:"read_at"`
}

func (MessageRead) TableName() string {
	return "message_reads"
}

type ConvMember struct {
	ConvID    int64  `json:"conv_id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}
