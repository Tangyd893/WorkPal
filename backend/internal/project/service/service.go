package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/project/engine"
	"github.com/Tangyd893/WorkPal/backend/internal/project/model"
)

var ErrNotFound = errors.New("item not found")

type Repository interface {
	ListProjects(ctx context.Context) ([]*model.Project, error)
	CreateProject(ctx context.Context, project *model.Project) error
	GetProject(ctx context.Context, projectID int64) (*model.Project, error)
	SaveProject(ctx context.Context, project *model.Project) error
	DeleteProject(ctx context.Context, projectID int64) error
	NextIssueKeySeq(ctx context.Context, projectID int64) (int64, error)
	ListIssues(ctx context.Context, projectID int64) ([]*model.Issue, error)
	CreateIssue(ctx context.Context, issue *model.Issue) error
	GetIssue(ctx context.Context, issueID int64) (*model.Issue, error)
	SaveIssue(ctx context.Context, issue *model.Issue) error
	DeleteIssue(ctx context.Context, issueID int64) error
	CreateChangelog(ctx context.Context, log *model.IssueChangelog) error
	ListChangelogs(ctx context.Context, issueID int64) ([]*model.IssueChangelog, error)
	CreateAssociation(ctx context.Context, assoc *model.Association) error
	ListAssociations(ctx context.Context, sourceType string, sourceID int64) ([]*model.Association, error)
	ListIssueTypes(ctx context.Context, projectID int64) ([]*model.IssueType, error)
	ListWorkflows(ctx context.Context, projectID int64) ([]*model.Workflow, error)
	GetWorkflow(ctx context.Context, workflowID int64) (*model.Workflow, error)
	CreateWorkflow(ctx context.Context, workflow *model.Workflow) error
	UpdateWorkflow(ctx context.Context, workflow *model.Workflow) error
	DeleteWorkflow(ctx context.Context, workflowID int64) error
	GetActiveWorkflow(ctx context.Context, projectID int64) (*model.Workflow, error)
	ListCustomFieldDefs(ctx context.Context, projectID int64) ([]*model.CustomFieldDef, error)
	CreateCustomFieldDef(ctx context.Context, def *model.CustomFieldDef) error
	UpdateCustomFieldDef(ctx context.Context, def *model.CustomFieldDef) error
	DeleteCustomFieldDef(ctx context.Context, fieldID int64) error
	ListCustomFieldValues(ctx context.Context, issueID int64) ([]*model.CustomFieldValue, error)
	UpsertCustomFieldValue(ctx context.Context, val *model.CustomFieldValue) error
	DeleteCustomFieldValues(ctx context.Context, issueID int64) error
}

type Service struct {
	repo   Repository
	engine *engine.Engine
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:   repo,
		engine: engine.NewEngine(),
	}
}

type ProjectDTO struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LeadID      int64  `json:"lead_id"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
	IsArchived  bool   `json:"is_archived"`
}

type ProjectInput struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LeadID      int64  `json:"lead_id"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
}

type IssueDTO struct {
	ID            string                 `json:"id"`
	ProjectID     int64                  `json:"project_id"`
	IssueTypeID   int64                  `json:"issue_type_id"`
	IssueTypeName string                 `json:"issue_type_name"`
	ParentID      *int64                 `json:"parent_id"`
	Key           string                 `json:"key"`
	Summary       string                 `json:"summary"`
	Description   string                 `json:"description"`
	Status        string                 `json:"status"`
	Priority      string                 `json:"priority"`
	AssigneeID    *int64                 `json:"assignee_id"`
	ReporterID    int64                  `json:"reporter_id"`
	DueDate       *string                `json:"due_date"`
	StoryPoints   *float64               `json:"story_points"`
	Resolution    string                 `json:"resolution"`
	VersionIDs    []int64                `json:"version_ids"`
	FixVersionIDs []int64                `json:"fix_version_ids"`
	TimeEstimate  int                    `json:"time_estimate"`
	TimeSpent     int                    `json:"time_spent"`
	SortOrder     int                    `json:"sort_order"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	Changelogs    []IssueChangelogDTO    `json:"changelogs,omitempty"`
}

type IssueInput struct {
	ProjectID     int64   `json:"project_id"`
	IssueTypeID   int64   `json:"issue_type_id"`
	ParentID      *int64  `json:"parent_id"`
	Summary       string  `json:"summary"`
	Description   string  `json:"description"`
	Priority      string  `json:"priority"`
	AssigneeID    *int64  `json:"assignee_id"`
	ReporterID    int64   `json:"reporter_id"`
	DueDate       *string `json:"due_date"`
	StoryPoints   *float64 `json:"story_points"`
	VersionIDs    []int64 `json:"version_ids"`
	TimeEstimate  int     `json:"time_estimate"`
}

type IssueStatusInput struct {
	Status string `json:"status"`
}

type IssueChangelogDTO struct {
	ID        int64  `json:"id"`
	Field     string `json:"field"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
	ChangedBy int64  `json:"changed_by"`
	CreatedAt string `json:"created_at"`
}

