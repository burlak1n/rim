package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"rim/internal/auth/repository"
	contactRepo "rim/internal/contact/repository"
	"rim/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidTelegramAuth = errors.New("invalid telegram authentication data")
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrUserNotFound        = errors.New("user not found")
	ErrContactNotFound     = errors.New("contact not found")
)

// TelegramAuthData представляет данные авторизации от Telegram
type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// UseCase определяет интерфейс для auth бизнес-логики
type UseCase interface {
	AuthenticateWithTelegram(ctx context.Context, authData TelegramAuthData, botToken string) (*domain.UserSession, error)
	GetUserBySession(ctx context.Context, sessionToken string) (*domain.User, error)
	Logout(ctx context.Context, sessionToken string) error
}

type authUseCase struct {
	authRepo    repository.Repository
	contactRepo contactRepo.Repository
	logger      *slog.Logger
}

// NewAuthUseCase создает новый экземпляр auth usecase
func NewAuthUseCase(authRepo repository.Repository, contactRepo contactRepo.Repository, logger *slog.Logger) UseCase {
	return &authUseCase{
		authRepo:    authRepo,
		contactRepo: contactRepo,
		logger:      logger,
	}
}

// AuthenticateWithTelegram аутентифицирует пользователя через Telegram
func (uc *authUseCase) AuthenticateWithTelegram(ctx context.Context, authData TelegramAuthData, botToken string) (*domain.UserSession, error) {
	// Проверяем подлинность данных от Telegram
	if !uc.verifyTelegramAuth(authData, botToken) {
		uc.logger.WarnContext(ctx, "Invalid telegram authentication", slog.Int64("telegram_id", authData.ID))
		return nil, ErrInvalidTelegramAuth
	}

	// Ищем существующего пользователя
	user, err := uc.authRepo.GetUserByTelegramID(ctx, authData.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.logger.ErrorContext(ctx, "Failed to get user by telegram ID", slog.Int64("telegram_id", authData.ID), slog.Any("error", err))
		return nil, err
	}

	// Если пользователь не найден, ищем контакт по Telegram ID
	if user == nil {
		contact, err := uc.contactRepo.GetByTelegramID(ctx, authData.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to get contact by telegram ID", slog.Int64("telegram_id", authData.ID), slog.Any("error", err))
			return nil, err
		}

		// Создаем нового пользователя
		user = &domain.User{
			TelegramID: authData.ID,
			IsActive:   true,
		}

		// Если найден контакт, связываем с ним
		if contact != nil {
			user.ContactID = &contact.ID
		}

		user, err = uc.authRepo.CreateUser(ctx, user)
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to create user", slog.Int64("telegram_id", authData.ID), slog.Any("error", err))
			return nil, err
		}

		uc.logger.InfoContext(ctx, "New user created", slog.Uint64("user_id", uint64(user.ID)), slog.Int64("telegram_id", authData.ID))
	}

	// Проверяем, активен ли пользователь
	if !user.IsActive {
		uc.logger.WarnContext(ctx, "User is not active", slog.Uint64("user_id", uint64(user.ID)))
		return nil, ErrUserNotFound
	}

	// Создаем новую сессию
	sessionToken := uuid.New().String()
	session := &domain.UserSession{
		SessionToken: sessionToken,
		UserID:       user.ID,
		CreatedAt:    time.Now(),
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour), // 7 дней
	}

	if err := uc.authRepo.CreateSession(ctx, session); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create session", slog.Uint64("user_id", uint64(user.ID)), slog.Any("error", err))
		return nil, err
	}

	uc.logger.InfoContext(ctx, "User authenticated successfully", slog.Uint64("user_id", uint64(user.ID)), slog.Int64("telegram_id", authData.ID))
	return session, nil
}

// GetUserBySession получает пользователя по сессии
func (uc *authUseCase) GetUserBySession(ctx context.Context, sessionToken string) (*domain.User, error) {
	session, err := uc.authRepo.GetSession(ctx, sessionToken)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrSessionNotFound
		}
		if strings.Contains(err.Error(), "expired") {
			return nil, ErrSessionExpired
		}
		return nil, err
	}

	user, err := uc.authRepo.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		uc.logger.ErrorContext(ctx, "Failed to get user by ID", slog.Uint64("user_id", uint64(session.UserID)), slog.Any("error", err))
		return nil, err
	}

	if !user.IsActive {
		uc.logger.WarnContext(ctx, "User is not active", slog.Uint64("user_id", uint64(user.ID)))
		return nil, ErrUserNotFound
	}

	return user, nil
}

// Logout завершает сессию пользователя
func (uc *authUseCase) Logout(ctx context.Context, sessionToken string) error {
	return uc.authRepo.DeleteSession(ctx, sessionToken)
}

// verifyTelegramAuth проверяет подлинность данных авторизации от Telegram
func (uc *authUseCase) verifyTelegramAuth(authData TelegramAuthData, botToken string) bool {
	// Создаем строку для проверки подписи
	dataCheckString := uc.createDataCheckString(authData)

	// Создаем секретный ключ
	secretKey := sha256.Sum256([]byte(botToken))

	// Вычисляем HMAC
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// Проверяем совпадение хэшей
	if expectedHash != authData.Hash {
		return false
	}

	// Проверяем актуальность данных (не старше 1 дня)
	authTime := time.Unix(authData.AuthDate, 0)
	if time.Since(authTime) > 24*time.Hour {
		return false
	}

	return true
}

// createDataCheckString создает строку для проверки подписи
func (uc *authUseCase) createDataCheckString(authData TelegramAuthData) string {
	params := make(map[string]string)

	params["id"] = strconv.FormatInt(authData.ID, 10)
	params["auth_date"] = strconv.FormatInt(authData.AuthDate, 10)

	if authData.FirstName != "" {
		params["first_name"] = authData.FirstName
	}
	if authData.LastName != "" {
		params["last_name"] = authData.LastName
	}
	if authData.Username != "" {
		params["username"] = authData.Username
	}
	if authData.PhotoURL != "" {
		params["photo_url"] = authData.PhotoURL
	}

	// Сортируем ключи
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Создаем строку проверки
	var parts []string
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, url.QueryEscape(params[key])))
	}

	return strings.Join(parts, "\n")
}
