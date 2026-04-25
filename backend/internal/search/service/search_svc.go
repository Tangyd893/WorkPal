package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/blevesearch/bleve"
)

// MessageDoc Bleve 索引文档
type MessageDoc struct {
	ID        string `json:"id"`
	ConvID    int64  `json:"conv_id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	Type      int8   `json:"type"`
	CreatedAt string `json:"created_at"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Messages []*model.Message `json:"messages"`
	Total    int64           `json:"total"`
}

// SearchService 搜索服务（Bleve 全文索引）
type SearchService struct {
	index bleve.Index
}

// NewSearchService 创建 Bleve 搜索服务
func NewSearchService(indexPath string) (*SearchService, error) {
	// 创建内存索引（开发环境）或个人存储路径
	index, err := bleve.New(indexPath, bleve.NewIndexMapping())
	if err != nil {
		// 索引已存在，尝试打开
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, fmt.Errorf("创建搜索索引失败: %w", err)
		}
	}
	return &SearchService{index: index}, nil
}

// SearchInConv 会话内搜索
func (s *SearchService) SearchInConv(ctx context.Context, convID int64, query string, page, pageSize int) (*SearchResult, error) {
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}

	offset := (page - 1) * pageSize

	// 使用 Bleve MatchQuery
	searchRequest := bleve.NewSearchRequest(bleve.NewMatchQuery(query))
	searchRequest.From = offset
	searchRequest.Size = pageSize
	searchRequest.SortBy([]string{"-CreatedAt"})

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var messages []*model.Message
	for _, hit := range result.Hits {
		var doc MessageDoc
		// 尝试从 _source 字段获取
		if src, ok := hit.Fields["_source"]; ok {
			if data, ok := src.([]byte); ok {
				if err := json.Unmarshal(data, &doc); err != nil {
					continue
				}
			}
		}
		// 过滤当前会话
		if doc.ConvID == convID || convID == 0 {
			messages = append(messages, &model.Message{
				ID:        parseInt64(hit.ID),
				ConvID:    doc.ConvID,
				SenderID:  doc.SenderID,
				Content:   doc.Content,
				Type:      doc.Type,
				CreatedAt: parseTime(doc.CreatedAt),
			})
		}
	}

	return &SearchResult{Messages: messages, Total: int64(result.Total)}, nil
}

// GlobalSearch 全局搜索
func (s *SearchService) GlobalSearch(ctx context.Context, query string, page, pageSize int) (*SearchResult, error) {
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}

	offset := (page - 1) * pageSize

	searchRequest := bleve.NewSearchRequest(bleve.NewMatchQuery(query))
	searchRequest.From = offset
	searchRequest.Size = pageSize
	searchRequest.SortBy([]string{"-CreatedAt"})

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var messages []*model.Message
	for _, hit := range result.Hits {
		var doc MessageDoc
		if src, ok := hit.Fields["_source"]; ok {
			if data, ok := src.([]byte); ok {
				if err := json.Unmarshal(data, &doc); err != nil {
					continue
				}
			}
		}
		messages = append(messages, &model.Message{
			ID:        parseInt64(hit.ID),
			ConvID:    doc.ConvID,
			SenderID:  doc.SenderID,
			Content:   doc.Content,
			Type:      doc.Type,
			CreatedAt: parseTime(doc.CreatedAt),
		})
	}

	return &SearchResult{Messages: messages, Total: int64(result.Total)}, nil
}

// IndexMessage 索引单条消息
func (s *SearchService) IndexMessage(msg *model.Message) error {
	if s.index == nil {
		return nil
	}
	doc := MessageDoc{
		ID:        fmt.Sprintf("%d", msg.ID),
		ConvID:    msg.ConvID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		Type:      msg.Type,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	}
	return s.index.Index(doc.ID, doc)
}

// IndexMessages 批量索引
func (s *SearchService) IndexMessages(msgs []*model.Message) error {
	if s.index == nil {
		return nil
	}
	batch := s.index.NewBatch()
	for _, msg := range msgs {
		doc := MessageDoc{
			ID:        fmt.Sprintf("%d", msg.ID),
			ConvID:    msg.ConvID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			Type:      msg.Type,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		}
		if err := batch.Index(doc.ID, doc); err != nil {
			return err
		}
	}
	return s.index.Batch(batch)
}

// Close 关闭索引
func (s *SearchService) Close() error {
	if s.index != nil {
		return s.index.Close()
	}
	return nil
}

// Search 搜索（兼容）
func (s *SearchService) Search(ctx context.Context, query string, page, pageSize int) (*SearchResult, error) {
	return s.GlobalSearch(ctx, query, page, pageSize)
}

// 辅助函数
func parseInt64(s string) int64 {
	var n int64
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0
	}
	return n
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
