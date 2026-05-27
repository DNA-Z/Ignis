package service

import (
	"context"
	"errors"
	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/DNA-Z/Ignis/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrMessageNotFound  = errors.New("message not found")
	ErrNotMessageAuthor = errors.New("user is not the author of this message")
	ErrMessageDeleted   = errors.New("message is deleted")
)

type MessageService interface {
	SendMessage(ctx context.Context, userID, chatID uuid.UUID, req *models.SendMessageRequest) (*models.MessageResponse, error)
	GetMessages(ctx context.Context, userID, chatID uuid.UUID, limit, offset int) (*models.MessagesListResponse, error)
	UpdateMessage(ctx context.Context, userID, messageID uuid.UUID, req *models.UpdateMessageRequest) error
	DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) error
	MarkAsRead(ctx context.Context, userID, messageID uuid.UUID) error
}

type messageService struct {
	messageRepo     repository.MessageRepository
	participantRepo repository.ParticipantRepository
	userRepo        repository.UserRepository
	fileRepo        repository.FileRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	participantRepo repository.ParticipantRepository,
	userRepo repository.UserRepository,
	fileRepo repository.FileRepository,
) MessageService {
	return &messageService{
		messageRepo:     messageRepo,
		participantRepo: participantRepo,
		userRepo:        userRepo,
		fileRepo:        fileRepo,
	}
}

func (s *messageService) SendMessage(ctx context.Context, userID, chatID uuid.UUID, req *models.SendMessageRequest) (*models.MessageResponse, error) {
	// Проверяем, является ли пользователь участником чата
	isParticipant, err := s.participantRepo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	message := &models.Message{
		ChatID: chatID,
		UserID: userID,
		Text:   req.Text,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return s.buildMessageResponse(ctx, message)
}

func (s *messageService) GetMessages(ctx context.Context, userID, chatID uuid.UUID, limit, offset int) (*models.MessagesListResponse, error) {
	// Проверяем доступ
	isParticipant, err := s.participantRepo.IsParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	messages, total, err := s.messageRepo.GetChatMessages(ctx, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]models.MessageResponse, len(messages))
	for i, msg := range messages {
		resp, err := s.buildMessageResponse(ctx, msg)
		if err != nil {
			continue
		}
		responses[i] = *resp
	}

	return &models.MessagesListResponse{
		Messages: responses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *messageService) UpdateMessage(ctx context.Context, userID, messageID uuid.UUID, req *models.UpdateMessageRequest) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return ErrMessageNotFound
	}

	if message.UserID != userID {
		return ErrNotMessageAuthor
	}

	if message.IsDeleted {
		return ErrMessageDeleted
	}

	message.Text = req.Text
	return s.messageRepo.Update(ctx, message)
}

func (s *messageService) DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return ErrMessageNotFound
	}

	if message.UserID != userID {
		return ErrNotMessageAuthor
	}

	return s.messageRepo.Delete(ctx, messageID)
}

func (s *messageService) MarkAsRead(ctx context.Context, userID, messageID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return ErrMessageNotFound
	}

	// Проверяем, что пользователь участник чата
	isParticipant, err := s.participantRepo.IsParticipant(ctx, message.ChatID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return ErrNotParticipant
	}

	// Нельзя отметить своё сообщение как прочитанное
	if message.UserID == userID {
		return nil
	}

	// Обновляем last_read_at для пользователя в этом чате
	if err := s.participantRepo.UpdateLastRead(ctx, message.ChatID, userID); err != nil {
		return err
	}

	return s.messageRepo.MarkAsRead(ctx, messageID, userID, message.ChatID)
}

func (s *messageService) buildMessageResponse(ctx context.Context, message *models.Message) (*models.MessageResponse, error) {
	user, err := s.userRepo.FindByID(ctx, message.UserID)
	if err != nil {
		return nil, err
	}

	files, _ := s.fileRepo.GetByMessageID(ctx, message.ID)
	fileResponses := make([]*models.FileResponse, len(files))
	for i, f := range files {
		fileResponses[i] = &models.FileResponse{
			ID:        f.ID,
			Name:      f.Name,
			Size:      f.Size,
			MimeType:  f.MimeType,
			URL:       "/api/files/" + f.ID.String(),
			CreatedAt: f.CreatedAt,
		}
	}

	text := message.Text
	if message.IsDeleted {
		text = "[deleted]"
	}

	return &models.MessageResponse{
		ID:     message.ID,
		ChatID: message.ChatID,
		User: &models.UserResponse{
			ID:        user.ID,
			Login:     user.Login,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		Text:      text,
		IsDeleted: message.IsDeleted,
		CreatedAt: message.CreatedAt,
		EditedAt:  message.EditedAt,
		Files:     fileResponses,
	}, nil
}
