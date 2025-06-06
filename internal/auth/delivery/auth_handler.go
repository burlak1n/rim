package delivery

import (
	"log/slog"
	"net/http"

	"rim/internal/auth/usecase"

	"github.com/gofiber/fiber/v2"
)

// Handler представляет HTTP обработчик для auth
type Handler struct {
	authUseCase usecase.UseCase
	logger      *slog.Logger
	botToken    string
}

// NewHandler создает новый экземпляр auth handler
func NewHandler(authUseCase usecase.UseCase, botToken string, logger *slog.Logger) *Handler {
	return &Handler{
		authUseCase: authUseCase,
		logger:      logger,
		botToken:    botToken,
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
	Contact    *ContactResponse `json:"contact,omitempty"`
	CreatedAt  string           `json:"created_at"`
}

// ContactResponse представляет информацию о контакте
type ContactResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Transport string `json:"transport"`
	Printer   string `json:"printer"`
	Allergies string `json:"allergies"`
	VK        string `json:"vk"`
	Telegram  string `json:"telegram"`
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

	response := UserResponse{
		ID:         user.ID,
		TelegramID: user.TelegramID,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Если у пользователя есть связанный контакт, добавляем его информацию
	if user.Contact != nil {
		response.Contact = &ContactResponse{
			ID:        user.Contact.ID,
			Name:      user.Contact.Name,
			Phone:     user.Contact.Phone,
			Email:     user.Contact.Email,
			Transport: user.Contact.Transport,
			Printer:   user.Contact.Printer,
			Allergies: user.Contact.Allergies,
			VK:        user.Contact.VK,
			Telegram:  user.Contact.Telegram,
		}
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
