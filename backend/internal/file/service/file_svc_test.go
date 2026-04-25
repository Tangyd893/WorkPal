package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockFileStore 内存文件存储（测试用）
type mockFileStore struct {
	mu    sync.RWMutex
	files map[string]string // key -> content
}

func newMockFileStore() *mockFileStore {
	return &mockFileStore{files: make(map[string]string)}
}

func (m *mockFileStore) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.files[key] = string(data)
	m.mu.Unlock()
	return nil
}

func (m *mockFileStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if content, ok := m.files[key]; ok {
		return io.NopCloser(strings.NewReader(content)), nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFileStore) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	delete(m.files, key)
	m.mu.Unlock()
	return nil
}

func (m *mockFileStore) GetURL(ctx context.Context, key string) (string, error) {
	return "/files/" + key, nil
}

// === LocalFileStore 单元测试 ===

func TestLocalFileStore_Upload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "localstore-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewLocalFileStore(tmpDir)

	t.Run("成功上传", func(t *testing.T) {
		err := store.Upload(context.Background(), "uploads/user1/file.txt", strings.NewReader("hello world"), 11, "text/plain")
		assert.NoError(t, err)

		// 验证文件存在
		path := filepath.Join(tmpDir, "uploads/user1/file.txt")
		data, err := os.ReadFile(path)
		assert.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
	})

	t.Run("上传目录自动创建", func(t *testing.T) {
		err := store.Upload(context.Background(), "deep/nested/path/file.txt", strings.NewReader("content"), 7, "text/plain")
		assert.NoError(t, err)
	})
}

func TestLocalFileStore_Download(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "localstore-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewLocalFileStore(tmpDir)

	// 先写入文件
	subDir := filepath.Join(tmpDir, "downloads")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "test.txt"), []byte("download content"), 0644)

	t.Run("成功下载", func(t *testing.T) {
		rc, err := store.Download(context.Background(), "downloads/test.txt")
		assert.NoError(t, err)
		data, _ := io.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "download content", string(data))
	})

	t.Run("文件不存在", func(t *testing.T) {
		_, err := store.Download(context.Background(), "nonexistent/file.txt")
		assert.Error(t, err)
	})
}

func TestLocalFileStore_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "localstore-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewLocalFileStore(tmpDir)

	// 先写入文件
	os.MkdirAll(filepath.Join(tmpDir, "todelete"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "todelete/file.txt"), []byte("to delete"), 0644)

	err = store.Delete(context.Background(), "todelete/file.txt")
	assert.NoError(t, err)

	// 验证文件已删除
	_, err = os.ReadFile(filepath.Join(tmpDir, "todelete/file.txt"))
	assert.True(t, os.IsNotExist(err))
}

func TestLocalFileStore_GetURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "localstore-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewLocalFileStore(tmpDir)

	url, err := store.GetURL(context.Background(), "uploads/image.png")
	assert.NoError(t, err)
	assert.Equal(t, "/files/uploads/image.png", url)
}

// === mockFileStore 单元测试 ===

func TestMockFileStore(t *testing.T) {
	store := newMockFileStore()

	t.Run("上传和下载", func(t *testing.T) {
		err := store.Upload(context.Background(), "key1", strings.NewReader("test data"), 9, "text/plain")
		assert.NoError(t, err)

		rc, err := store.Download(context.Background(), "key1")
		assert.NoError(t, err)
		data, _ := io.ReadAll(rc)
		rc.Close()
		assert.Equal(t, "test data", string(data))
	})

	t.Run("下载不存在的文件", func(t *testing.T) {
		_, err := store.Download(context.Background(), "nonexistent")
		assert.Error(t, err)
	})

	t.Run("删除", func(t *testing.T) {
		store.Upload(context.Background(), "tokill", strings.NewReader("data"), 4, "text/plain")
		err := store.Delete(context.Background(), "tokill")
		assert.NoError(t, err)

		_, err = store.Download(context.Background(), "tokill")
		assert.Error(t, err)
	})

	t.Run("GetURL", func(t *testing.T) {
		url, err := store.GetURL(context.Background(), "any/key")
		assert.NoError(t, err)
		assert.Equal(t, "/files/any/key", url)
	})
}