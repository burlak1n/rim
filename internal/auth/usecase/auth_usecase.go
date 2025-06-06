package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
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
	GetContactByTelegramID(ctx context.Context, telegramID int64) (*domain.Contact, error)
	IsUserAdmin(ctx context.Context, userID uint) (bool, error)
	UpdateUserContact(ctx context.Context, userID uint, contactData UpdateUserContactData) (*domain.Contact, error)
	Logout(ctx context.Context, sessionToken string) error
}

// UpdateUserContactData определяет данные для обновления контакта пользователя
type UpdateUserContactData struct {
	Name       *string
	Phone      *string
	Email      *string
	Transport  *string
	Printer    *string
	Allergies  *string
	VK         *string
	Telegram   *string
	TelegramID *int64
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

// GetContactByTelegramID получает контакт по telegram_id пользователя
func (uc *authUseCase) GetContactByTelegramID(ctx context.Context, telegramID int64) (*domain.Contact, error) {
	contact, err := uc.contactRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContactNotFound
		}
		uc.logger.ErrorContext(ctx, "Failed to get contact by telegram ID", slog.Int64("telegram_id", telegramID), slog.Any("error", err))
		return nil, err
	}
	return contact, nil
}

// IsUserAdmin проверяет принадлежит ли пользователь к группе "Администраторы"
func (uc *authUseCase) IsUserAdmin(ctx context.Context, userID uint) (bool, error) {
	// Получаем пользователя с контактом
	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to get user for admin check", slog.Uint64("user_id", uint64(userID)), slog.Any("error", err))
		return false, err
	}

	// Ищем контакт по telegram_id
	contact, err := uc.contactRepo.GetByTelegramID(ctx, user.TelegramID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // Нет контакта - не администратор
		}
		uc.logger.ErrorContext(ctx, "Failed to get contact for admin check", slog.Int64("telegram_id", user.TelegramID), slog.Any("error", err))
		return false, err
	}

	// Логируем информацию о группах контакта для отладки
	groupNames := make([]string, len(contact.Groups))
	for i, group := range contact.Groups {
		groupNames[i] = group.Name
	}
	uc.logger.InfoContext(ctx, "Contact groups for admin check",
		slog.Int64("telegram_id", user.TelegramID),
		slog.Uint64("contact_id", uint64(contact.ID)),
		slog.Any("groups", groupNames))

	// Проверяем есть ли группа "Администраторы"
	for _, group := range contact.Groups {
		if group.Name == "Администраторы" {
			uc.logger.InfoContext(ctx, "User is admin", slog.Uint64("user_id", uint64(userID)))
			return true, nil
		}
	}

	uc.logger.InfoContext(ctx, "User is not admin", slog.Uint64("user_id", uint64(userID)))
	return false, nil
}

