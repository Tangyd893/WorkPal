package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/stretchr/testify/assert"
)

func newTestSearchService(t *testing.T) (*SearchService, func()) {
	tmpDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	svc, err := NewSearchService(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("创建 SearchService 失败: %v", err)
	}
	cleanup := func() {
		svc.Close()
		os.RemoveAll(tmpDir)
	}
	return svc, cleanup
}

func TestSearchService_IndexMessage(t *testing.T) {
	svc, cleanup := newTestSearchService(t)
	defer cleanup()

	msg := &model.Message{
		ID:        1,
		ConvID:    100,
		SenderID:  10,
		Type:      1,
		Content:   "Hello World",
		CreatedAt: time.Now(),
	}
	err := svc.IndexMessage(msg)
	assert.NoError(t, err)
}

func TestSearchService_IndexMessages(t *testing.T) {
	svc, cleanup := newTestSearchService(t)
	defer cleanup()

	msgs := []*model.Message{
		{ID: 1, ConvID: 100, SenderID: 10, Type: 1, Content: "Hi there", CreatedAt: time.Now()},
		{ID: 2, ConvID: 100, SenderID: 20, Type: 1, Content: "Good morning", CreatedAt: time.Now()},
		{ID: 3, ConvID: 200, SenderID: 10, Type: 1, Content: "Hello again", CreatedAt: time.Now()},
	}
	err := svc.IndexMessages(msgs)
	assert.NoError(t, err)
}

func TestSearchService_SearchInConv(t *testing.T) {
	svc, cleanup := newTestSearchService(t)
	defer cleanup()

	svc.IndexMessage(&model.Message{ID: 1, ConvID: 100, SenderID: 10, Type: 1, Content: "Hello World", CreatedAt: time.Now()})
	svc.IndexMessage(&model.Message{ID: 2, ConvID: 100, SenderID: 20, Type: 1, Content: "Good morning everyone", CreatedAt: time.Now()})
	svc.IndexMessage(&model.Message{ID: 3, ConvID: 200, SenderID: 30, Type: 1, Content: "Hello from another chat", CreatedAt: time.Now()})

	t.Run("空关键词返回空", func(t *testing.T) {
		res, err := svc.SearchInConv(context.Background(), 100, "", 1, 20)
		assert.NoError(t, err)
		assert.Len(t, res.Messages, 0)
	})

	t.Run("无匹配结果", func(t *testing.T) {
		res, err := svc.SearchInConv(context.Background(), 100, "xyz_not_exist", 1, 20)
		assert.NoError(t, err)
		assert.Len(t, res.Messages, 0)
	})

	t.Run("会话内搜索有结果", func(t *testing.T) {
		res, err := svc.SearchInConv(context.Background(), 100, "Hello", 1, 20)
		assert.NoError(t, err)
		assert.True(t, res.Total >= 1 || len(res.Messages) >= 1)
	})
}

func TestSearchService_GlobalSearch(t *testing.T) {
	svc, cleanup := newTestSearchService(t)
	defer cleanup()

	svc.IndexMessage(&model.Message{ID: 1, ConvID: 100, SenderID: 10, Type: 1, Content: "Hello World", CreatedAt: time.Now()})
	svc.IndexMessage(&model.Message{ID: 2, ConvID: 200, SenderID: 20, Type: 1, Content: "Hello again", CreatedAt: time.Now()})
	svc.IndexMessage(&model.Message{ID: 3, ConvID: 100, SenderID: 30, Type: 1, Content: "Good morning", CreatedAt: time.Now()})

	t.Run("全局搜索 Hello", func(t *testing.T) {
		res, err := svc.GlobalSearch(context.Background(), "Hello", 1, 20)
		assert.NoError(t, err)
		assert.Len(t, res.Messages, 2) // ID=1 和 ID=2
	})

	t.Run("分页搜索", func(t *testing.T) {
		res, err := svc.GlobalSearch(context.Background(), "Hello", 1, 1)
		assert.NoError(t, err)
		assert.Len(t, res.Messages, 1)
		assert.Equal(t, int64(2), res.Total)
	})

	t.Run("空关键词", func(t *testing.T) {
		res, err := svc.GlobalSearch(context.Background(), "", 1, 20)
		assert.NoError(t, err)
		assert.Len(t, res.Messages, 0)
	})
}

func TestSearchService_Search(t *testing.T) {
	svc, cleanup := newTestSearchService(t)
	defer cleanup()

	svc.IndexMessage(&model.Message{ID: 1, ConvID: 100, SenderID: 10, Type: 1, Content: "test content", CreatedAt: time.Now()})

	res, err := svc.Search(context.Background(), "test", 1, 20)
	assert.NoError(t, err)
	assert.Len(t, res.Messages, 1)
}

func TestSearchService_Close(t *testing.T) {
	svc, cleanup := newTestSearchService(t)
	defer cleanup()

	err := svc.Close()
	assert.NoError(t, err)
}