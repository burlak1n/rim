package database

import (
	"context"
	"log/slog"

	"rim/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient создает и настраивает нового клиента Redis.
func NewRedisClient(cfg *config.Config, logger *slog.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Проверяем соединение
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis",
			slog.String("address", cfg.RedisAddr),
			slog.Int("db", cfg.RedisDB),
			slog.Any("error", err))
		return nil, err
	}

	logger.Info("Successfully connected to Redis", slog.String("address", cfg.RedisAddr), slog.Int("db", cfg.RedisDB))
	return rdb, nil
}
