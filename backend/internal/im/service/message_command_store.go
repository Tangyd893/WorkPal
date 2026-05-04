package service

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/im/model"
	"github.com/Tangyd893/WorkPal/backend/internal/im/repo"
	"gorm.io/gorm"
)

type GormMessageCommandStore struct {
	db         *gorm.DB
	msgRepo    *repo.MessageRepo
	outboxRepo *repo.OutboxRepo
}

func NewGormMessageCommandStore(db *gorm.DB, msgRepo *repo.MessageRepo, outboxRepo *repo.OutboxRepo) *GormMessageCommandStore {
	return &GormMessageCommandStore{
		db:         db,
		msgRepo:    msgRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *GormMessageCommandStore) WithinTx(ctx context.Context, fn func(MessageTxStore) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&gormMessageTxStore{
			tx:         tx,
			msgRepo:    s.msgRepo,
			outboxRepo: s.outboxRepo,
		})
	})
}

type gormMessageTxStore struct {
	tx         *gorm.DB
	msgRepo    *repo.MessageRepo
	outboxRepo *repo.OutboxRepo
}

func (s *gormMessageTxStore) CreateMessage(msg *model.Message) error {
	return s.msgRepo.CreateWithTx(s.tx, msg)
}

func (s *gormMessageTxStore) GetMessage(id int64) (*model.Message, error) {
	return s.msgRepo.GetByIDWithTx(s.tx, id)
}

func (s *gormMessageTxStore) UpdateMessage(msg *model.Message) error {
	return s.msgRepo.UpdateWithTx(s.tx, msg)
}

func (s *gormMessageTxStore) SoftDeleteMessage(msgID int64) error {
	return s.msgRepo.SoftDeleteWithTx(s.tx, msgID)
}

func (s *gormMessageTxStore) EnqueueOutbox(event *model.MessageOutbox) error {
	return s.outboxRepo.CreateWithTx(s.tx, event)
}
