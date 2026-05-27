package repository

import (
	"context"
	"time"

	"github.com/DNA-Z/Ignis/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, int64, error)
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
	MarkAsRead(ctx context.Context, messageID, userID, chatID uuid.UUID) error
	IsMessageReadByUser(ctx context.Context, messageID, userID uuid.UUID) (bool, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&message).Error

	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*models.Message, int64, error) {
	var messages []*models.Message
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("chat_id = ? AND deleted_at IS NULL", chatID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	now := time.Now()
	message.EditedAt = &now
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *messageRepository) MarkAsRead(ctx context.Context, messageID, userID, chatID uuid.UUID) error {
	receipt := &models.ReadReceipt{
		MessageID: messageID,
		UserID:    userID,
		ChatID:    chatID,
	}

	return r.db.WithContext(ctx).
		Clauses(). // OnConflict можно добавить при необходимости
		Create(receipt).Error
}

func (r *messageRepository) IsMessageReadByUser(ctx context.Context, messageID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ReadReceipt{}).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		Count(&count).Error

	return count > 0, err
}
