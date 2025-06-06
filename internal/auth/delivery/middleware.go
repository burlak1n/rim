package delivery

import (
	"net/http"

	"rim/internal/auth/usecase"
	"rim/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware проверяет авторизацию пользователя
func (h *Handler) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionToken := h.extractSessionToken(c)
		if sessionToken == "" {
			// Если токена нет, сохраняем информацию о том, что пользователь не авторизован
			c.Locals("user", nil)
			c.Locals("isAuthenticated", false)
			return c.Next()
		}

		user, err := h.authUseCase.GetUserBySession(c.Context(), sessionToken)
		if err != nil {
			// Если сессия недействительна, считаем пользователя неавторизованным
			c.Locals("user", nil)
			c.Locals("isAuthenticated", false)
			return c.Next()
		}

		// Сохраняем информацию о пользователе в контексте
		c.Locals("user", user)
		c.Locals("isAuthenticated", true)
		return c.Next()
	}
}

// RequireAuth middleware, который требует обязательной авторизации
func (h *Handler) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
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
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}

		// Сохраняем информацию о пользователе в контексте
		c.Locals("user", user)
		c.Locals("isAuthenticated", true)
		return c.Next()
	}
}

// GetUserFromContext получает пользователя из контекста Fiber
func GetUserFromContext(c *fiber.Ctx) (*domain.User, bool) {
	user := c.Locals("user")
	isAuth := c.Locals("isAuthenticated")

	if isAuth == nil || user == nil {
		return nil, false
	}

	if isAuthBool, ok := isAuth.(bool); ok && isAuthBool {
		if userTyped, ok := user.(*domain.User); ok {
			return userTyped, true
		}
	}
	return nil, false
}

// IsAuthenticated проверяет, авторизован ли пользователь
func IsAuthenticated(c *fiber.Ctx) bool {
	isAuth := c.Locals("isAuthenticated")
	if isAuth == nil {
		return false
	}
	if isAuthBool, ok := isAuth.(bool); ok {
		return isAuthBool
	}
	return false
}
