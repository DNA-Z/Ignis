package repository

import (
	"context"

	"github.com/DNA-Z/Ignis/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository interface {
	GenericRepository[models.Message]
	GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, int64, error)
	GetLastMessage(ctx context.Context, chatID uuid.UUID) (*models.Message, error)
	MarkAsRead(ctx context.Context, messageID, userID, chatID uuid.UUID) error
}

type messageRepository struct {
	GenericRepository[models.Message]
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{
		GenericRepository: NewGenericRepository[models.Message](db),
		db:                db,
	}
}

func (r *messageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, int64, error) {
	return r.FindAll(ctx,
		WithCondition("chat_id = ?", chatID),
		WithCondition("deleted_at IS NULL"),
		WithOrderBy("created_at DESC"),
		WithPagination(limit, offset),
	)
}

func (r *messageRepository) GetLastMessage(ctx context.Context, chatID uuid.UUID) (*models.Message, error) {
	messages, _, err := r.FindAll(ctx,
		WithCondition("chat_id = ?", chatID),
		WithCondition("deleted_at IS NULL"),
		WithOrderBy("created_at DESC"),
		WithPagination(1, 0),
	)
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, ErrNotFound
	}
	return messages[0], nil
}

func (r *messageRepository) MarkAsRead(ctx context.Context, messageID, userID, chatID uuid.UUID) error {
	receipt := &models.ReadReceipt{
		MessageID: messageID,
		UserID:    userID,
		ChatID:    chatID,
	}

	return r.db.WithContext(ctx).
		Clauses().
		Create(receipt).Error
}
