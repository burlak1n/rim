package delivery

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"rim/internal/domain"
	"rim/internal/group/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Handler отвечает за обработку HTTP-запросов, связанных с группами.
type Handler struct {
	groupUseCase usecase.UseCase
	logger       *slog.Logger
	validate     *validator.Validate
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(groupUC usecase.UseCase, logger *slog.Logger) *Handler {
	return &Handler{
		groupUseCase: groupUC,
		logger:       logger,
		validate:     validator.New(),
	}
}

// CreateGroup обрабатывает запрос на создание новой группы.
// @Summary Создать новую группу
// @Description Создает новую группу с указанным именем.
// @Tags groups
// @Accept json
// @Produce json
// @Param group body CreateGroupRequest true "Данные для создания группы"
// @Success 201 {object} GroupResponse "Группа успешно создана"
// @Failure 400 {object} ErrorResponse "Ошибка валидации или некорректный запрос"
// @Failure 409 {object} ErrorResponse "Группа с таким именем уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /groups [post]
func (h *Handler) CreateGroup(c *fiber.Ctx) error {
	var req CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("Failed to parse request body for create group", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Message: "Invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Warn("Validation failed for create group request", slog.Any("error", err))
		// Можно вернуть более детализированные ошибки валидации
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Message: fmt.Sprintf("Validation failed: %s", err.Error())})
	}

	group, err := h.groupUseCase.CreateGroup(c.Context(), req.Name)
	if err != nil {
		if errors.Is(err, usecase.ErrGroupNameEmpty) || errors.Is(err, usecase.ErrGroupNameExists) {
			h.logger.Warn("Failed to create group due to business rule violation", slog.String("name", req.Name), slog.Any("error", err))
			status := fiber.StatusBadRequest
			if errors.Is(err, usecase.ErrGroupNameExists) {
				status = fiber.StatusConflict
			}
			return c.Status(status).JSON(ErrorResponse{Message: err.Error()})
		}
		h.logger.Error("Failed to create group via use case", slog.String("name", req.Name), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Message: "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(toGroupResponse(group))
}

// GetGroupByID обрабатывает запрос на получение группы по ID.
// @Summary Получить группу по ID
// @Description Возвращает информацию о группе по ее уникальному идентификатору.
// @Tags groups
// @Produce json
// @Param id path int true "ID группы"
// @Success 200 {object} GroupResponse "Информация о группе"
// @Failure 400 {object} ErrorResponse "Некорректный ID"
// @Failure 404 {object} ErrorResponse "Группа не найдена"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /groups/{id} [get]
func (h *Handler) GetGroupByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid group ID format", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Message: "Invalid group ID format"})
	}

	group, err := h.groupUseCase.GetGroupByID(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, usecase.ErrGroupNotFound) {
			h.logger.Warn("Group not found by ID in handler", slog.Uint64("id", id))
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Message: err.Error()})
		}
		h.logger.Error("Failed to get group by ID from use case", slog.Uint64("id", id), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Message: "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(toGroupResponse(group))
}

// GetAllGroups обрабатывает запрос на получение всех групп.
// @Summary Получить все группы
// @Description Возвращает список всех существующих групп.
// @Tags groups
// @Produce json
// @Success 200 {array} GroupResponse "Список групп"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /groups [get]
func (h *Handler) GetAllGroups(c *fiber.Ctx) error {
	groups, err := h.groupUseCase.GetAllGroups(c.Context())
	if err != nil {
		h.logger.Error("Failed to get all groups from use case", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Message: "Internal server error"})
	}

	resp := make([]GroupResponse, len(groups))
	for i, g := range groups {
		resp[i] = toGroupResponse(&g)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

// UpdateGroup обрабатывает запрос на обновление существующей группы.
// @Summary Обновить группу
// @Description Обновляет имя существующей группы по ее ID.
// @Tags groups
// @Accept json
// @Produce json
// @Param id path int true "ID группы для обновления"
// @Param group body UpdateGroupRequest true "Новое имя для группы"
// @Success 200 {object} GroupResponse "Группа успешно обновлена"
// @Failure 400 {object} ErrorResponse "Ошибка валидации, некорректный ID или некорректный запрос"
// @Failure 404 {object} ErrorResponse "Группа не найдена"
// @Failure 409 {object} ErrorResponse "Группа с таким новым именем уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /groups/{id} [put]
func (h *Handler) UpdateGroup(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid group ID format for update", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Message: "Invalid group ID format"})
	}

	var req UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("Failed to parse request body for update group", slog.Uint64("id", id), slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Message: "Invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Warn("Validation failed for update group request", slog.Uint64("id", id), slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Message: fmt.Sprintf("Validation failed: %s", err.Error())})
	}

	updatedGroup, err := h.groupUseCase.UpdateGroup(c.Context(), uint(id), req.Name)
	if err != nil {
		if errors.Is(err, usecase.ErrGroupNotFound) {
			h.logger.Warn("Group not found for update in handler", slog.Uint64("id", id), slog.String("newName", req.Name))
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Message: err.Error()})
		}
		if errors.Is(err, usecase.ErrGroupNameEmpty) || errors.Is(err, usecase.ErrGroupNameExists) {
			status := fiber.StatusBadRequest
			if errors.Is(err, usecase.ErrGroupNameExists) {
				status = fiber.StatusConflict
			}
			h.logger.Warn("Failed to update group due to business rule violation", slog.Uint64("id", id), slog.String("newName", req.Name), slog.Any("error", err))
			return c.Status(status).JSON(ErrorResponse{Message: err.Error()})
		}
		h.logger.Error("Failed to update group via use case", slog.Uint64("id", id), slog.String("newName", req.Name), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Message: "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(toGroupResponse(updatedGroup))
}

// DeleteGroup обрабатывает запрос на удаление группы.
// @Summary Удалить группу
// @Description Удаляет группу по ее уникальному идентификатору.
// @Tags groups
// @Produce json
// @Param id path int true "ID группы для удаления"
// @Success 204 "Группа успешно удалена (нет содержимого)"
// @Failure 400 {object} ErrorResponse "Некорректный ID"
// @Failure 404 {object} ErrorResponse "Группа не найдена"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /groups/{id} [delete]
func (h *Handler) DeleteGroup(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid group ID format for delete", slog.String("id", idStr), slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Message: "Invalid group ID format"})
	}

	if err := h.groupUseCase.DeleteGroup(c.Context(), uint(id)); err != nil {
		if errors.Is(err, usecase.ErrGroupNotFound) {
			h.logger.Warn("Group not found for delete in handler", slog.Uint64("id", id))
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Message: err.Error()})
		}
		// ErrCannotDeleteGroup также может быть здесь, если use case его возвращает
		h.logger.Error("Failed to delete group via use case", slog.Uint64("id", id), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Message: "Internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// toGroupResponse преобразует domain.Group в GroupResponse DTO.
func toGroupResponse(group *domain.Group) GroupResponse {
	return GroupResponse{
		ID:        group.ID,
		Name:      group.Name,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
}
