package model

import "time"

type ApprovalTemplate struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID      *int64    `gorm:"index" json:"project_id"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	FormSchema     string    `gorm:"type:jsonb;not null;default:'{}'" json:"form_schema"`
	FlowDefinition string    `gorm:"type:jsonb;not null;default:'{}'" json:"flow_definition"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type ApprovalInstance struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID    int64      `gorm:"not null" json:"template_id"`
	ProjectID     *int64     `gorm:"index" json:"project_id"`
	Title         string     `gorm:"size:500;not null" json:"title"`
	FormData      string     `gorm:"type:jsonb;not null;default:'{}'" json:"form_data"`
	Status        string     `gorm:"size:30;not null;default:pending" json:"status"`
	SubmitterID   int64      `gorm:"not null;index" json:"submitter_id"`
	CurrentNodeID string     `gorm:"size:64" json:"current_node_id"`
	SubmittedAt   time.Time  `json:"submitted_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ApprovalAction struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	InstanceID int64     `gorm:"not null;index" json:"instance_id"`
	NodeID     string    `gorm:"size:64;not null" json:"node_id"`
	Action     string    `gorm:"size:20;not null" json:"action"`
	Comment    string    `gorm:"type:text" json:"comment"`
	UserID     int64     `gorm:"not null" json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}
