package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/DNA-Z/Ignis/internal/repository"

	"github.com/google/uuid"
)

type FileService interface {
	UploadFile(ctx context.Context, userID, messageID uuid.UUID, fileHeader *multipart.FileHeader) (*models.FileResponse, error)
	GetFile(ctx context.Context, userID, fileID uuid.UUID) (*models.File, io.ReadCloser, error)
	DeleteFile(ctx context.Context, userID, fileID uuid.UUID) error
}

type fileService struct {
	fileRepo        repository.FileRepository
	messageRepo     repository.MessageRepository
	participantRepo repository.ParticipantRepository
	uploadPath      string
	maxFileSize     int64
}

func NewFileService(
	fileRepo repository.FileRepository,
	messageRepo repository.MessageRepository,
	participantRepo repository.ParticipantRepository,
	uploadPath string,
	maxFileSize int64,
) FileService {
	return &fileService{
		fileRepo:        fileRepo,
		messageRepo:     messageRepo,
		participantRepo: participantRepo,
		uploadPath:      uploadPath,
		maxFileSize:     maxFileSize,
	}
}

func (s *fileService) UploadFile(ctx context.Context, userID, messageID uuid.UUID, fileHeader *multipart.FileHeader) (*models.FileResponse, error) {
	// Проверяем размер файла
	if fileHeader.Size > s.maxFileSize {
		return nil, fmt.Errorf("file too large: max %d bytes", s.maxFileSize)
	}

	// Получаем сообщение
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, ErrMessageNotFound
	}

	// Проверяем, что пользователь автор сообщения
	if message.UserID != userID {
		return nil, ErrNotMessageAuthor
	}

	// Проверяем, что сообщение не удалено
	if message.IsDeleted {
		return nil, ErrMessageDeleted
	}

	// Открываем файл
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Создаём уникальное имя файла
	ext := filepath.Ext(fileHeader.Filename)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadPath, fileName)

	// Создаём директорию, если не существует
	if err := os.MkdirAll(s.uploadPath, 0755); err != nil {
		return nil, err
	}

	// Сохраняем файл
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	// Сохраняем информацию в БД
	file := &models.File{
		MessageID: messageID,
		UserID:    userID,
		Name:      fileHeader.Filename,
		Size:      fileHeader.Size,
		MimeType:  fileHeader.Header.Get("Content-Type"),
		Path:      filePath,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		os.Remove(filePath)
		return nil, err
	}

	return &models.FileResponse{
		ID:        file.ID,
		Name:      file.Name,
		Size:      file.Size,
		MimeType:  file.MimeType,
		URL:       "/api/files/" + file.ID.String(),
		CreatedAt: file.CreatedAt,
	}, nil
}

func (s *fileService) GetFile(ctx context.Context, userID, fileID uuid.UUID) (*models.File, io.ReadCloser, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}

	// Получаем сообщение
	message, err := s.messageRepo.GetByID(ctx, file.MessageID)
	if err != nil {
		return nil, nil, err
	}

	// Проверяем, что пользователь имеет доступ к чату
	isParticipant, err := s.participantRepo.IsParticipant(ctx, message.ChatID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !isParticipant {
		return nil, nil, ErrNotParticipant
	}

	// Открываем файл
	f, err := os.Open(file.Path)
	if err != nil {
		return nil, nil, err
	}

	return file, f, nil
}

func (s *fileService) DeleteFile(ctx context.Context, userID, fileID uuid.UUID) error {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}

	message, err := s.messageRepo.GetByID(ctx, file.MessageID)
	if err != nil {
		return err
	}

	// Только автор сообщения может удалить файл
	if message.UserID != userID {
		return ErrNotMessageAuthor
	}

	// Удаляем физический файл
	if err := os.Remove(file.Path); err != nil {
		log.Print("error deleting the file")
	}

	return s.fileRepo.Delete(ctx, fileID)
}
