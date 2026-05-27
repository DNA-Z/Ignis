package repository

import (
	"context"
	"errors"
	"github.com/DNA-Z/Ignis/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Chat, error)
	GetUserChats(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Chat, int64, error)
	Update(ctx context.Context, chat *models.Chat) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPrivateChat(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Chat, error)
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(ctx context.Context, chat *models.Chat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}

func (r *chatRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&chat).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &chat, err
}

func (r *chatRepository) GetUserChats(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Chat, int64, error) {
	var chats []*models.Chat
	var total int64

	subQuery := r.db.Table("chat_participants").
		Select("chat_id").
		Where("user_id = ?", userID)

	query := r.db.WithContext(ctx).
		Where("id IN (?)", subQuery).
		Where("deleted_at IS NULL")

	if err := query.Model(&models.Chat{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&chats).Error; err != nil {
		return nil, 0, err
	}

	return chats, total, nil
}

func (r *chatRepository) Update(ctx context.Context, chat *models.Chat) error {
	return r.db.WithContext(ctx).Save(chat).Error
}

func (r *chatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Chat{}).Error
}

func (r *chatRepository) GetPrivateChat(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Chat, error) {
	var chat models.Chat

	err := r.db.WithContext(ctx).
		Joins("JOIN chat_participants cp1 ON cp1.chat_id = chats.id").
		Joins("JOIN chat_participants cp2 ON cp2.chat_id = chats.id").
		Where("chats.type = ?", models.PrivateChat).
		Where("chats.deleted_at IS NULL").
		Where("cp1.user_id = ?", userID1).
		Where("cp2.user_id = ?", userID2).
		Group("chats.id").
		Having("COUNT(DISTINCT cp1.user_id) = 2").
		First(&chat).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &chat, err
}
