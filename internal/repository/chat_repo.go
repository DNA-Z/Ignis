package repository

import (
	"context"
	"errors"

	"github.com/DNA-Z/Ignis/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository interface {
	GenericRepository[models.Chat]
	GetUserChatIDs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]uuid.UUID, int64, error)
	GetPrivateChat(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Chat, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Chat, error)
}

type chatRepository struct {
	GenericRepository[models.Chat]
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{
		GenericRepository: NewGenericRepository[models.Chat](db),
		db:                db,
	}
}

func (r *chatRepository) GetUserChatIDs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]uuid.UUID, int64, error) {
	var chatIDs []uuid.UUID
	var total int64

	query := r.db.WithContext(ctx).
		Table("chat_participants").
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Select("chat_id").
		Limit(limit).
		Offset(offset).
		Pluck("chat_id", &chatIDs).Error

	return chatIDs, total, err
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
		return nil, ErrNotFound
	}
	return &chat, err
}

func (r *chatRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Chat, error) {
	var chats []*models.Chat
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Where("deleted_at IS NULL").
		Find(&chats).Error
	return chats, err
}