type IssueTypeDTO struct {
	ID             int64  `json:"id"`
	ProjectID      int64  `json:"project_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	IconURL        string `json:"icon_url"`
	HierarchyLevel int    `json:"hierarchy_level"`
	IsStandard     bool   `json:"is_standard"`
}

type WorkflowDTO struct {
	ID            string             `json:"id"`
	ProjectID     int64              `json:"project_id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	IsActive      bool               `json:"is_active"`
	DSLDefinition *model.WorkflowDSL `json:"dsl_definition"`
	CreatedAt     string             `json:"created_at"`
}

type WorkflowInput struct {
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	IsActive      *bool              `json:"is_active"`
	DSLDefinition *model.WorkflowDSL `json:"dsl_definition"`
}

type AvailableStatusesResult struct {
	CurrentStatus string   `json:"current_status"`
	Statuses      []string `json:"statuses"`
}

func (s *Service) ListProjects(ctx context.Context) ([]ProjectDTO, error) {
	projects, err := s.repo.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectDTO, 0, len(projects))
	for _, p := range projects {
		out = append(out, projectDTO(p))
	}
	return out, nil
}

func (s *Service) CreateProject(ctx context.Context, input ProjectInput) (ProjectDTO, error) {
	project := &model.Project{
		Key:         input.Key,
		Name:        input.Name,
		Description: input.Description,
		LeadID:      input.LeadID,
		Icon:        fallback(input.Icon, "folder"),
		Category:    fallback(input.Category, "software"),
	}
	if err := s.repo.CreateProject(ctx, project); err != nil {
		return ProjectDTO{}, err
	}
	s.seedDefaultIssueTypes(ctx, project.ID)
	s.seedDefaultWorkflow(ctx, project.ID)
	return projectDTO(project), nil
}

func (s *Service) GetProject(ctx context.Context, projectID int64) (ProjectDTO, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return ProjectDTO{}, ErrNotFound
	}
	return projectDTO(project), nil
}

func (s *Service) DeleteProject(ctx context.Context, projectID int64) error {
	return s.repo.DeleteProject(ctx, projectID)
}

func (s *Service) ListIssues(ctx context.Context, projectID int64) ([]IssueDTO, error) {
	issues, err := s.repo.ListIssues(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]IssueDTO, 0, len(issues))
	for _, issue := range issues {
		issueTypes, _ := s.repo.ListIssueTypes(ctx, projectID)
		out = append(out, issueDTO(issue, issueTypes))
	}
	return out, nil
}

func (s *Service) CreateIssue(ctx context.Context, input IssueInput) (IssueDTO, error) {
	seq, err := s.repo.NextIssueKeySeq(ctx, input.ProjectID)
	if err != nil {
		return IssueDTO{}, fmt.Errorf("generate issue key: %w", err)
	}
	project, err := s.repo.GetProject(ctx, input.ProjectID)
	if err != nil {
		return IssueDTO{}, ErrNotFound
	}
	issueKey := fmt.Sprintf("%s-%d", project.Key, seq)
	issue := &model.Issue{
		ProjectID:     input.ProjectID,
		IssueTypeID:   input.IssueTypeID,
		ParentID:      input.ParentID,
		Key:           issueKey,
		Summary:       input.Summary,
		Description:   input.Description,
		Status:        "Open",
		Priority:      fallback(input.Priority, "Medium"),
		AssigneeID:    input.AssigneeID,
		ReporterID:    input.ReporterID,
		DueDate:       input.DueDate,
		StoryPoints:   input.StoryPoints,
		VersionIDs:    marshalInt64Slice(input.VersionIDs),
		TimeEstimate:  input.TimeEstimate,
	}
	if err := s.repo.CreateIssue(ctx, issue); err != nil {
		return IssueDTO{}, err
	}
	s.recordChangelog(ctx, issue.ID, "创建", "", issueKey, input.ReporterID)
	issueTypes, _ := s.repo.ListIssueTypes(ctx, input.ProjectID)
	return issueDTO(issue, issueTypes), nil
}

