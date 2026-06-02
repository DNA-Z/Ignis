package repository

import (
	"context"
	"time"

	"github.com/DNA-Z/Ignis/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ParticipantRepository interface {
	AddParticipants(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) error
	RemoveParticipant(ctx context.Context, chatID, userID uuid.UUID) error
	GetParticipants(ctx context.Context, chatID uuid.UUID) ([]*models.User, error)
	UpdateRole(ctx context.Context, chatID, userID uuid.UUID, role models.ParticipantRole) error
	IsParticipant(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
	GetUserRole(ctx context.Context, chatID, userID uuid.UUID) (models.ParticipantRole, error)
	UpdateLastRead(ctx context.Context, chatID, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int64, error)
}

type participantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) ParticipantRepository {
	return &participantRepository{db: db}
}

func (r *participantRepository) AddParticipants(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) error {
	participants := make([]models.ChatParticipant, len(userIDs))
	for i, userID := range userIDs {
		participants[i] = models.ChatParticipant{
			ChatID: chatID,
			UserID: userID,
			Role:   models.RoleMember,
		}
	}
	return r.db.WithContext(ctx).Create(&participants).Error
}

func (r *participantRepository) RemoveParticipant(ctx context.Context, chatID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&models.ChatParticipant{}).Error
}

func (r *participantRepository) GetParticipants(ctx context.Context, chatID uuid.UUID) ([]*models.User, error) {
	var users []*models.User

	err := r.db.WithContext(ctx).
		Table("users").
		Joins("JOIN chat_participants ON chat_participants.user_id = users.id").
		Where("chat_participants.chat_id = ?", chatID).
		Find(&users).Error

	return users, err
}

func (r *participantRepository) UpdateRole(ctx context.Context, chatID, userID uuid.UUID, role models.ParticipantRole) error {
	return r.db.WithContext(ctx).
		Model(&models.ChatParticipant{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("role", role).Error
}

func (r *participantRepository) IsParticipant(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChatParticipant{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error

	return count > 0, err
}

func (r *participantRepository) GetUserRole(ctx context.Context, chatID, userID uuid.UUID) (models.ParticipantRole, error) {
	var participant models.ChatParticipant
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&participant).Error

	return participant.Role, err
}

func (r *participantRepository) UpdateLastRead(ctx context.Context, chatID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ChatParticipant{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("last_read_at", now).Error
}

func (r *participantRepository) GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int64, error) {
	var lastRead models.ChatParticipant
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&lastRead).Error

	if err != nil {
		return 0, err
	}

	var count int64
	query := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("chat_id = ? AND is_deleted = false", chatID)

	if lastRead.LastReadAt != nil {
		query = query.Where("created_at > ?", *lastRead.LastReadAt)
	}

	err = query.Count(&count).Error
	return count, err
}
