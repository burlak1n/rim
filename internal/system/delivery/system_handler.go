package delivery

import (
	"net/http"

	"log/slog"

	systemUseCase "rim/internal/system/usecase"

	"github.com/gofiber/fiber/v2"
)

// Handler обрабатывает HTTP запросы для системных настроек
type Handler struct {
	systemUseCase systemUseCase.UseCase
	logger        *slog.Logger
}

// NewHandler создает новый экземпляр Handler для системных настроек
func NewHandler(systemUseCase systemUseCase.UseCase, logger *slog.Logger) *Handler {
	return &Handler{
		systemUseCase: systemUseCase,
		logger:        logger,
	}
}

// DebugModeResponse представляет ответ с состоянием отладочного режима
type DebugModeResponse struct {
	Enabled bool `json:"enabled"`
}

// DebugModeRequest представляет запрос на изменение отладочного режима
type DebugModeRequest struct {
	Enabled bool `json:"enabled"`
}

// GetDebugMode обрабатывает запрос на получение состояния отладочного режима
// @Summary Получить состояние отладочного режима
// @Description Возвращает текущее состояние отладочного режима системы
// @Tags system
// @Produce json
// @Success 200 {object} DebugModeResponse
// @Failure 500 {object} map[string]string
// @Router /system/debug-mode [get]
func (h *Handler) GetDebugMode(c *fiber.Ctx) error {
	enabled, err := h.systemUseCase.GetDebugMode(c.Context())
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to get debug mode", slog.Any("error", err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(DebugModeResponse{
		Enabled: enabled,
	})
}

// SetDebugMode обрабатывает запрос на изменение состояния отладочного режима
// @Summary Установить состояние отладочного режима
// @Description Изменяет состояние отладочного режима системы (только для администраторов)
// @Tags system
// @Accept json
// @Produce json
// @Param debug_mode body DebugModeRequest true "Новое состояние отладочного режима"
// @Success 200 {object} DebugModeResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /system/debug-mode [put]
func (h *Handler) SetDebugMode(c *fiber.Ctx) error {
	var req DebugModeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.systemUseCase.SetDebugMode(c.Context(), req.Enabled); err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to set debug mode", slog.Bool("enabled", req.Enabled), slog.Any("error", err))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(DebugModeResponse{
		Enabled: req.Enabled,
	})
}
