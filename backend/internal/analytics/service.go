package analytics

import (
	"context"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/project/model"
	"gorm.io/gorm"
)

type IssueRepo interface {
	ListIssues(ctx context.Context, projectID int64) ([]*model.Issue, error)
}

type Service struct {
	db  *gorm.DB
	repo IssueRepo
}

func NewService(db *gorm.DB, repo IssueRepo) *Service {
	return &Service{db: db, repo: repo}
}

type BurnDownPoint struct {
	Date  string `json:"date"`
	Open  int64  `json:"open"`
	Done  int64  `json:"done"`
	Total int64  `json:"total"`
}

type BurnDownReport struct {
	Sprint  string           `json:"sprint"`
	Points  []BurnDownPoint  `json:"points"`
	Ideal   []BurnDownPoint  `json:"ideal"`
}

type ThroughputReport struct {
	Period    string           `json:"period"`
	Created   int64            `json:"created"`
	Completed int64            `json:"completed"`
	ByStatus  map[string]int64 `json:"by_status"`
	ByPriority map[string]int64 `json:"by_priority"`
}

type TeamDashboard struct {
	TotalIssues    int64   `json:"total_issues"`
	OpenIssues     int64   `json:"open_issues"`
	InProgressCount int64  `json:"in_progress_count"`
	DoneCount      int64   `json:"done_count"`
	AvgResolutionDays float64 `json:"avg_resolution_days"`
	ProjectCount   int64   `json:"project_count"`
}

func (s *Service) GetThroughput(ctx context.Context, projectID int64, days int) (*ThroughputReport, error) {
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)
	issues, err := s.repo.ListIssues(ctx, projectID)
	if err != nil {
		return nil, err
	}
	rpt := &ThroughputReport{
		Period:    since.Format("2006-01-02") + " ~ " + time.Now().Format("2006-01-02"),
		ByStatus:  make(map[string]int64),
		ByPriority: make(map[string]int64),
	}
	for _, issue := range issues {
		if issue.CreatedAt.After(since) {
			rpt.Created++
		}
		rpt.ByStatus[issue.Status]++
		rpt.ByPriority[issue.Priority]++
		if issue.Status == "Done" && issue.UpdatedAt.After(since) {
			rpt.Completed++
		}
	}
	return rpt, nil
}

func (s *Service) GetTeamDashboard(ctx context.Context, projectID int64) (*TeamDashboard, error) {
	issues, err := s.repo.ListIssues(ctx, projectID)
	if err != nil {
		return nil, err
	}
	d := &TeamDashboard{TotalIssues: int64(len(issues))}
	for _, issue := range issues {
		switch issue.Status {
		case "Open":
			d.OpenIssues++
		case "In Progress":
			d.InProgressCount++
		case "In Review":
			d.InProgressCount++
		case "Done":
			d.DoneCount++
			resolution := issue.UpdatedAt.Sub(issue.CreatedAt).Hours() / 24
			if resolution > 0 {
				d.AvgResolutionDays += resolution
			}
		}
	}
	if d.DoneCount > 0 {
		d.AvgResolutionDays /= float64(d.DoneCount)
	}
	var projectCount int64
	s.db.WithContext(ctx).Model(&model.Project{}).Count(&projectCount)
	d.ProjectCount = projectCount
	return d, nil
}
