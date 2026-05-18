package model

import (
	"encoding/json"
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

// Workflow 工作流定义模型
type Workflow struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID     int64     `gorm:"not null;index" json:"project_id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	DSLDefinition string    `gorm:"type:jsonb;not null;default:'{}'" json:"-"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// WorkflowDSL 工作流 DSL 定义（JSON 序列化结构）
type WorkflowDSL struct {
	Statuses    []string     `json:"statuses"`
	Transitions []Transition `json:"transitions"`
}

// Transition 状态转换规则
type Transition struct {
	From          string           `json:"from"`
	To            string           `json:"to"`
	Conditions    []Condition      `json:"conditions,omitempty"`
	Validators    []ValidatorDef   `json:"validators,omitempty"`
	PostFunctions []PostFunctionDef `json:"post_functions,omitempty"`
}

// Condition 转换条件
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

// ValidatorDef 校验器定义
type ValidatorDef struct {
	Class string                 `json:"class"`
	Args  map[string]interface{} `json:"args,omitempty"`
}

// PostFunctionDef 后处理函数定义
type PostFunctionDef struct {
	Class string                 `json:"class"`
	Args  map[string]interface{} `json:"args,omitempty"`
}

// ParseDSL 解析 JSONB 字段为 WorkflowDSL
func (w *Workflow) ParseDSL() (*WorkflowDSL, error) {
	var dsl WorkflowDSL
	if w.DSLDefinition == "" || w.DSLDefinition == "{}" {
		return &dsl, nil
	}
	if err := json.Unmarshal([]byte(w.DSLDefinition), &dsl); err != nil {
		return nil, err
	}
	return &dsl, nil
}

// SetDSL 将 WorkflowDSL 序列化写入 JSONB 字段
func (w *Workflow) SetDSL(dsl *WorkflowDSL) error {
	data, err := json.Marshal(dsl)
	if err != nil {
		return err
	}
	w.DSLDefinition = string(data)
	return nil
}

// CustomFieldDef 自定义字段定义
type CustomFieldDef struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID  int64     `gorm:"not null;index" json:"project_id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	FieldType  string    `gorm:"size:30;not null;default:text" json:"field_type"`
	Options    string    `gorm:"type:jsonb;default:'[]'" json:"-"`
	IsRequired bool      `gorm:"default:false" json:"is_required"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
}

// ParseOptions 解析 JSONB 选项字段
func (c *CustomFieldDef) ParseOptions() ([]string, error) {
	if c.Options == "" || c.Options == "[]" {
		return nil, nil
	}
	var opts []string
	if err := json.Unmarshal([]byte(c.Options), &opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// SetOptions 设置选项列表
func (c *CustomFieldDef) SetOptions(opts []string) error {
	data, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	c.Options = string(data)
	return nil
}

// CustomFieldValue 自定义字段值
type CustomFieldValue struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID     int64     `gorm:"not null;uniqueIndex:idx_cfv_issue_field" json:"issue_id"`
	FieldID     int64     `gorm:"not null;uniqueIndex:idx_cfv_issue_field" json:"field_id"`
	ValueText   string    `gorm:"type:text" json:"value_text"`
	ValueNumber float64   `json:"value_number"`
	ValueDate   *string   `json:"value_date"`
	ValueJSON   string    `gorm:"type:jsonb;default:'{}'" json:"-"`
}
