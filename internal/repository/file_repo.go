package repository

import (
	"context"
	"github.com/DNA-Z/Ignis/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileRepository interface {
	GenericRepository[models.File]
	GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]*models.File, error)
}

type fileRepository struct {
	GenericRepository[models.File]
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{
		GenericRepository: NewGenericRepository[models.File](db),
		db:                db,
	}
}

func (r *fileRepository) GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]*models.File, error) {
	files, _, err := r.FindAll(ctx,
		WithCondition("message_id = ?", messageID),
		WithOrderBy("created_at ASC"),
	)
	return files, err
}
