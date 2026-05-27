package service

import (
	"context"
	"errors"
	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/DNA-Z/Ignis/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrNotParticipant    = errors.New("user is not a participant of this chat")
	ErrNotAdmin          = errors.New("user is not an admin of this chat")
	ErrInvalidChatType   = errors.New("invalid chat type")
	ErrPrivateChatExists = errors.New("private chat already exists")
)

type ChatService interface {
	CreateChat(ctx context.Context, userID uuid.UUID, req *models.CreateChatRequest) (*models.ChatResponse, error)
	GetChat(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatResponse, error)
	GetUserChats(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.ChatResponse, int64, error)
	UpdateChat(ctx context.Context, userID, chatID uuid.UUID, req *models.UpdateChatRequest) error
	DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error
	AddParticipants(ctx context.Context, userID, chatID uuid.UUID, req *models.AddParticipantRequest) error
	RemoveParticipant(ctx context.Context, userID, chatID, targetUserID uuid.UUID) error
	UpdateRole(ctx context.Context, userID, chatID, targetUserID uuid.UUID, req *models.UpdateRoleRequest) error
}

type chatService struct {
	chatRepo        repository.ChatRepository
	participantRepo repository.ParticipantRepository
	userRepo        repository.UserRepository
}

func NewChatService(
	chatRepo repository.ChatRepository,
	participantRepo repository.ParticipantRepository,
	userRepo repository.UserRepository,
) ChatService {
	return &chatService{
		chatRepo:        chatRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
	}
}

func (s *chatService) CreateChat(ctx context.Context, userID uuid.UUID, req *models.CreateChatRequest) (*models.ChatResponse, error) {
	// Для private чата проверяем, существует ли уже чат между этими пользователями
	if req.Type == models.PrivateChat {
		if len(req.UserIDs) != 1 {
			return nil, errors.New("private chat must have exactly one other user")
		}

		existingChat, err := s.chatRepo.GetPrivateChat(ctx, userID, req.UserIDs[0])
		if err == nil && existingChat != nil {
			return s.buildChatResponse(ctx, existingChat, userID)
		}
	}

	chat := &models.Chat{
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	// Добавляем создателя как участника
	allUserIDs := append([]uuid.UUID{userID}, req.UserIDs...)
	if err := s.participantRepo.AddParticipants(ctx, chat.ID, allUserIDs); err != nil {
		return nil, err
	}

	return s.buildChatResponse(ctx, chat, userID)
}

func (s *chatService) GetChat(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatResponse, error) {
	// Проверяем, является ли пользователь участником
	isParticipant, err := s.participantRepo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, ErrChatNotFound
	}

	return s.buildChatResponse(ctx, chat, userID)
}

func (s *chatService) GetUserChats(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.ChatResponse, int64, error) {
	chats, total, err := s.chatRepo.GetUserChats(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.ChatResponse, len(chats))
	for i, chat := range chats {
		resp, err := s.buildChatResponse(ctx, chat, userID)
		if err != nil {
			continue
		}
		responses[i] = resp
	}

	return responses, total, nil
}

func (s *chatService) UpdateChat(ctx context.Context, userID, chatID uuid.UUID, req *models.UpdateChatRequest) error {
	// Проверяем права (только создатель или админ)
	role, err := s.participantRepo.GetUserRole(ctx, chatID, userID)
	if err != nil {
		return err
	}

	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return ErrChatNotFound
	}

	// Только создатель или админ могут обновлять
	if chat.CreatedBy != userID && role != models.RoleAdmin {
		return ErrNotAdmin
	}

	if req.Name != nil {
		chat.Name = *req.Name
	}
	if req.Description != nil {
		chat.Description = *req.Description
	}

	return s.chatRepo.Update(ctx, chat)
}

func (s *chatService) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return ErrChatNotFound
	}

	// Только создатель может удалить чат
	if chat.CreatedBy != userID {
		return ErrNotAdmin
	}

	return s.chatRepo.Delete(ctx, chatID)
}

func (s *chatService) AddParticipants(ctx context.Context, userID, chatID uuid.UUID, req *models.AddParticipantRequest) error {
	// Проверяем права
	role, err := s.participantRepo.GetUserRole(ctx, chatID, userID)
	if err != nil {
		return err
	}

	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return ErrChatNotFound
	}

	if chat.CreatedBy != userID && role != models.RoleAdmin {
		return ErrNotAdmin
	}

	return s.participantRepo.AddParticipants(ctx, chatID, req.UserIDs)
}

func (s *chatService) RemoveParticipant(ctx context.Context, userID, chatID, targetUserID uuid.UUID) error {
	role, err := s.participantRepo.GetUserRole(ctx, chatID, userID)
	if err != nil {
		return err
	}

	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return ErrChatNotFound
	}

	// Админ или создатель может исключить, либо пользователь сам себя
	if targetUserID != userID && chat.CreatedBy != userID && role != models.RoleAdmin {
		return ErrNotAdmin
	}

	return s.participantRepo.RemoveParticipant(ctx, chatID, targetUserID)
}

func (s *chatService) UpdateRole(ctx context.Context, userID, chatID, targetUserID uuid.UUID, req *models.UpdateRoleRequest) error {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return ErrChatNotFound
	}

	// Только создатель может назначать админов
	if chat.CreatedBy != userID {
		return ErrNotAdmin
	}

	return s.participantRepo.UpdateRole(ctx, chatID, targetUserID, req.Role)
}

func (s *chatService) buildChatResponse(ctx context.Context, chat *models.Chat, currentUserID uuid.UUID) (*models.ChatResponse, error) {
	participants, err := s.participantRepo.GetParticipants(ctx, chat.ID)
	if err != nil {
		return nil, err
	}

	unreadCount, _ := s.participantRepo.GetUnreadCount(ctx, chat.ID, currentUserID)

	participantResponses := make([]models.UserResponse, len(participants))
	for i, p := range participants {
		participantResponses[i] = models.UserResponse{
			ID:        p.ID,
			Login:     p.Login,
			Name:      p.Name,
			CreatedAt: p.CreatedAt,
		}
	}

	resp := &models.ChatResponse{
		ID:           chat.ID,
		Type:         chat.Type,
		Name:         chat.Name,
		Description:  chat.Description,
		CreatedBy:    chat.CreatedBy,
		CreatedAt:    chat.CreatedAt,
		UpdatedAt:    chat.UpdatedAt,
		Participants: participantResponses,
		UnreadCount:  unreadCount,
	}

	return resp, nil
}
