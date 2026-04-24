package service

import (
	"context"
	"strings"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"gorm.io/gorm"
)

// SearchService 搜索服务（基于 PostgreSQL ILIKE）
type SearchService struct {
	db *gorm.DB
}

func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{db: db}
}

// SearchResult 搜索结果
type SearchResult struct {
	Messages []*model.Message `json:"messages"`
	Total    int64            `json:"total"`
}

// IndexMessage 索引消息（PostgreSQL 无需显式索引，消息创建时自动可用）
func (s *SearchService) IndexMessage(msg *model.Message) error {
	// PostgreSQL 的 ILIKE 会自动使用 messages.content 上的索引（如果有）
	// 消息创建时会自动被搜索到，无需额外索引操作
	return nil
}

// IndexMessages 批量索引
func (s *SearchService) IndexMessages(msgs []*model.Message) error {
	return nil
}

// SearchInConv 会话内搜索
func (s *SearchService) SearchInConv(ctx context.Context, convID int64, query string, page, pageSize int) (*SearchResult, error) {
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}

	offset := (page - 1) * pageSize
	query = strings.TrimSpace(query)
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}

	var messages []*model.Message
	var total int64

	// ILIKE 搜索（大小写不敏感）
	searchPattern := "%" + query + "%"

	db := s.db.WithContext(ctx).Model(&model.Message{}).
		Where("conv_id = ? AND deleted_at IS NULL AND content ILIKE ?", convID, searchPattern)

	db.Count(&total)

	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return &SearchResult{Messages: messages, Total: total}, nil
}

// GlobalSearch 全局搜索
func (s *SearchService) GlobalSearch(ctx context.Context, query string, page, pageSize int) (*SearchResult, error) {
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}

	offset := (page - 1) * pageSize
	query = strings.TrimSpace(query)
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}

	var messages []*model.Message
	var total int64

	searchPattern := "%" + query + "%"

	db := s.db.WithContext(ctx).Model(&model.Message{}).
		Where("deleted_at IS NULL AND content ILIKE ?", searchPattern)

	db.Count(&total)

	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return &SearchResult{Messages: messages, Total: total}, nil
}

// Search 搜索（兼容接口）
func (s *SearchService) Search(ctx context.Context, query string, page, pageSize int) (*SearchResult, error) {
	return s.GlobalSearch(ctx, query, page, pageSize)
}

// Close 关闭（无资源需要释放）
func (s *SearchService) Close() error {
	return nil
}
