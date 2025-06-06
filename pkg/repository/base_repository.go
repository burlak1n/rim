package repository

import (
	"context"
	"log/slog"

	"gorm.io/gorm"
)

// BaseRepository предоставляет базовые CRUD операции для любой модели
type BaseRepository[T any] struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewBaseRepository создает новый экземпляр BaseRepository
func NewBaseRepository[T any](db *gorm.DB, logger *slog.Logger) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:     db,
		logger: logger,
	}
}

// Create создает новую запись в базе данных
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to create entity", slog.Any("error", err))
		return nil, err
	}
	return entity, nil
}

// GetByID получает запись по ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Entity not found", slog.Uint64("id", uint64(id)))
		} else {
			r.logger.ErrorContext(ctx, "Failed to get entity by ID", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		}
		return nil, err
	}
	return &entity, nil
}

// GetAll получает все записи
func (r *BaseRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	var entities []T
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to get all entities", slog.Any("error", err))
		return nil, err
	}
	return entities, nil
}

// Update обновляет запись в базе данных
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) (*T, error) {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to update entity", slog.Any("error", err))
		return nil, err
	}
	return entity, nil
}

// Delete удаляет запись (мягкое удаление)
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	if err := r.db.WithContext(ctx).Delete(&entity, id).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete entity", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return err
	}
	return nil
}

// HardDelete удаляет запись из базы данных навсегда
func (r *BaseRepository[T]) HardDelete(ctx context.Context, id uint) error {
	var entity T
	if err := r.db.WithContext(ctx).Unscoped().Delete(&entity, id).Error; err != nil {
		r.logger.ErrorContext(ctx, "Failed to hard delete entity", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return err
	}
	return nil
}

// DB возвращает указатель на базу данных для кастомных запросов
func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

// Logger возвращает логгер для использования в наследуемых репозиториях
func (r *BaseRepository[T]) Logger() *slog.Logger {
	return r.logger
}
