package service

import (
	"context"
	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/DNA-Z/Ignis/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SearchService interface {
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, chatID *uuid.UUID, limit, offset int) ([]*models.MessageResponse, int64, error)
	SearchChats(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*models.ChatResponse, int64, error)
}

type searchService struct {
	db              *gorm.DB
	userRepo        repository.UserRepository
	chatRepo        repository.ChatRepository
	messageRepo     repository.MessageRepository
	participantRepo repository.ParticipantRepository
}

func NewSearchService(
	db *gorm.DB,
	userRepo repository.UserRepository,
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	participantRepo repository.ParticipantRepository,
) SearchService {
	return &searchService{
		db:              db,
		userRepo:        userRepo,
		chatRepo:        chatRepo,
		messageRepo:     messageRepo,
		participantRepo: participantRepo,
	}
}

func (s *searchService) SearchMessages(ctx context.Context, userID uuid.UUID, query string, chatID *uuid.UUID, limit, offset int) ([]*models.MessageResponse, int64, error) {
	// Получаем ID чатов пользователя
	chatIDs, _, err := s.chatRepo.GetUserChatIDs(ctx, userID, 1000, 0)
	if err != nil {
		return nil, 0, err
	}

	if len(chatIDs) == 0 {
		return []*models.MessageResponse{}, 0, nil
	}

	// Строим запрос на поиск сообщений
	messageQuery := s.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("chat_id IN (?)", chatIDs).
		Where("is_deleted = false")

	if chatID != nil {
		messageQuery = messageQuery.Where("chat_id = ?", *chatID)
	}

	if query != "" {
		messageQuery = messageQuery.Where("text ILIKE ?", "%"+query+"%")
	}

	var total int64
	if err := messageQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var messages []*models.Message
	if err := messageQuery.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	responses := make([]*models.MessageResponse, len(messages))
	for i, msg := range messages {
		user, err := s.userRepo.GetByID(ctx, msg.UserID)
		if err != nil {
			continue
		}

		responses[i] = &models.MessageResponse{
			ID:     msg.ID,
			ChatID: msg.ChatID,
			User: &models.UserResponse{
				ID:        user.ID,
				Login:     user.Login,
				Name:      user.Name,
				CreatedAt: user.CreatedAt,
			},
			Text:      msg.Text,
			IsDeleted: msg.IsDeleted,
			CreatedAt: msg.CreatedAt,
			EditedAt:  msg.EditedAt,
		}
	}

	return responses, total, nil
}

func (s *searchService) SearchChats(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*models.ChatResponse, int64, error) {
	chatIDs, total, err := s.chatRepo.GetUserChatIDs(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	if len(chatIDs) == 0 {
		return []*models.ChatResponse{}, total, nil
	}

	// Поиск среди чатов пользователя
	var chats []*models.Chat

	searchQuery := s.db.WithContext(ctx).
		Where("id IN ?", chatIDs).
		Where("deleted_at IS NULL")

	if query != "" {
		searchQuery = searchQuery.Where("name ILIKE ?", "%"+query+"%")
	}

	if err := searchQuery.
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&chats).Error; err != nil {
		return nil, 0, err
	}

	responses := make([]*models.ChatResponse, len(chats))
	for i, chat := range chats {
		participants, _ := s.participantRepo.GetParticipants(ctx, chat.ID)
		participantResponses := make([]models.UserResponse, len(participants))
		for j, p := range participants {
			participantResponses[j] = models.UserResponse{
				ID:        p.ID,
				Login:     p.Login,
				Name:      p.Name,
				CreatedAt: p.CreatedAt,
			}
		}

		responses[i] = &models.ChatResponse{
			ID:           chat.ID,
			Type:         chat.Type,
			Name:         chat.Name,
			Description:  chat.Description,
			CreatedBy:    chat.CreatedBy,
			CreatedAt:    chat.CreatedAt,
			UpdatedAt:    chat.UpdatedAt,
			Participants: participantResponses,
		}
	}

	return responses, total, nil
}
