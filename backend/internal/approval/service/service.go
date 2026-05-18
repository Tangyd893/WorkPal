package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/approval/model"
)

type Repository interface {
	ListTemplates(ctx context.Context, projectID *int64) ([]*model.ApprovalTemplate, error)
	GetTemplate(ctx context.Context, id int64) (*model.ApprovalTemplate, error)
	CreateTemplate(ctx context.Context, t *model.ApprovalTemplate) error
	ListInstances(ctx context.Context, submitterID *int64, status *string) ([]*model.ApprovalInstance, error)
	GetInstance(ctx context.Context, id int64) (*model.ApprovalInstance, error)
	CreateInstance(ctx context.Context, i *model.ApprovalInstance) error
	UpdateInstance(ctx context.Context, i *model.ApprovalInstance) error
	CreateAction(ctx context.Context, a *model.ApprovalAction) error
	ListActions(ctx context.Context, instanceID int64) ([]*model.ApprovalAction, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

type TemplateDTO struct {
	ID             int64  `json:"id"`
	ProjectID      *int64 `json:"project_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	FormSchema     string `json:"form_schema"`
	FlowDefinition string `json:"flow_definition"`
	IsActive       bool   `json:"is_active"`
}

type TemplateInput struct {
	ProjectID      *int64 `json:"project_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	FormSchema     string `json:"form_schema"`
	FlowDefinition string `json:"flow_definition"`
}

type InstanceDTO struct {
	ID            int64          `json:"id"`
	TemplateID    int64          `json:"template_id"`
	Title         string         `json:"title"`
	FormData      string         `json:"form_data"`
	Status        string         `json:"status"`
	SubmitterID   int64          `json:"submitter_id"`
	CurrentNodeID string         `json:"current_node_id"`
	SubmittedAt   string         `json:"submitted_at"`
	Actions       []ActionDTO    `json:"actions,omitempty"`
}

type InstanceInput struct {
	TemplateID int64  `json:"template_id"`
	Title      string `json:"title"`
	FormData   string `json:"form_data"`
}

type ActionInput struct {
	Action  string `json:"action"`
	Comment string `json:"comment"`
}

type ActionDTO struct {
	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	Action    string `json:"action"`
	Comment   string `json:"comment"`
	UserID    int64  `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

func (s *Service) ListTemplates(ctx context.Context, projectID *int64) ([]TemplateDTO, error) {
	ts, err := s.repo.ListTemplates(ctx, projectID)
	if err != nil { return nil, err }
	out := make([]TemplateDTO, len(ts))
	for i, t := range ts {
		out[i] = TemplateDTO{ID: t.ID, ProjectID: t.ProjectID, Name: t.Name, Description: t.Description, FormSchema: t.FormSchema, FlowDefinition: t.FlowDefinition, IsActive: t.IsActive}
	}
	return out, nil
}

func (s *Service) CreateTemplate(ctx context.Context, input TemplateInput) (TemplateDTO, error) {
	t := &model.ApprovalTemplate{ProjectID: input.ProjectID, Name: input.Name, Description: input.Description, FormSchema: input.FormSchema, FlowDefinition: input.FlowDefinition}
	if err := s.repo.CreateTemplate(ctx, t); err != nil { return TemplateDTO{}, err }
	return TemplateDTO{ID: t.ID, ProjectID: t.ProjectID, Name: t.Name, Description: t.Description, FormSchema: t.FormSchema, FlowDefinition: t.FlowDefinition, IsActive: t.IsActive}, nil
}

func (s *Service) ListInstances(ctx context.Context, submitterID *int64, status *string) ([]InstanceDTO, error) {
	insts, err := s.repo.ListInstances(ctx, submitterID, status)
	if err != nil { return nil, err }
	out := make([]InstanceDTO, len(insts))
	for i, inst := range insts {
		out[i] = InstanceDTO{ID: inst.ID, TemplateID: inst.TemplateID, Title: inst.Title, FormData: inst.FormData, Status: inst.Status, SubmitterID: inst.SubmitterID, CurrentNodeID: inst.CurrentNodeID, SubmittedAt: inst.SubmittedAt.Format(time.RFC3339)}
	}
	return out, nil
}

func (s *Service) GetInstance(ctx context.Context, id int64) (InstanceDTO, error) {
	inst, err := s.repo.GetInstance(ctx, id)
	if err != nil { return InstanceDTO{}, err }
	dto := InstanceDTO{ID: inst.ID, TemplateID: inst.TemplateID, Title: inst.Title, FormData: inst.FormData, Status: inst.Status, SubmitterID: inst.SubmitterID, CurrentNodeID: inst.CurrentNodeID, SubmittedAt: inst.SubmittedAt.Format(time.RFC3339)}
	acts, _ := s.repo.ListActions(ctx, id)
	for _, a := range acts {
		dto.Actions = append(dto.Actions, ActionDTO{ID: a.ID, NodeID: a.NodeID, Action: a.Action, Comment: a.Comment, UserID: a.UserID, CreatedAt: a.CreatedAt.Format(time.RFC3339)})
	}
	return dto, nil
}

func (s *Service) CreateInstance(ctx context.Context, submitterID int64, input InstanceInput) (InstanceDTO, error) {
	_, err := s.repo.GetTemplate(ctx, input.TemplateID)
	if err != nil { return InstanceDTO{}, fmt.Errorf("template not found") }
	inst := &model.ApprovalInstance{
		TemplateID: input.TemplateID, Title: input.Title, FormData: input.FormData,
		Status: "pending", SubmitterID: submitterID, CurrentNodeID: "start",
		SubmittedAt: time.Now(),
	}
	if err := s.repo.CreateInstance(ctx, inst); err != nil { return InstanceDTO{}, err }
	return s.GetInstance(ctx, inst.ID)
}

func (s *Service) ProcessAction(ctx context.Context, instanceID int64, userID int64, input ActionInput) (InstanceDTO, error) {
	inst, err := s.repo.GetInstance(ctx, instanceID)
	if err != nil { return InstanceDTO{}, err }
	act := &model.ApprovalAction{InstanceID: instanceID, NodeID: inst.CurrentNodeID, Action: input.Action, Comment: input.Comment, UserID: userID}
	_ = s.repo.CreateAction(ctx, act)
	switch input.Action {
	case "approve":
		inst.Status = "approved"
		now := time.Now()
		inst.CompletedAt = &now
	case "reject":
		inst.Status = "rejected"
		now := time.Now()
		inst.CompletedAt = &now
	}
	if err := s.repo.UpdateInstance(ctx, inst); err != nil { return InstanceDTO{}, err }
	return s.GetInstance(ctx, instanceID)
}
