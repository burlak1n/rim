package delivery

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SecurityMiddleware добавляет заголовки безопасности для защиты от XSS
func SecurityMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Защита от XSS
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy для защиты от XSS
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://telegram.org; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self' http://localhost:3000; " +
			"font-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'"
		c.Set("Content-Security-Policy", csp)

		return c.Next()
	}
}

// CSRFMiddleware генерирует и проверяет CSRF токены
func (h *Handler) CSRFMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Пропускаем CSRF проверку для GET запросов и некоторых эндпоинтов
		if c.Method() == "GET" ||
			strings.HasSuffix(c.Path(), "/telegram") ||
			strings.HasSuffix(c.Path(), "/debug-mode") {
			return c.Next()
		}

		// Проверяем CSRF токен
		csrfToken := c.Get("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.FormValue("_csrf")
		}

		sessionToken := h.extractSessionToken(c)
		if sessionToken == "" {
			// Если нет сессии, пропускаем CSRF проверку
			return c.Next()
		}

		// Проверяем валидность CSRF токена
		if !h.validateCSRFToken(c, sessionToken, csrfToken) {
			h.logger.WarnContext(c.Context(), "Invalid CSRF token",
				"ip", c.IP(),
				"user_agent", c.Get("User-Agent"))
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid CSRF token",
			})
		}

		return c.Next()
	}
}

// GetCSRFToken возвращает CSRF токен для текущей сессии
func (h *Handler) GetCSRFToken(c *fiber.Ctx) error {
	sessionToken := h.extractSessionToken(c)
	if sessionToken == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Session required",
		})
	}

	csrfToken := h.generateCSRFToken(sessionToken)

	return c.JSON(fiber.Map{
		"csrf_token": csrfToken,
	})
}

// generateCSRFToken генерирует CSRF токен на основе сессии
func (h *Handler) generateCSRFToken(sessionToken string) string {
	// Генерируем случайные байты
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)

	// Комбинируем с частью токена сессии для связки
	sessionPart := sessionToken[:8] // Первые 8 символов сессии
	csrfToken := hex.EncodeToString(randomBytes) + sessionPart

	return csrfToken
}

// validateCSRFToken проверяет валидность CSRF токена
func (h *Handler) validateCSRFToken(c *fiber.Ctx, sessionToken, csrfToken string) bool {
	if len(csrfToken) < 40 { // 32 (hex) + 8 (session part)
		return false
	}

	// Извлекаем часть сессии из CSRF токена
	sessionPart := csrfToken[32:] // Последние 8 символов
	expectedSessionPart := sessionToken[:8]

	return sessionPart == expectedSessionPart
}

// CookieAuthMiddleware работает с cookies вместо localStorage
func (h *Handler) CookieAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Сначала пробуем получить токен из cookie
		sessionToken := c.Cookies("session_token")

		// Если нет в cookie, пробуем заголовок (для обратной совместимости)
		if sessionToken == "" {
			sessionToken = h.extractSessionToken(c)
		}

		if sessionToken == "" {
			c.Locals("user", nil)
			c.Locals("isAuthenticated", false)
			return c.Next()
		}

		user, err := h.authUseCase.GetUserBySession(c.Context(), sessionToken)
		if err != nil {
			// Удаляем невалидный cookie
			c.Cookie(&fiber.Cookie{
				Name:     "session_token",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HTTPOnly: true,
				Secure:   true,
				SameSite: "Strict",
			})
			c.Locals("user", nil)
			c.Locals("isAuthenticated", false)
			return c.Next()
		}

		c.Locals("user", user)
		c.Locals("user_id", user.ID)
		c.Locals("isAuthenticated", true)
		return c.Next()
	}
}

// RequireAuthCookie требует авторизации через cookie
func (h *Handler) RequireAuthCookie() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionToken := c.Cookies("session_token")

		// Поддержка заголовка для API клиентов
		if sessionToken == "" {
			sessionToken = h.extractSessionToken(c)
		}

		if sessionToken == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		user, err := h.authUseCase.GetUserBySession(c.Context(), sessionToken)
		if err != nil {
			// Удаляем невалидный cookie
			c.Cookie(&fiber.Cookie{
				Name:     "session_token",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HTTPOnly: true,
				Secure:   true,
				SameSite: "Strict",
			})
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		c.Locals("user", user)
		c.Locals("user_id", user.ID)
		c.Locals("isAuthenticated", true)
		return c.Next()
	}
}
