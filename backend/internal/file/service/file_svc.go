package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/Tangyd893/WorkPal/backend/internal/file/model"
	"github.com/Tangyd893/WorkPal/backend/internal/file/repo"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v6"
)

type FileStore interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (string, error)
}

type LocalFileStore struct {
	BasePath string
}

func NewLocalFileStore(basePath string) *LocalFileStore {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		panic("failed to create base directory: " + err.Error())
	}
	return &LocalFileStore{BasePath: basePath}
}

func (s *LocalFileStore) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	fullPath := filepath.Join(s.BasePath, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *LocalFileStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.BasePath, key))
}

func (s *LocalFileStore) Delete(ctx context.Context, key string) error {
	return os.Remove(filepath.Join(s.BasePath, key))
}

func (s *LocalFileStore) GetURL(ctx context.Context, key string) (string, error) {
	return "/files/" + key, nil
}

type MinIOFileStore struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

func NewMinIOFileStore(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOFileStore, error) {
	client, err := minio.New(endpoint, accessKey, secretKey, useSSL)
	if err != nil {
		return nil, err
	}
	return &MinIOFileStore{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

func (s *MinIOFileStore) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinIOFileStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *MinIOFileStore) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(s.bucket, key)
}

func (s *MinIOFileStore) GetURL(ctx context.Context, key string) (string, error) {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucket, key), nil
}

type FileService struct {
	repo      *repo.FileRepo
	store     FileStore
	maxSizeMB int
}

func NewFileService(repo *repo.FileRepo, store FileStore, maxSizeMB int) *FileService {
	return &FileService{
		repo:      repo,
		store:     store,
		maxSizeMB: maxSizeMB,
	}
}

func (s *FileService) Upload(ctx context.Context, userID int64, convID int64, fileHeader *multipart.FileHeader) (*model.File, error) {
	if int(fileHeader.Size) > s.maxSizeMB*1024*1024 {
		return nil, fmt.Errorf("文件大小超过限制 (%dMB)", s.maxSizeMB)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	ext := filepath.Ext(fileHeader.Filename)
	key := fmt.Sprintf("%d/%d/%s%s", userID, convID, uuid.New().String(), ext)

	if err := s.store.Upload(ctx, key, src, fileHeader.Size, fileHeader.Header.Get("Content-Type")); err != nil {
		return nil, err
	}

	file := &model.File{
		UserID:      userID,
		ConvID:      convID,
		Name:        fileHeader.Filename,
		Key:         key,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
		MimeType:    fileHeader.Header.Get("Content-Type"),
		CreatedAt:   time.Now(),
	}
	if err := s.repo.Create(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileService) GetURL(ctx context.Context, fileID int64) (string, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return "", err
	}
	return s.store.GetURL(ctx, file.Key)
}

func (s *FileService) GetByID(ctx context.Context, fileID int64) (*model.File, error) {
	return s.repo.GetByID(ctx, fileID)
}

func (s *FileService) Download(ctx context.Context, fileID int64) (io.ReadCloser, *model.File, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}
	reader, err := s.store.Download(ctx, file.Key)
	return reader, file, err
}

func (s *FileService) Delete(ctx context.Context, fileID int64) (*model.File, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Delete(ctx, fileID); err != nil {
		return nil, err
	}

	if err := s.store.Delete(ctx, file.Key); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return file, nil
}

func (s *FileService) ListByUser(ctx context.Context, userID int64, offset, limit int) ([]*model.File, error) {
	return s.repo.ListByUser(ctx, userID, offset, limit)
}

func (s *FileService) ListByConv(ctx context.Context, convID int64, offset, limit int) ([]*model.File, error) {
	return s.repo.ListByConv(ctx, convID, offset, limit)
}
