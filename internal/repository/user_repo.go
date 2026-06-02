package repository

import (
	"context"

	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	GenericRepository[models.User]
	FindByLogin(ctx context.Context, login string) (*models.User, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error)
}

type userRepository struct {
	GenericRepository[models.User]
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		GenericRepository: NewGenericRepository[models.User](db),
		db:                db,
	}
}

func (r *userRepository) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	return r.FindOne(ctx, map[string]interface{}{"login": login})
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&users).Error
	return users, err
}
