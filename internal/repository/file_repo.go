package repository

import (
	"context"
	"github.com/DNA-Z/Ignis/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileRepository interface {
	Create(ctx context.Context, file *models.File) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.File, error)
	GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]*models.File, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&file).Error
	return &file, err
}

func (r *fileRepository) GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]*models.File, error) {
	var files []*models.File
	err := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Find(&files).Error
	return files, err
}

func (r *fileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.File{}, "id = ?", id).Error
}