func (s *Service) GetIssue(ctx context.Context, issueID int64) (IssueDTO, error) {
	issue, err := s.repo.GetIssue(ctx, issueID)
	if err != nil {
		return IssueDTO{}, ErrNotFound
	}
	issueTypes, _ := s.repo.ListIssueTypes(ctx, issue.ProjectID)
	changelogs, _ := s.repo.ListChangelogs(ctx, issueID)
	dto := issueDTO(issue, issueTypes)
	dto.Changelogs = make([]IssueChangelogDTO, 0, len(changelogs))
	for _, log := range changelogs {
		dto.Changelogs = append(dto.Changelogs, IssueChangelogDTO{
			ID:        log.ID,
			Field:     log.Field,
			OldValue:  log.OldValue,
			NewValue:  log.NewValue,
			ChangedBy: log.ChangedBy,
			CreatedAt: log.CreatedAt.Format(time.RFC3339),
		})
	}
	return dto, nil
}

func (s *Service) UpdateIssueStatus(ctx context.Context, issueID int64, userID int64, targetStatus string) (IssueDTO, error) {
	issue, err := s.repo.GetIssue(ctx, issueID)
	if err != nil {
		return IssueDTO{}, ErrNotFound
	}
	oldStatus := issue.Status
	if oldStatus == targetStatus {
		issueTypes, _ := s.repo.ListIssueTypes(ctx, issue.ProjectID)
		return issueDTO(issue, issueTypes), nil
	}

	workflow, err := s.repo.GetActiveWorkflow(ctx, issue.ProjectID)
	if err == nil && workflow != nil {
		dsl, parseErr := workflow.ParseDSL()
		if parseErr != nil {
			return IssueDTO{}, fmt.Errorf("解析工作流定义失败: %w", parseErr)
		}
		transition, validateErr := s.engine.ValidateTransition(dsl, issue, userID, targetStatus)
		if validateErr != nil {
			return IssueDTO{}, validateErr
		}
		s.executePostFunctions(ctx, transition, issue, userID, oldStatus, targetStatus)
	}

	issue.Status = targetStatus
	s.recordChangelog(ctx, issueID, "状态", oldStatus, targetStatus, userID)
	if err := s.repo.SaveIssue(ctx, issue); err != nil {
		return IssueDTO{}, err
	}
	issueTypes, _ := s.repo.ListIssueTypes(ctx, issue.ProjectID)
	return issueDTO(issue, issueTypes), nil
}

func (s *Service) UpdateIssue(ctx context.Context, issueID int64, userID int64, input IssueInput) (IssueDTO, error) {
	issue, err := s.repo.GetIssue(ctx, issueID)
	if err != nil {
		return IssueDTO{}, ErrNotFound
	}
	if input.Summary != "" && input.Summary != issue.Summary {
		s.recordChangelog(ctx, issueID, "摘要", issue.Summary, input.Summary, userID)
		issue.Summary = input.Summary
	}
	if input.Priority != "" && input.Priority != issue.Priority {
		s.recordChangelog(ctx, issueID, "优先级", issue.Priority, input.Priority, userID)
		issue.Priority = input.Priority
	}
	if input.AssigneeID != nil && (issue.AssigneeID == nil || *input.AssigneeID != *issue.AssigneeID) {
		old := ""
		if issue.AssigneeID != nil {
			old = strconv.FormatInt(*issue.AssigneeID, 10)
		}
		s.recordChangelog(ctx, issueID, "指派人", old, strconv.FormatInt(*input.AssigneeID, 10), userID)
		issue.AssigneeID = input.AssigneeID
	}
	if input.Description != "" && input.Description != issue.Description {
		issue.Description = input.Description
	}
	if len(input.VersionIDs) > 0 {
		issue.VersionIDs = marshalInt64Slice(input.VersionIDs)
	}
	if input.TimeEstimate > 0 {
		issue.TimeEstimate = input.TimeEstimate
	}
	if input.DueDate != nil {
		issue.DueDate = input.DueDate
	}
	if input.StoryPoints != nil {
		issue.StoryPoints = input.StoryPoints
	}
	if err := s.repo.SaveIssue(ctx, issue); err != nil {
		return IssueDTO{}, err
	}
	issueTypes, _ := s.repo.ListIssueTypes(ctx, issue.ProjectID)
	return issueDTO(issue, issueTypes), nil
}

