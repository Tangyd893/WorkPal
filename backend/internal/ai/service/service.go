package service

import (
	"context"
	"fmt"
	"strings"
)

type SearchClient interface {
	SearchMessages(ctx context.Context, query string, convID *int64, page, pageSize int) ([]MessageResult, error)
	SearchIssues(ctx context.Context, query string, projectID *int64) ([]IssueResult, error)
}

type MessageResult struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
	ConvID  int64  `json:"conv_id"`
}

type IssueResult struct {
	ID      int64  `json:"id"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
	Key     string `json:"key"`
}

type Service struct {
	search SearchClient
}

func NewService(search SearchClient) *Service {
	return &Service{search: search}
}

type SmartSearchResult struct {
	Query    string         `json:"query"`
	Messages []MessageResult `json:"messages,omitempty"`
	Issues   []IssueResult   `json:"issues,omitempty"`
	Summary  string          `json:"summary"`
}

type TaskSummaryInput struct {
	IssueIDs []int64 `json:"issue_ids"`
	Titles   []string `json:"titles"`
}

type TaskSummaryResult struct {
	Summary string `json:"summary"`
}

func (s *Service) SmartSearch(ctx context.Context, query string, projectID *int64) (*SmartSearchResult, error) {
	result := &SmartSearchResult{Query: query}
	issues, _ := s.search.SearchIssues(ctx, query, projectID)
	messages, _ := s.search.SearchMessages(ctx, query, nil, 1, 5)
	result.Issues = issues
	result.Messages = messages
	result.Summary = fmt.Sprintf("搜索 \"%s\" 返回 %d 个事项和 %d 条消息", query, len(issues), len(messages))
	return result, nil
}

func (s *Service) SummarizeTasks(ctx context.Context, input TaskSummaryInput) (*TaskSummaryResult, error) {
	var parts []string
	parts = append(parts, fmt.Sprintf("共 %d 个事项", len(input.Titles)))
	for i, t := range input.Titles {
		parts = append(parts, fmt.Sprintf("%d. %s", i+1, t))
	}
	return &TaskSummaryResult{Summary: strings.Join(parts, "\n")}, nil
}
