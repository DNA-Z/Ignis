package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
)

type QueryOption func(*gorm.DB) *gorm.DB

func WithPagination(limit int, offset int) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		if limit > 0 {
			db = db.Limit(limit)
		}
		if offset > 0 {
			db = db.Offset(offset)
		}
		return db
	}
}

func WithOrderBy(order string) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(order)
	}
}

func WithCondition(query string, args ...interface{}) QueryOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	}
}

type GenericRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	FindOne(ctx context.Context, condition map[string]interface{}) (*T, error)
	FindAll(ctx context.Context, opts ...QueryOption) ([]*T, int64, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	Count(ctx context.Context, condition map[string]interface{}) (int64, error)
}

type genericRepository[T any] struct {
	db *gorm.DB
}

func NewGenericRepository[T any](db *gorm.DB) GenericRepository[T] {
	return &genericRepository[T]{db: db}
}

func (r *genericRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *genericRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *genericRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&entity).Error
}

func (r *genericRepository[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&entity).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &entity, err
}

func (r *genericRepository[T]) FindOne(ctx context.Context, condition map[string]interface{}) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).
		Where(condition).
		First(&entity).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &entity, err
}

func (r *genericRepository[T]) FindAll(ctx context.Context, opts ...QueryOption) ([]*T, int64, error) {
	var entities []*T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T))
	for _, opt := range opts {
		query = opt(query)
	}

	countQuery := r.db.WithContext(ctx).Model(new(T))
	for _, opt := range opts {
		if opt != nil {
			countQuery = opt(countQuery)
		}
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (r *genericRepository[T]) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := r.Count(ctx, map[string]interface{}{"id": id})
	return count > 0, err
}

func (r *genericRepository[T]) Count(ctx context.Context, condition map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))

	if len(condition) > 0 {
		query = query.Where(condition)
	}

	err := query.Count(&count).Error
	return count, err
}
