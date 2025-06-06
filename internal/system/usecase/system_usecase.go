package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	systemRepo "rim/internal/system/repository"

	"gorm.io/gorm"
)

const (
	DebugModeKey = "debug_mode"
)

var (
	ErrSettingNotFound = errors.New("setting not found")
)

// UseCase определяет интерфейс для системной бизнес-логики
type UseCase interface {
	GetDebugMode(ctx context.Context) (bool, error)
	SetDebugMode(ctx context.Context, enabled bool) error
}

type systemUseCase struct {
	systemRepo systemRepo.Repository
	logger     *slog.Logger
}

// NewSystemUseCase создает новый экземпляр системного UseCase
func NewSystemUseCase(systemRepo systemRepo.Repository, logger *slog.Logger) UseCase {
	return &systemUseCase{
		systemRepo: systemRepo,
		logger:     logger,
	}
}

func (uc *systemUseCase) GetDebugMode(ctx context.Context) (bool, error) {
	setting, err := uc.systemRepo.GetSetting(ctx, DebugModeKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если настройка не найдена, возвращаем false по умолчанию
			return false, nil
		}
		uc.logger.ErrorContext(ctx, "Failed to get debug mode setting", slog.Any("error", err))
		return false, err
	}

	debugMode, err := strconv.ParseBool(setting.Value)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to parse debug mode value", slog.String("value", setting.Value), slog.Any("error", err))
		return false, err
	}

	return debugMode, nil
}

func (uc *systemUseCase) SetDebugMode(ctx context.Context, enabled bool) error {
	value := strconv.FormatBool(enabled)
	if err := uc.systemRepo.SetSetting(ctx, DebugModeKey, value); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to set debug mode setting", slog.Bool("enabled", enabled), slog.Any("error", err))
		return err
	}

	uc.logger.InfoContext(ctx, "Debug mode setting updated", slog.Bool("enabled", enabled))
	return nil
}
