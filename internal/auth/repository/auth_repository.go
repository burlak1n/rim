package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"rim/internal/domain"
	"rim/pkg/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Repository определяет интерфейс для auth репозитория
type Repository interface {
	// Базовые операции с пользователями
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByID(ctx context.Context, id uint) (*domain.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)

	// Операции с сессиями в Redis
	CreateSession(ctx context.Context, session *domain.UserSession) error
	GetSession(ctx context.Context, sessionToken string) (*domain.UserSession, error)
	DeleteSession(ctx context.Context, sessionToken string) error
	DeleteAllUserSessions(ctx context.Context, userID uint) error
}

type authRepository struct {
	*repository.BaseRepository[domain.User]
	redisClient *redis.Client
}

// NewAuthRepository создает новый экземпляр auth репозитория
func NewAuthRepository(db *gorm.DB, redisClient *redis.Client, logger *slog.Logger) Repository {
	return &authRepository{
		BaseRepository: repository.NewBaseRepository[domain.User](db, logger),
		redisClient:    redisClient,
	}
}

// CreateUser создает нового пользователя
func (r *authRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return r.BaseRepository.Create(ctx, user)
}

// GetUserByID получает пользователя по ID
func (r *authRepository) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	return r.BaseRepository.GetByID(ctx, id)
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (r *authRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	var user domain.User
	err := r.DB().WithContext(ctx).Preload("Contact").Where("telegram_id = ? AND is_active = ?", telegramID, true).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger().WarnContext(ctx, "User not found by telegram ID", slog.Int64("telegram_id", telegramID))
		} else {
			r.Logger().ErrorContext(ctx, "Failed to get user by telegram ID", slog.Int64("telegram_id", telegramID), slog.Any("error", err))
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser обновляет данные пользователя
func (r *authRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return r.BaseRepository.Update(ctx, user)
}

// CreateSession создает сессию в Redis
func (r *authRepository) CreateSession(ctx context.Context, session *domain.UserSession) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		r.Logger().ErrorContext(ctx, "Failed to marshal session", slog.Any("error", err))
		return err
	}

	key := r.getSessionKey(session.SessionToken)
	ttl := time.Until(session.ExpiredAt)

	if err := r.redisClient.Set(ctx, key, sessionData, ttl).Err(); err != nil {
		r.Logger().ErrorContext(ctx, "Failed to create session in Redis", slog.String("session_token", session.SessionToken), slog.Any("error", err))
		return err
	}

	r.Logger().InfoContext(ctx, "Session created successfully", slog.String("session_token", session.SessionToken), slog.Uint64("user_id", uint64(session.UserID)))
	return nil
}

// GetSession получает сессию из Redis
func (r *authRepository) GetSession(ctx context.Context, sessionToken string) (*domain.UserSession, error) {
	key := r.getSessionKey(sessionToken)

	sessionData, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.Logger().WarnContext(ctx, "Session not found", slog.String("session_token", sessionToken))
			return nil, fmt.Errorf("session not found")
		}
		r.Logger().ErrorContext(ctx, "Failed to get session from Redis", slog.String("session_token", sessionToken), slog.Any("error", err))
		return nil, err
	}

	var session domain.UserSession
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		r.Logger().ErrorContext(ctx, "Failed to unmarshal session", slog.String("session_token", sessionToken), slog.Any("error", err))
		return nil, err
	}

	// Проверяем, не истекла ли сессия
	if time.Now().After(session.ExpiredAt) {
		r.Logger().WarnContext(ctx, "Session expired", slog.String("session_token", sessionToken))
		// Удаляем истекшую сессию
		r.DeleteSession(ctx, sessionToken)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// DeleteSession удаляет сессию из Redis
func (r *authRepository) DeleteSession(ctx context.Context, sessionToken string) error {
	key := r.getSessionKey(sessionToken)

	if err := r.redisClient.Del(ctx, key).Err(); err != nil {
		r.Logger().ErrorContext(ctx, "Failed to delete session from Redis", slog.String("session_token", sessionToken), slog.Any("error", err))
		return err
	}

	r.Logger().InfoContext(ctx, "Session deleted successfully", slog.String("session_token", sessionToken))
	return nil
}

// DeleteAllUserSessions удаляет все сессии пользователя
func (r *authRepository) DeleteAllUserSessions(ctx context.Context, userID uint) error {
	pattern := fmt.Sprintf("session:user:%d:*", userID)

	keys, err := r.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		r.Logger().ErrorContext(ctx, "Failed to get user sessions keys", slog.Uint64("user_id", uint64(userID)), slog.Any("error", err))
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	if err := r.redisClient.Del(ctx, keys...).Err(); err != nil {
		r.Logger().ErrorContext(ctx, "Failed to delete user sessions", slog.Uint64("user_id", uint64(userID)), slog.Any("error", err))
		return err
	}

	r.Logger().InfoContext(ctx, "All user sessions deleted", slog.Uint64("user_id", uint64(userID)), slog.Int("count", len(keys)))
	return nil
}

// getSessionKey формирует ключ для хранения сессии в Redis
func (r *authRepository) getSessionKey(sessionToken string) string {
	return fmt.Sprintf("session:%s", sessionToken)
}
