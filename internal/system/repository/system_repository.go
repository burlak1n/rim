package repository

import (
	"context"
	"log/slog"

	"rim/internal/domain"

	"gorm.io/gorm"
)

// Repository определяет интерфейс для операций с системными настройками
type Repository interface {
	GetSetting(ctx context.Context, key string) (*domain.SystemSetting, error)
	SetSetting(ctx context.Context, key, value string) error
}

type sqliteRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewSQLiteRepository создает новый экземпляр sqliteRepository для системных настроек
func NewSQLiteRepository(db *gorm.DB, logger *slog.Logger) Repository {
	return &sqliteRepository{
		db:     db,
		logger: logger,
	}
}

func (r *sqliteRepository) GetSetting(ctx context.Context, key string) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.InfoContext(ctx, "System setting not found", slog.String("key", key))
			return nil, err
		}
		r.logger.ErrorContext(ctx, "Error getting system setting", slog.String("key", key), slog.Any("error", err))
		return nil, err
	}
	return &setting, nil
}

func (r *sqliteRepository) SetSetting(ctx context.Context, key, value string) error {
	setting := &domain.SystemSetting{
		Key:   key,
		Value: value,
	}

	// Используем OnConflict для обновления существующего значения
	if err := r.db.WithContext(ctx).
		Where("key = ?", key).
		Assign(domain.SystemSetting{Value: value}).
		FirstOrCreate(setting).Error; err != nil {
		r.logger.ErrorContext(ctx, "Error setting system setting", slog.String("key", key), slog.String("value", value), slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, "Successfully set system setting", slog.String("key", key), slog.String("value", value))
	return nil
}