func (s *Service) DeleteIssue(ctx context.Context, issueID int64) error {
	return s.repo.DeleteIssue(ctx, issueID)
}

func (s *Service) ListChangelogs(ctx context.Context, issueID int64) ([]IssueChangelogDTO, error) {
	logs, err := s.repo.ListChangelogs(ctx, issueID)
	if err != nil {
		return nil, err
	}
	out := make([]IssueChangelogDTO, 0, len(logs))
	for _, log := range logs {
		out = append(out, IssueChangelogDTO{
			ID:        log.ID,
			Field:     log.Field,
			OldValue:  log.OldValue,
			NewValue:  log.NewValue,
			ChangedBy: log.ChangedBy,
			CreatedAt: log.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *Service) CreateAssociation(ctx context.Context, sourceType string, sourceID int64, targetType string, targetID int64, linkType string, userID int64) error {
	assoc := &model.Association{
		SourceType: sourceType,
		SourceID:   sourceID,
		TargetType: targetType,
		TargetID:   targetID,
		LinkType:   fallback(linkType, "related"),
		CreatedBy:  userID,
	}
	return s.repo.CreateAssociation(ctx, assoc)
}

func (s *Service) ListAssociations(ctx context.Context, sourceType string, sourceID int64) ([]struct {
	ID         int64  `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   int64  `json:"target_id"`
	LinkType   string `json:"link_type"`
}, error) {
	assocs, err := s.repo.ListAssociations(ctx, sourceType, sourceID)
	if err != nil {
		return nil, err
	}
	out := make([]struct {
		ID         int64  `json:"id"`
		TargetType string `json:"target_type"`
		TargetID   int64  `json:"target_id"`
		LinkType   string `json:"link_type"`
	}, 0, len(assocs))
	for _, a := range assocs {
		out = append(out, struct {
			ID         int64  `json:"id"`
			TargetType string `json:"target_type"`
			TargetID   int64  `json:"target_id"`
			LinkType   string `json:"link_type"`
		}{a.ID, a.TargetType, a.TargetID, a.LinkType})
	}
	return out, nil
}

func (s *Service) ListIssueTypes(ctx context.Context, projectID int64) ([]IssueTypeDTO, error) {
	types, err := s.repo.ListIssueTypes(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]IssueTypeDTO, 0, len(types))
	for _, t := range types {
		out = append(out, IssueTypeDTO{
			ID:             t.ID,
			ProjectID:      t.ProjectID,
			Name:           t.Name,
			Description:    t.Description,
			IconURL:        t.IconURL,
			HierarchyLevel: t.HierarchyLevel,
			IsStandard:     t.IsStandard,
		})
	}
	return out, nil
}

func (s *Service) ListWorkflows(ctx context.Context, projectID int64) ([]WorkflowDTO, error) {
	workflows, err := s.repo.ListWorkflows(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]WorkflowDTO, 0, len(workflows))
	for _, w := range workflows {
		dsl, _ := w.ParseDSL()
		out = append(out, WorkflowDTO{
			ID:            formatWFID(w.ID),
			ProjectID:     w.ProjectID,
			Name:          w.Name,
			Description:   w.Description,
			IsActive:      w.IsActive,
			DSLDefinition: dsl,
			CreatedAt:     w.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *Service) CreateWorkflow(ctx context.Context, projectID int64, input WorkflowInput) (WorkflowDTO, error) {
	dsl := input.DSLDefinition
	if dsl == nil {
		dsl = engine.DefaultWorkflowDSL()
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	workflow := &model.Workflow{
		ProjectID:   projectID,
		Name:        fallback(input.Name, "默认工作流"),
		Description: input.Description,
		IsActive:    isActive,
	}
	if err := workflow.SetDSL(dsl); err != nil {
		return WorkflowDTO{}, fmt.Errorf("序列化工作流定义失败: %w", err)
	}
	if err := s.repo.CreateWorkflow(ctx, workflow); err != nil {
		return WorkflowDTO{}, err
	}
	return WorkflowDTO{
		ID:            formatWFID(workflow.ID),
		ProjectID:     workflow.ProjectID,
		Name:          workflow.Name,
		Description:   workflow.Description,
		IsActive:      workflow.IsActive,
		DSLDefinition: dsl,
		CreatedAt:     workflow.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) GetWorkflow(ctx context.Context, workflowID int64) (WorkflowDTO, error) {
	workflow, err := s.repo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return WorkflowDTO{}, ErrNotFound
	}
	dsl, _ := workflow.ParseDSL()
	return WorkflowDTO{
		ID:            formatWFID(workflow.ID),
		ProjectID:     workflow.ProjectID,
		Name:          workflow.Name,
		Description:   workflow.Description,
		IsActive:      workflow.IsActive,
		DSLDefinition: dsl,
		CreatedAt:     workflow.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) UpdateWorkflow(ctx context.Context, workflowID int64, input WorkflowInput) (WorkflowDTO, error) {
	workflow, err := s.repo.GetWorkflow(ctx, workflowID)
	if err != nil {
		return WorkflowDTO{}, ErrNotFound
	}
	if input.Name != "" {
		workflow.Name = input.Name
	}
	if input.Description != "" {
		workflow.Description = input.Description
	}
	if input.IsActive != nil {
		workflow.IsActive = *input.IsActive
	}
	if input.DSLDefinition != nil {
		if err := workflow.SetDSL(input.DSLDefinition); err != nil {
			return WorkflowDTO{}, fmt.Errorf("序列化工作流定义失败: %w", err)
		}
	}
	if err := s.repo.UpdateWorkflow(ctx, workflow); err != nil {
		return WorkflowDTO{}, err
	}
	dsl, _ := workflow.ParseDSL()
	return WorkflowDTO{
		ID:            formatWFID(workflow.ID),
		ProjectID:     workflow.ProjectID,
		Name:          workflow.Name,
		Description:   workflow.Description,
		IsActive:      workflow.IsActive,
		DSLDefinition: dsl,
		CreatedAt:     workflow.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) DeleteWorkflow(ctx context.Context, workflowID int64) error {
	return s.repo.DeleteWorkflow(ctx, workflowID)
}

func (s *Service) GetAvailableStatuses(ctx context.Context, issueID int64) (AvailableStatusesResult, error) {
	issue, err := s.repo.GetIssue(ctx, issueID)
	if err != nil {
		return AvailableStatusesResult{}, ErrNotFound
	}
	result := AvailableStatusesResult{CurrentStatus: issue.Status}
	workflow, err := s.repo.GetActiveWorkflow(ctx, issue.ProjectID)
	if err != nil || workflow == nil {
		result.Statuses = s.engine.GetAvailableStatuses(engine.DefaultWorkflowDSL(), issue.Status)
		return result, nil
	}
	dsl, parseErr := workflow.ParseDSL()
	if parseErr != nil {
		result.Statuses = s.engine.GetAvailableStatuses(engine.DefaultWorkflowDSL(), issue.Status)
		return result, nil
	}
	if len(dsl.Transitions) == 0 {
		result.Statuses = s.engine.GetAvailableStatuses(engine.DefaultWorkflowDSL(), issue.Status)
		return result, nil
	}
	result.Statuses = s.engine.GetAvailableStatuses(dsl, issue.Status)
	return result, nil
}

func (s *Service) seedDefaultWorkflow(ctx context.Context, projectID int64) {
	dsl := engine.DefaultWorkflowDSL()
	workflow := &model.Workflow{
		ProjectID: projectID,
		Name:      "默认工作流",
		IsActive:  true,
	}
	if err := workflow.SetDSL(dsl); err != nil {
		return
	}
	_ = s.repo.CreateWorkflow(ctx, workflow)
}

func (s *Service) executePostFunctions(ctx context.Context, transition *model.Transition, issue *model.Issue, userID int64, oldStatus, newStatus string) {
	for _, pf := range transition.PostFunctions {
		switch pf.Class {
		case "UpdateHistory":
			s.recordChangelog(ctx, issue.ID, "状态", oldStatus, newStatus, userID)
		case "SendNotification":
			// 预留通知事件发送接口
		}
	}
}

func (s *Service) recordChangelog(ctx context.Context, issueID int64, field, oldValue, newValue string, changedBy int64) {
	logEntry := &model.IssueChangelog{
		IssueID:   issueID,
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		ChangedBy: changedBy,
	}
	_ = s.repo.CreateChangelog(ctx, logEntry)
}

func (s *Service) seedDefaultIssueTypes(ctx context.Context, projectID int64) {
	defaults := []struct {
		name           string
		hierarchyLevel int
	}{
		{"Epic", 0},
		{"Story", 1},
		{"Task", 2},
		{"Sub-task", 3},
		{"Bug", 2},
	}
	for _, d := range defaults {
		et := &model.IssueType{
			ProjectID:      projectID,
			Name:           d.name,
			HierarchyLevel: d.hierarchyLevel,
			IsStandard:     true,
		}
		_ = s.repo.(interface{ CreateIssueType(ctx context.Context, t *model.IssueType) error }).(interface{ CreateIssueType(ctx context.Context, t *model.IssueType) error })
		_ = et
	}
}

type CustomFieldDefDTO struct {
	ID        int64    `json:"id"`
	ProjectID int64    `json:"project_id"`
	Name      string   `json:"name"`
	FieldType string   `json:"field_type"`
	Options   []string `json:"options"`
	IsRequired bool    `json:"is_required"`
	SortOrder int      `json:"sort_order"`
}

type CustomFieldDefInput struct {
	Name      string   `json:"name"`
	FieldType string   `json:"field_type"`
	Options   []string `json:"options"`
	IsRequired bool    `json:"is_required"`
	SortOrder int      `json:"sort_order"`
}

type CustomFieldValueDTO struct {
	ID          int64   `json:"id"`
	IssueID     int64   `json:"issue_id"`
	FieldID     int64   `json:"field_id"`
	FieldName   string  `json:"field_name"`
	FieldType   string  `json:"field_type"`
	ValueText   string  `json:"value_text"`
	ValueNumber float64 `json:"value_number"`
	ValueDate   *string `json:"value_date"`
}

type CustomFieldValueInput struct {
	FieldID     int64   `json:"field_id"`
	ValueText   string  `json:"value_text"`
	ValueNumber float64 `json:"value_number"`
	ValueDate   *string `json:"value_date"`
}

func (s *Service) ListCustomFieldDefs(ctx context.Context, projectID int64) ([]CustomFieldDefDTO, error) {
	defs, err := s.repo.ListCustomFieldDefs(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]CustomFieldDefDTO, 0, len(defs))
	for _, d := range defs {
		opts, _ := d.ParseOptions()
		out = append(out, CustomFieldDefDTO{
			ID:         d.ID,
			ProjectID:  d.ProjectID,
			Name:       d.Name,
			FieldType:  d.FieldType,
			Options:    opts,
			IsRequired: d.IsRequired,
			SortOrder:  d.SortOrder,
		})
	}
	return out, nil
}

func (s *Service) CreateCustomFieldDef(ctx context.Context, projectID int64, input CustomFieldDefInput) (CustomFieldDefDTO, error) {
	def := &model.CustomFieldDef{
		ProjectID:  projectID,
		Name:       input.Name,
		FieldType:  fallback(input.FieldType, "text"),
		IsRequired: input.IsRequired,
		SortOrder:  input.SortOrder,
	}
	if len(input.Options) > 0 {
		_ = def.SetOptions(input.Options)
	}
	if err := s.repo.CreateCustomFieldDef(ctx, def); err != nil {
		return CustomFieldDefDTO{}, err
	}
	opts, _ := def.ParseOptions()
	return CustomFieldDefDTO{
		ID:         def.ID,
		ProjectID:  def.ProjectID,
		Name:       def.Name,
		FieldType:  def.FieldType,
		Options:    opts,
		IsRequired: def.IsRequired,
		SortOrder:  def.SortOrder,
	}, nil
}

func (s *Service) UpdateCustomFieldDef(ctx context.Context, fieldID int64, input CustomFieldDefInput) (CustomFieldDefDTO, error) {
	defs, _ := s.repo.ListCustomFieldDefs(ctx, 0)
	var def *model.CustomFieldDef
	for _, d := range defs {
		if d.ID == fieldID {
			def = d
			break
		}
	}
	if def == nil {
		return CustomFieldDefDTO{}, ErrNotFound
	}
	if input.Name != "" {
		def.Name = input.Name
	}
	if input.FieldType != "" {
		def.FieldType = input.FieldType
	}
	def.IsRequired = input.IsRequired
	if input.SortOrder != 0 {
		def.SortOrder = input.SortOrder
	}
	if len(input.Options) > 0 {
		_ = def.SetOptions(input.Options)
	}
	if err := s.repo.UpdateCustomFieldDef(ctx, def); err != nil {
		return CustomFieldDefDTO{}, err
	}
	opts, _ := def.ParseOptions()
	return CustomFieldDefDTO{
		ID:         def.ID,
		ProjectID:  def.ProjectID,
		Name:       def.Name,
		FieldType:  def.FieldType,
		Options:    opts,
		IsRequired: def.IsRequired,
		SortOrder:  def.SortOrder,
	}, nil
}

func (s *Service) DeleteCustomFieldDef(ctx context.Context, fieldID int64) error {
	return s.repo.DeleteCustomFieldDef(ctx, fieldID)
}

func (s *Service) GetIssueCustomFieldValues(ctx context.Context, issueID int64) ([]CustomFieldValueDTO, error) {
	vals, err := s.repo.ListCustomFieldValues(ctx, issueID)
	if err != nil {
		return nil, err
	}
	issue, err := s.repo.GetIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	defs, _ := s.repo.ListCustomFieldDefs(ctx, issue.ProjectID)
	defMap := make(map[int64]*model.CustomFieldDef)
	for _, d := range defs {
		defMap[d.ID] = d
	}
	out := make([]CustomFieldValueDTO, 0, len(vals))
	for _, v := range vals {
		dto := CustomFieldValueDTO{
			ID:          v.ID,
			IssueID:     v.IssueID,
			FieldID:     v.FieldID,
			ValueText:   v.ValueText,
			ValueNumber: v.ValueNumber,
			ValueDate:   v.ValueDate,
		}
		if def, ok := defMap[v.FieldID]; ok {
			dto.FieldName = def.Name
			dto.FieldType = def.FieldType
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *Service) UpsertCustomFieldValue(ctx context.Context, issueID int64, input CustomFieldValueInput) error {
	val := &model.CustomFieldValue{
		IssueID:     issueID,
		FieldID:     input.FieldID,
		ValueText:   input.ValueText,
		ValueNumber: input.ValueNumber,
		ValueDate:   input.ValueDate,
	}
	return s.repo.UpsertCustomFieldValue(ctx, val)
}

func projectDTO(p *model.Project) ProjectDTO {
	return ProjectDTO{
		ID:          formatID(p.ID),
		Key:         p.Key,
		Name:        p.Name,
		Description: p.Description,
		LeadID:      p.LeadID,
		Icon:        p.Icon,
		Category:    p.Category,
		IsArchived:  p.IsArchived,
	}
}

func issueDTO(issue *model.Issue, issueTypes []*model.IssueType) IssueDTO {
	typeName := ""
	for _, t := range issueTypes {
		if t.ID == issue.IssueTypeID {
			typeName = t.Name
			break
		}
	}
	return IssueDTO{
		ID:            formatID(issue.ID),
		ProjectID:     issue.ProjectID,
		IssueTypeID:   issue.IssueTypeID,
		IssueTypeName: typeName,
		ParentID:      issue.ParentID,
		Key:           issue.Key,
		Summary:       issue.Summary,
		Description:   issue.Description,
		Status:        issue.Status,
		Priority:      issue.Priority,
		AssigneeID:    issue.AssigneeID,
		ReporterID:    issue.ReporterID,
		DueDate:       issue.DueDate,
		StoryPoints:   issue.StoryPoints,
		Resolution:    issue.Resolution,
		VersionIDs:    unmarshalInt64Slice(issue.VersionIDs),
		FixVersionIDs: unmarshalInt64Slice(issue.FixVersionIDs),
		TimeEstimate:  issue.TimeEstimate,
		TimeSpent:     issue.TimeSpent,
		SortOrder:     issue.SortOrder,
		CreatedAt:     issue.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     issue.UpdatedAt.Format(time.RFC3339),
	}
}

func formatID(id int64) string {
	return "prj-" + strconv.FormatInt(id, 10)
}

func formatWFID(id int64) string {
	return "wf-" + strconv.FormatInt(id, 10)
}

func ParseWFID(id string) (int64, error) {
	id = strings.TrimPrefix(id, "wf-")
	return strconv.ParseInt(id, 10, 64)
}

func ParseID(id string) (int64, error) {
	id = strings.TrimPrefix(id, "prj-")
	return strconv.ParseInt(id, 10, 64)
}

func marshalInt64Slice(value []int64) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func unmarshalInt64Slice(value string) []int64 {
	if value == "" {
		return []int64{}
	}
	var out []int64
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return []int64{}
	}
	return out
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