// UpdateUserContact обновляет контакт пользователя
func (uc *authUseCase) UpdateUserContact(ctx context.Context, userID uint, contactData UpdateUserContactData) (*domain.Contact, error) {
	// Получаем пользователя
	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to get user for contact update", slog.Uint64("user_id", uint64(userID)), slog.Any("error", err))
		return nil, err
	}

	// Ищем контакт по telegram_id
	contact, err := uc.contactRepo.GetByTelegramID(ctx, user.TelegramID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContactNotFound
		}
		uc.logger.ErrorContext(ctx, "Failed to get contact for update", slog.Int64("telegram_id", user.TelegramID), slog.Any("error", err))
		return nil, err
	}

	// Обновляем поля контакта
	changed := false
	if contactData.Name != nil {
		name := strings.TrimSpace(*contactData.Name)
		if name != "" && contact.Name != name {
			contact.Name = name
			changed = true
		}
	}
	if contactData.Email != nil {
		email := strings.TrimSpace(*contactData.Email)
		if email != "" && contact.Email != email {
			// Проверка уникальности email
			existingByEmail, err := uc.contactRepo.GetByEmail(ctx, email)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			if existingByEmail != nil && existingByEmail.ID != contact.ID {
				return nil, errors.New("email already exists")
			}
			contact.Email = email
			changed = true
		}
	}
	if contactData.Phone != nil {
		phone := strings.TrimSpace(*contactData.Phone)
		if phone != "" && contact.Phone != phone {
			// Проверка уникальности phone
			existingByPhone, err := uc.contactRepo.GetByPhone(ctx, phone)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			if existingByPhone != nil && existingByPhone.ID != contact.ID {
				return nil, errors.New("phone already exists")
			}
			contact.Phone = phone
			changed = true
		}
	}
	if contactData.Transport != nil && contact.Transport != *contactData.Transport {
		contact.Transport = *contactData.Transport
		changed = true
	}
	if contactData.Printer != nil && contact.Printer != *contactData.Printer {
		contact.Printer = *contactData.Printer
		changed = true
	}
	if contactData.Allergies != nil && contact.Allergies != *contactData.Allergies {
		contact.Allergies = *contactData.Allergies
		changed = true
	}
	if contactData.VK != nil && contact.VK != *contactData.VK {
		contact.VK = *contactData.VK
		changed = true
	}
	if contactData.Telegram != nil && contact.Telegram != *contactData.Telegram {
		contact.Telegram = *contactData.Telegram
		changed = true
	}
	if contactData.TelegramID != nil && contact.TelegramID != *contactData.TelegramID {
		contact.TelegramID = *contactData.TelegramID
		changed = true
	}

	if !changed {
		return contact, nil
	}

	// Обновляем контакт
	if err := uc.contactRepo.Update(ctx, contact); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update user contact", slog.Uint64("contact_id", uint64(contact.ID)), slog.Any("error", err))
		return nil, err
	}

	return contact, nil
}

// Logout завершает сессию пользователя
func (uc *authUseCase) Logout(ctx context.Context, sessionToken string) error {
	return uc.authRepo.DeleteSession(ctx, sessionToken)
}

// verifyTelegramAuth проверяет подлинность данных авторизации от Telegram
func (uc *authUseCase) verifyTelegramAuth(authData TelegramAuthData, botToken string) bool {
	// Добавляем логирование для диагностики
	uc.logger.Debug("Verifying telegram auth",
		slog.Int64("telegram_id", authData.ID),
		slog.String("first_name", authData.FirstName),
		slog.String("username", authData.Username),
		slog.Int64("auth_date", authData.AuthDate),
		slog.String("received_hash", authData.Hash))

	// Создаем строку для проверки подписи
	dataCheckString := uc.createDataCheckString(authData)
	uc.logger.Debug("Data check string created", slog.String("data_check_string", dataCheckString))

	// Создаем секретный ключ
	secretKey := sha256.Sum256([]byte(botToken))
	uc.logger.Debug("Bot token hash created", slog.String("bot_token_length", fmt.Sprintf("%d", len(botToken))))

	// Вычисляем HMAC
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))
	uc.logger.Debug("Expected hash calculated", slog.String("expected_hash", expectedHash))

	// Проверяем совпадение хэшей
	if expectedHash != authData.Hash {
		uc.logger.Warn("Hash mismatch",
			slog.String("expected", expectedHash),
			slog.String("received", authData.Hash))
		return false
	}

	// Проверяем актуальность данных (не старше 1 дня)
	authTime := time.Unix(authData.AuthDate, 0)
	timeSince := time.Since(authTime)
	uc.logger.Debug("Checking auth time",
		slog.Time("auth_time", authTime),
		slog.Duration("time_since", timeSince))

	if timeSince > 24*time.Hour {
		uc.logger.Warn("Auth data too old", slog.Duration("age", timeSince))
		return false
	}

	uc.logger.Debug("Telegram auth verification successful")
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

	// Создаем строку проверки БЕЗ URL-кодирования согласно документации Telegram
	var parts []string
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, params[key]))
	}

	return strings.Join(parts, "\n")
}
