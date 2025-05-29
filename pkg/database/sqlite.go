package database

import (
	"log/slog"

	"rim/internal/config"
	"rim/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewSQLiteConnection устанавливает соединение с базой данных SQLite.
// Также выполняет автоматическую миграцию для моделей Contact и Group.
func NewSQLiteConnection(cfg *config.Config, logger *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to SQLite", slog.String("path", cfg.SQLitePath), slog.Any("error", err))
		return nil, err
	}

	logger.Info("Successfully connected to SQLite", slog.String("path", cfg.SQLitePath))

	// Выполняем автомиграцию для моделей Contact и Group
	err = db.AutoMigrate(&domain.Contact{}, &domain.Group{})
	if err != nil {
		logger.Error("Failed to migrate database schema", slog.Any("error", err))
		return nil, err
	}
	logger.Info("Database schema migrated successfully for Contact and Group models")

	return db, nil
}
