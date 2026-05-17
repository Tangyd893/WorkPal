package model

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Key         string         `gorm:"size:10;not null;uniqueIndex" json:"key"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	LeadID      int64          `gorm:"not null;index" json:"lead_id"`
	Icon        string         `gorm:"size:50" json:"icon"`
	Category    string         `gorm:"size:50;default:software" json:"category"`
	IsArchived  bool           `gorm:"default:false" json:"is_archived"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type IssueType struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID      int64     `gorm:"not null;index" json:"project_id"`
	Name           string    `gorm:"size:100;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	IconURL        string    `gorm:"size:255" json:"icon_url"`
	HierarchyLevel int       `gorm:"default:0" json:"hierarchy_level"`
	IsStandard     bool      `gorm:"default:true" json:"is_standard"`
	CreatedAt      time.Time `json:"created_at"`
}

type Issue struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID     int64          `gorm:"not null;index" json:"project_id"`
	IssueTypeID   int64          `gorm:"not null;index" json:"issue_type_id"`
	ParentID      *int64         `gorm:"index" json:"parent_id"`
	Key           string         `gorm:"size:50;not null;uniqueIndex" json:"key"`
	Summary       string         `gorm:"size:500;not null" json:"summary"`
	Description   string         `gorm:"type:text" json:"description"`
	Status        string         `gorm:"size:50;not null;default:Open;index" json:"status"`
	Priority      string         `gorm:"size:20;not null;default:Medium" json:"priority"`
	AssigneeID    *int64         `gorm:"index" json:"assignee_id"`
	ReporterID    int64          `gorm:"not null;index" json:"reporter_id"`
	DueDate       *string        `json:"due_date"`
	StoryPoints   *float64       `json:"story_points"`
	Resolution    string         `gorm:"size:50" json:"resolution"`
	SprintID      *int64         `json:"sprint_id"`
	VersionIDs    string         `gorm:"type:jsonb" json:"-"`
	FixVersionIDs string         `gorm:"type:jsonb" json:"-"`
	TimeEstimate  int            `gorm:"default:0" json:"time_estimate"`
	TimeSpent     int            `gorm:"default:0" json:"time_spent"`
	SortOrder     int            `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Board struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID int64     `gorm:"not null;index" json:"project_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	BoardType string    `gorm:"size:20;not null;default:kanban" json:"board_type"`
	Config    string    `gorm:"type:jsonb;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Version struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID   int64     `gorm:"not null;index" json:"project_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	StartDate   *string   `json:"start_date"`
	ReleaseDate *string   `json:"release_date"`
	IsArchived  bool      `gorm:"default:false" json:"is_archived"`
	IsReleased  bool      `gorm:"default:false" json:"is_released"`
	CreatedAt   time.Time `json:"created_at"`
}

type IssueChangelog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID   int64     `gorm:"not null;index" json:"issue_id"`
	Field     string    `gorm:"size:100;not null" json:"field"`
	OldValue  string    `gorm:"type:text" json:"old_value"`
	NewValue  string    `gorm:"type:text" json:"new_value"`
	ChangedBy int64     `gorm:"not null" json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

type Association struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceType string    `gorm:"size:50;not null;index:idx_assoc_source,composite:source" json:"source_type"`
	SourceID   int64     `gorm:"not null;index:idx_assoc_source,composite:source" json:"source_id"`
	TargetType string    `gorm:"size:50;not null;index:idx_assoc_target,composite:target" json:"target_type"`
	TargetID   int64     `gorm:"not null;index:idx_assoc_target,composite:target" json:"target_id"`
	LinkType   string    `gorm:"size:50;not null;default:related" json:"link_type"`
	CreatedBy  int64     `gorm:"not null" json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}
