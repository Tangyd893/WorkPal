package engine

import (
	"fmt"

	"github.com/Tangyd893/WorkPal/backend/internal/project/model"
)

// Engine 工作流引擎，负责校验状态转换规则
type Engine struct{}

// NewEngine 创建工作流引擎实例
func NewEngine() *Engine {
	return &Engine{}
}

// ValidateTransition 校验 issue 从当前状态到目标状态的转换是否合法
// 返回匹配到的转换规则，如果转换不被允许则返回 error
func (e *Engine) ValidateTransition(dsl *model.WorkflowDSL, issue *model.Issue, userID int64, targetStatus string) (*model.Transition, error) {
	if dsl == nil || len(dsl.Transitions) == 0 {
		return nil, fmt.Errorf("工作流未定义任何转换规则")
	}

	fromStatus := issue.Status
	for i := range dsl.Transitions {
		t := &dsl.Transitions[i]
		if t.From == fromStatus && t.To == targetStatus {
			if err := e.evaluateConditions(t.Conditions, issue); err != nil {
				return nil, fmt.Errorf("条件不满足: %w", err)
			}
			if err := e.executeValidators(t.Validators, issue, userID); err != nil {
				return nil, fmt.Errorf("校验不通过: %w", err)
			}
			return t, nil
		}
	}

	return nil, fmt.Errorf("状态转换 %s → %s 不被允许", fromStatus, targetStatus)
}

// GetAvailableStatuses 返回当前状态下所有可转换的目标状态列表
func (e *Engine) GetAvailableStatuses(dsl *model.WorkflowDSL, currentStatus string) []string {
	if dsl == nil {
		return nil
	}
	targets := make([]string, 0)
	seen := make(map[string]bool)
	for _, t := range dsl.Transitions {
		if t.From == currentStatus && !seen[t.To] {
			targets = append(targets, t.To)
			seen[t.To] = true
		}
	}
	return targets
}

// GetStatuses 返回工作流中所有状态列表
func (e *Engine) GetStatuses(dsl *model.WorkflowDSL) []string {
	if dsl == nil {
		return nil
	}
	return dsl.Statuses
}

func (e *Engine) evaluateConditions(conditions []model.Condition, issue *model.Issue) error {
	for _, cond := range conditions {
		if err := e.evaluateCondition(cond, issue); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) evaluateCondition(cond model.Condition, issue *model.Issue) error {
	switch cond.Field {
	case "assignee":
		return e.checkAssigneeCondition(cond, issue)
	case "priority":
		return e.checkStringFieldCondition(cond, issue.Priority)
	case "resolution":
		return e.checkStringFieldCondition(cond, issue.Resolution)
	default:
		return fmt.Errorf("未知的条件字段: %s", cond.Field)
	}
}

func (e *Engine) checkAssigneeCondition(cond model.Condition, issue *model.Issue) error {
	switch cond.Operator {
	case "not_null":
		if issue.AssigneeID == nil {
			return fmt.Errorf("指派人不能为空")
		}
		return nil
	case "is_null":
		if issue.AssigneeID != nil {
			return fmt.Errorf("指派人必须为空")
		}
		return nil
	default:
		return fmt.Errorf("不支持的指派人条件运算符: %s", cond.Operator)
	}
}

func (e *Engine) checkStringFieldCondition(cond model.Condition, fieldValue string) error {
	switch cond.Operator {
	case "eq":
		if fieldValue != cond.Value {
			return fmt.Errorf("字段值 %s 不等于 %s", fieldValue, cond.Value)
		}
		return nil
	case "neq":
		if fieldValue == cond.Value {
			return fmt.Errorf("字段值不能等于 %s", cond.Value)
		}
		return nil
	default:
		return fmt.Errorf("不支持的条件运算符: %s", cond.Operator)
	}
}

func (e *Engine) executeValidators(validators []model.ValidatorDef, issue *model.Issue, userID int64) error {
	for _, v := range validators {
		if err := e.executeValidator(v, issue, userID); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) executeValidator(v model.ValidatorDef, issue *model.Issue, userID int64) error {
	switch v.Class {
	case "PermissionValidator":
		return e.executePermissionValidator(v, issue, userID)
	default:
		return fmt.Errorf("未知的校验器类型: %s", v.Class)
	}
}

func (e *Engine) executePermissionValidator(v model.ValidatorDef, issue *model.Issue, userID int64) error {
	role, _ := v.Args["role"].(string)
	switch role {
	case "developer":
		if issue.AssigneeID != nil && *issue.AssigneeID == userID {
			return nil
		}
		if issue.ReporterID == userID {
			return nil
		}
		return fmt.Errorf("只有事项负责人或指派人才能执行此操作")
	case "reporter":
		if issue.ReporterID == userID {
			return nil
		}
		return fmt.Errorf("只有事项报告人才能执行此操作")
	case "":
		return nil
	default:
		return fmt.Errorf("权限校验器不支持的的角色: %s", role)
	}
}

// HasPostFunctions 检查转换规则是否包含后处理函数
func HasPostFunctions(t *model.Transition) bool {
	return len(t.PostFunctions) > 0
}

// DefaultWorkflowDSL 返回默认的工作流 DSL
func DefaultWorkflowDSL() *model.WorkflowDSL {
	return &model.WorkflowDSL{
		Statuses: []string{"Open", "In Progress", "In Review", "Done"},
		Transitions: []model.Transition{
			{From: "Open", To: "In Progress"},
			{From: "Open", To: "In Review"},
			{From: "In Progress", To: "In Review"},
			{From: "In Progress", To: "Done"},
			{From: "In Review", To: "Done"},
			{From: "In Review", To: "In Progress"},
			{From: "Done", To: "Open"},
		},
	}
}
