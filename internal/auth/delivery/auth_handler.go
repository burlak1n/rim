package delivery

import (
	"log/slog"
	"net/http"
	"time"

	"rim/internal/auth/usecase"
	systemUseCase "rim/internal/system/usecase"

	"github.com/gofiber/fiber/v2"
)

// Handler представляет HTTP обработчик для auth
type Handler struct {
	authUseCase    usecase.UseCase
	systemUseCase  systemUseCase.UseCase
	logger         *slog.Logger
	botToken       string
	forceDebugMode bool
}

// NewHandler создает новый экземпляр auth handler
func NewHandler(authUseCase usecase.UseCase, systemUseCase systemUseCase.UseCase, botToken string, forceDebugMode bool, logger *slog.Logger) *Handler {
	return &Handler{
		authUseCase:    authUseCase,
		systemUseCase:  systemUseCase,
		logger:         logger,
		botToken:       botToken,
		forceDebugMode: forceDebugMode,
	}
}

// TelegramAuthRequest представляет запрос авторизации через Telegram
type TelegramAuthRequest struct {
	ID        int64  `json:"id" validate:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date" validate:"required"`
	Hash      string `json:"hash" validate:"required"`
}

// SessionResponse представляет ответ с токеном сессии
type SessionResponse struct {
	SessionToken string `json:"session_token"`
	ExpiresAt    string `json:"expires_at"`
}

// UserResponse представляет информацию о пользователе
type UserResponse struct {
	ID         uint             `json:"id"`
	TelegramID int64            `json:"telegram_id"`
	IsActive   bool             `json:"is_active"`
	IsAdmin    bool             `json:"is_admin"` // Флаг администратора
	Contact    *ContactResponse `json:"contact,omitempty"`
	CreatedAt  string           `json:"created_at"`
}

// ContactResponse представляет информацию о контакте
type ContactResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Transport  string `json:"transport"`
	Printer    string `json:"printer"`
	Allergies  string `json:"allergies"`
	VK         string `json:"vk"`
	Telegram   string `json:"telegram"`
	TelegramID int64  `json:"telegram_id,omitempty"`
}

// UpdateContactRequest представляет запрос на обновление контакта пользователя
type UpdateContactRequest struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone      *string `json:"phone,omitempty" validate:"omitempty,e164"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email"`
	Transport  *string `json:"transport,omitempty" validate:"omitempty,oneof='есть машина' 'есть права' 'нет ничего'"`
	Printer    *string `json:"printer,omitempty" validate:"omitempty,oneof='цветной' 'обычный' 'нет'"`
	Allergies  *string `json:"allergies,omitempty" validate:"omitempty,max=255"`
	VK         *string `json:"vk,omitempty" validate:"omitempty,url"`
	Telegram   *string `json:"telegram,omitempty" validate:"omitempty,alphanum"`
	TelegramID *int64  `json:"telegram_id,omitempty"`
}

