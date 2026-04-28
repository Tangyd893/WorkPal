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
	Total    int64            `json:"total"`
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
	if convID == 0 {
		return s.GlobalSearch(ctx, query, page, pageSize)
	}
	return s.SearchInConvs(ctx, []int64{convID}, query, page, pageSize)
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
	searchRequest.Fields = []string{"*"}
	searchRequest.SortBy([]string{"-created_at"})

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var messages []*model.Message
	for _, hit := range result.Hits {
		doc := docFromFields(hit.ID, hit.Fields)
		messages = append(messages, &model.Message{
			ID:        parseInt64(doc.ID),
			ConvID:    doc.ConvID,
			SenderID:  doc.SenderID,
			Content:   doc.Content,
			Type:      doc.Type,
			CreatedAt: parseTime(doc.CreatedAt),
		})
	}

	return &SearchResult{Messages: messages, Total: int64(result.Total)}, nil
}

func (s *SearchService) SearchInConvs(ctx context.Context, convIDs []int64, query string, page, pageSize int) (*SearchResult, error) {
	if len(convIDs) == 0 {
		return &SearchResult{Messages: []*model.Message{}, Total: 0}, nil
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	allowed := make(map[int64]struct{}, len(convIDs))
	for _, convID := range convIDs {
		allowed[convID] = struct{}{}
	}

	fetchSize := page * pageSize * 3
	if fetchSize < pageSize {
		fetchSize = pageSize
	}
	if fetchSize > 1000 {
		fetchSize = 1000
	}
	result, err := s.GlobalSearch(ctx, query, 1, fetchSize)
	if err != nil {
		return nil, err
	}

	filtered := make([]*model.Message, 0, len(result.Messages))
	for _, msg := range result.Messages {
		if _, ok := allowed[msg.ConvID]; ok {
			filtered = append(filtered, msg)
		}
	}

	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return &SearchResult{Messages: []*model.Message{}, Total: int64(len(filtered))}, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return &SearchResult{Messages: filtered[start:end], Total: int64(len(filtered))}, nil
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

func (s *SearchService) DeleteMessage(messageID int64) error {
	if s.index == nil {
		return nil
	}
	return s.index.Delete(fmt.Sprintf("%d", messageID))
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

func docFromFields(id string, fields map[string]interface{}) MessageDoc {
	if src, ok := fields["_source"]; ok {
		var doc MessageDoc
		switch v := src.(type) {
		case []byte:
			if json.Unmarshal(v, &doc) == nil {
				return doc
			}
		case string:
			if json.Unmarshal([]byte(v), &doc) == nil {
				return doc
			}
		}
	}
	return MessageDoc{
		ID:        fieldString(fields, id, "id", "ID"),
		ConvID:    fieldInt64(fields, "conv_id", "ConvID"),
		SenderID:  fieldInt64(fields, "sender_id", "SenderID"),
		Content:   fieldString(fields, "", "content", "Content"),
		Type:      int8(fieldInt64(fields, "type", "Type")),
		CreatedAt: fieldString(fields, "", "created_at", "CreatedAt"),
	}
}

func fieldString(fields map[string]interface{}, fallback string, names ...string) string {
	for _, name := range names {
		if v, ok := fields[name]; ok {
			switch value := v.(type) {
			case string:
				return value
			case fmt.Stringer:
				return value.String()
			}
		}
	}
	return fallback
}

func fieldInt64(fields map[string]interface{}, names ...string) int64 {
	for _, name := range names {
		if v, ok := fields[name]; ok {
			switch value := v.(type) {
			case int64:
				return value
			case int:
				return int64(value)
			case float64:
				return int64(value)
			case json.Number:
				n, _ := value.Int64()
				return n
			case string:
				return parseInt64(value)
			}
		}
	}
	return 0
}