// AuthWithTelegram обрабатывает авторизацию через Telegram
// @Summary Авторизация через Telegram
// @Description Аутентифицирует пользователя через Telegram Auth Widget
// @Tags auth
// @Accept json
// @Produce json
// @Param telegram_data body TelegramAuthRequest true "Данные авторизации от Telegram"
// @Success 200 {object} SessionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/telegram [post]
func (h *Handler) AuthWithTelegram(c *fiber.Ctx) error {
	var req TelegramAuthRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnContext(c.Context(), "Invalid request body", slog.Any("error", err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Преобразуем в структуру usecase
	authData := usecase.TelegramAuthData{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		PhotoURL:  req.PhotoURL,
		AuthDate:  req.AuthDate,
		Hash:      req.Hash,
	}

	session, err := h.authUseCase.AuthenticateWithTelegram(c.Context(), authData, h.botToken)
	if err != nil {
		switch err {
		case usecase.ErrInvalidTelegramAuth:
			h.logger.WarnContext(c.Context(), "Invalid telegram authentication", slog.Int64("telegram_id", req.ID))
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid telegram authentication",
			})
		default:
			h.logger.ErrorContext(c.Context(), "Failed to authenticate with telegram", slog.Any("error", err))
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	// Устанавливаем httpOnly cookie для защиты от XSS
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    session.SessionToken,
		Expires:  session.ExpiredAt,
		HTTPOnly: true,
		Secure:   true,     // Только по HTTPS в продакшене
		SameSite: "Strict", // Защита от CSRF
		Path:     "/",
	})

	response := SessionResponse{
		SessionToken: session.SessionToken,
		ExpiresAt:    session.ExpiredAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	h.logger.InfoContext(c.Context(), "User authenticated successfully", slog.Uint64("user_id", uint64(session.UserID)))
	return c.JSON(response)
}

// GetMe возвращает информацию о текущем пользователе
// @Summary Получить информацию о пользователе
// @Description Возвращает информацию о пользователе по токену сессии
// @Tags auth
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/me [get]
func (h *Handler) GetMe(c *fiber.Ctx) error {
	sessionToken := h.extractSessionToken(c)
	if sessionToken == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	user, err := h.authUseCase.GetUserBySession(c.Context(), sessionToken)
	if err != nil {
		switch err {
		case usecase.ErrSessionNotFound, usecase.ErrSessionExpired:
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		case usecase.ErrUserNotFound:
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		default:
			h.logger.ErrorContext(c.Context(), "Failed to get user by session", slog.Any("error", err))
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	// Проверяем права администратора (учитываем отладочный режим)
	isAdmin, err := h.authUseCase.IsUserAdmin(c.Context(), user.ID)
	if err != nil {
		h.logger.WarnContext(c.Context(), "Failed to check admin status", slog.Uint64("user_id", uint64(user.ID)), slog.Any("error", err))
		isAdmin = false // По умолчанию не администратор
	}

	// Если не администратор, проверяем отладочный режим
	if !isAdmin {
		// Сначала проверяем принудительный отладочный режим из переменной окружения
		if h.forceDebugMode {
			h.logger.InfoContext(c.Context(), "Force debug mode is enabled, user gets admin rights", slog.Uint64("user_id", uint64(user.ID)))
			isAdmin = true
		} else {
			// Затем проверяем отладочный режим из базы данных
			debugMode, err := h.systemUseCase.GetDebugMode(c.Context())
			if err != nil {
				h.logger.WarnContext(c.Context(), "Failed to get debug mode status", slog.Any("error", err))
			} else if debugMode {
				h.logger.InfoContext(c.Context(), "Debug mode is enabled, user gets admin rights", slog.Uint64("user_id", uint64(user.ID)))
				isAdmin = true
			}
		}
	}

	response := UserResponse{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		IsActive:   user.IsActive,
		IsAdmin:    isAdmin,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Ищем контакт по telegram_id пользователя
	contact, err := h.authUseCase.GetContactByTelegramID(c.Context(), user.TelegramID)
	if err != nil && err != usecase.ErrContactNotFound {
		h.logger.WarnContext(c.Context(), "Failed to get contact by telegram_id", slog.Int64("telegram_id", user.TelegramID), slog.Any("error", err))
	}

	// Если контакт найден, добавляем его информацию
	if contact != nil {
		response.Contact = &ContactResponse{
			ID:         contact.ID,
			Name:       contact.Name,
			Phone:      contact.Phone,
			Email:      contact.Email,
			Transport:  contact.Transport,
			Printer:    contact.Printer,
			Allergies:  contact.Allergies,
			VK:         contact.VK,
			Telegram:   contact.Telegram,
			TelegramID: contact.TelegramID,
		}
	}

	return c.JSON(response)
}

// UpdateMyContact обновляет контакт текущего пользователя
// @Summary Обновить свой контакт
// @Description Обновляет контакт пользователя, найденный по telegram_id
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param contact body UpdateContactRequest true "Данные для обновления контакта"
// @Success 200 {object} ContactResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/contact [put]
func (h *Handler) UpdateMyContact(c *fiber.Ctx) error {
	sessionToken := h.extractSessionToken(c)
	if sessionToken == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	user, err := h.authUseCase.GetUserBySession(c.Context(), sessionToken)
	if err != nil {
		switch err {
		case usecase.ErrSessionNotFound, usecase.ErrSessionExpired:
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		case usecase.ErrUserNotFound:
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		default:
			h.logger.ErrorContext(c.Context(), "Failed to get user by session", slog.Any("error", err))
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	var req UpdateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	contactData := usecase.UpdateUserContactData{
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      req.Email,
		Transport:  req.Transport,
		Printer:    req.Printer,
		Allergies:  req.Allergies,
		VK:         req.VK,
		Telegram:   req.Telegram,
		TelegramID: req.TelegramID,
	}

	updatedContact, err := h.authUseCase.UpdateUserContact(c.Context(), user.ID, contactData)
	if err != nil {
		if err == usecase.ErrContactNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Contact not found",
			})
		}
		h.logger.ErrorContext(c.Context(), "Failed to update user contact", slog.Uint64("user_id", uint64(user.ID)), slog.Any("error", err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	response := ContactResponse{
		ID:         updatedContact.ID,
		Name:       updatedContact.Name,
		Phone:      updatedContact.Phone,
		Email:      updatedContact.Email,
		Transport:  updatedContact.Transport,
		Printer:    updatedContact.Printer,
		Allergies:  updatedContact.Allergies,
		VK:         updatedContact.VK,
		Telegram:   updatedContact.Telegram,
		TelegramID: updatedContact.TelegramID,
	}

	return c.JSON(response)
}

// Logout завершает сессию пользователя
// @Summary Выход из системы
// @Description Завершает текущую сессию пользователя
// @Tags auth
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(c *fiber.Ctx) error {
	sessionToken := h.extractSessionToken(c)
	if sessionToken == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	err := h.authUseCase.Logout(c.Context(), sessionToken)
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to logout", slog.Any("error", err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Удаляем cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

// extractSessionToken извлекает токен сессии из заголовка Authorization
func (h *Handler) extractSessionToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Ожидаем формат "Bearer <token>"
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return ""
	}

	return authHeader[len(bearerPrefix):]
}
