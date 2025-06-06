package delivery

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	contactUseCase "rim/internal/contact/usecase"
	"rim/internal/domain"
	groupDelivery "rim/internal/group/delivery" // Для ErrorResponse и GroupResponse
	groupUseCase "rim/internal/group/usecase"
)

// Handler отвечает за обработку HTTP-запросов, связанных с контактами.
type Handler struct {
	contactUseCase contactUseCase.UseCase
	logger         *slog.Logger
	validate       *validator.Validate
}

// NewHandler создает новый экземпляр Handler для контактов.
func NewHandler(cu contactUseCase.UseCase, logger *slog.Logger) *Handler {
	return &Handler{
		contactUseCase: cu,
		logger:         logger,
		validate:       validator.New(),
	}
}

// CreateContact обрабатывает запрос на создание нового контакта.
// @Summary Создать новый контакт
// @Description Создает новый контакт с указанными данными и опционально добавляет в группы.
// @Tags contacts
// @Accept json
// @Produce json
// @Param contact body CreateContactRequest true "Данные для создания контакта"
// @Success 201 {object} ContactResponse "Контакт успешно создан"
// @Failure 400 {object} groupDelivery.ErrorResponse "Ошибка валидации или некорректный запрос"
// @Failure 404 {object} groupDelivery.ErrorResponse "Одна из указанных групп не найдена"
// @Failure 409 {object} groupDelivery.ErrorResponse "Контакт с таким email или телефоном уже существует"
// @Failure 500 {object} groupDelivery.ErrorResponse "Внутренняя ошибка сервера"
// @Router /contacts [post]
func (h *Handler) CreateContact(c *fiber.Ctx) error {
	var req CreateContactRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnContext(c.Context(), "Failed to parse request body for create contact", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.WarnContext(c.Context(), "Validation failed for create contact request", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: fmt.Sprintf("Validation failed: %s", err.Error())})
	}

	ucData := contactUseCase.CreateContactData{
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		Transport: req.Transport,
		Printer:   req.Printer,
		Allergies: req.Allergies,
		VK:        req.VK,
		Telegram:  req.Telegram,
		GroupIDs:  req.GroupIDs,
	}

	contact, err := h.contactUseCase.CreateContact(c.Context(), ucData)
	if err != nil {
		if errors.Is(err, contactUseCase.ErrContactNameEmpty) || errors.Is(err, contactUseCase.ErrContactPhoneEmpty) || errors.Is(err, contactUseCase.ErrContactEmailEmpty) {
			return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		if errors.Is(err, contactUseCase.ErrContactEmailExists) || errors.Is(err, contactUseCase.ErrContactPhoneExists) {
			return c.Status(fiber.StatusConflict).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		if errors.Is(err, groupUseCase.ErrGroupNotFound) { // Ошибка от contactUseCase, если группа не найдена
			return c.Status(fiber.StatusNotFound).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		h.logger.ErrorContext(c.Context(), "Failed to create contact via use case", slog.Any("request", req), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(toContactResponse(contact))
}

// GetContactByID обрабатывает запрос на получение контакта по ID.
// @Summary Получить контакт по ID
// @Description Возвращает информацию о контакте, включая группы, в которых он состоит.
// @Tags contacts
// @Produce json
// @Param id path int true "ID контакта"
// @Success 200 {object} ContactResponse "Информация о контакте"
// @Failure 400 {object} groupDelivery.ErrorResponse "Некорректный ID"
// @Failure 404 {object} groupDelivery.ErrorResponse "Контакт не найден"
// @Failure 500 {object} groupDelivery.ErrorResponse "Внутренняя ошибка сервера"
// @Router /contacts/{id} [get]
func (h *Handler) GetContactByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	contactID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid contact ID format"})
	}

	contact, err := h.contactUseCase.GetContactByID(c.Context(), uint(contactID))
	if err != nil {
		if errors.Is(err, contactUseCase.ErrContactNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		h.logger.ErrorContext(c.Context(), "Failed to get contact by ID from use case", slog.Uint64("id", contactID), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: "Internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(toContactResponse(contact))
}

// GetAllContacts обрабатывает запрос на получение всех контактов.
// @Summary Получить все контакты
// @Description Возвращает список всех контактов. Для неавторизованных пользователей возвращает только имена.
// @Tags contacts
// @Produce json
// @Success 200 {array} ContactResponse "Список контактов для авторизованных пользователей"
// @Success 200 {array} ContactBasicResponse "Список контактов для неавторизованных пользователей"
// @Failure 500 {object} groupDelivery.ErrorResponse "Внутренняя ошибка сервера"
// @Router /contacts [get]
func (h *Handler) GetAllContacts(c *fiber.Ctx) error {
	contacts, err := h.contactUseCase.GetAllContacts(c.Context())
	if err != nil {
		h.logger.ErrorContext(c.Context(), "Failed to get all contacts from use case", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: "Internal server error"})
	}

	// Проверяем авторизацию пользователя
	isAuthenticated := c.Locals("isAuthenticated")
	isAuth := false
	if isAuthenticated != nil {
		if isAuthBool, ok := isAuthenticated.(bool); ok {
			isAuth = isAuthBool
		}
	}

	if isAuth {
		// Возвращаем полную информацию для авторизованных пользователей
		resp := make([]ContactResponse, len(contacts))
		for i, ct := range contacts {
			resp[i] = toContactResponse(&ct)
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	} else {
		// Возвращаем только имена для неавторизованных пользователей
		resp := make([]ContactBasicResponse, len(contacts))
		for i, ct := range contacts {
			resp[i] = ContactBasicResponse{
				ID:   ct.ID,
				Name: ct.Name,
			}
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// UpdateContact обрабатывает запрос на обновление контакта.
// @Summary Обновить контакт
// @Description Обновляет данные контакта и/или список групп, в которых он состоит.
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path int true "ID контакта для обновления"
// @Param contact body UpdateContactRequest true "Данные для обновления контакта"
// @Success 200 {object} ContactResponse "Контакт успешно обновлен"
// @Failure 400 {object} groupDelivery.ErrorResponse "Ошибка валидации, некорректный ID или некорректный запрос"
// @Failure 404 {object} groupDelivery.ErrorResponse "Контакт или одна из указанных групп не найдена"
// @Failure 409 {object} groupDelivery.ErrorResponse "Конфликт данных (например, email или телефон уже занят)"
// @Failure 500 {object} groupDelivery.ErrorResponse "Внутренняя ошибка сервера"
// @Router /contacts/{id} [put]
func (h *Handler) UpdateContact(c *fiber.Ctx) error {
	idStr := c.Params("id")
	contactID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid contact ID format"})
	}

	var req UpdateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid request body"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: fmt.Sprintf("Validation failed: %s", err.Error())})
	}

	ucData := contactUseCase.UpdateContactData{
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		Transport: req.Transport,
		Printer:   req.Printer,
		Allergies: req.Allergies,
		VK:        req.VK,
		Telegram:  req.Telegram,
		GroupIDs:  req.GroupIDs,
	}

	updatedContact, err := h.contactUseCase.UpdateContact(c.Context(), uint(contactID), ucData)
	if err != nil {
		if errors.Is(err, contactUseCase.ErrContactNotFound) || errors.Is(err, groupUseCase.ErrGroupNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		if errors.Is(err, contactUseCase.ErrContactNameEmpty) || errors.Is(err, contactUseCase.ErrContactPhoneEmpty) || errors.Is(err, contactUseCase.ErrContactEmailEmpty) {
			return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		if errors.Is(err, contactUseCase.ErrContactEmailExists) || errors.Is(err, contactUseCase.ErrContactPhoneExists) {
			return c.Status(fiber.StatusConflict).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		h.logger.ErrorContext(c.Context(), "Failed to update contact via use case", slog.Uint64("id", contactID), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(toContactResponse(updatedContact))
}

// DeleteContact обрабатывает запрос на удаление контакта.
// @Summary Удалить контакт
// @Description Удаляет контакт по его ID.
// @Tags contacts
// @Produce json
// @Param id path int true "ID контакта для удаления"
// @Success 204 "Контакт успешно удален"
// @Failure 400 {object} groupDelivery.ErrorResponse "Некорректный ID"
// @Failure 404 {object} groupDelivery.ErrorResponse "Контакт не найден"
// @Failure 500 {object} groupDelivery.ErrorResponse "Внутренняя ошибка сервера"
// @Router /contacts/{id} [delete]
func (h *Handler) DeleteContact(c *fiber.Ctx) error {
	idStr := c.Params("id")
	contactID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid contact ID format"})
	}

	if err := h.contactUseCase.DeleteContact(c.Context(), uint(contactID)); err != nil {
		if errors.Is(err, contactUseCase.ErrContactNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		h.logger.ErrorContext(c.Context(), "Failed to delete contact via use case", slog.Uint64("id", contactID), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: "Internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// AddContactToGroup добавляет контакт в группу.
// @Summary Добавить контакт в группу
// @Description Добавляет существующий контакт в существующую группу.
// @Tags contacts
// @Produce json
// @Param contact_id path int true "ID контакта"
// @Param group_id path int true "ID группы"
// @Success 204 "Контакт успешно добавлен в группу"
// @Failure 400 {object} groupDelivery.ErrorResponse "Некорректный ID контакта или группы"
// @Failure 404 {object} groupDelivery.ErrorResponse "Контакт или группа не найдены"
// @Failure 500 {object} groupDelivery.ErrorResponse "Внутренняя ошибка сервера"
// @Router /contacts/{contact_id}/groups/{group_id} [post]
func (h *Handler) AddContactToGroup(c *fiber.Ctx) error {
	contactIDStr := c.Params("contact_id")
	contactID, err := strconv.ParseUint(contactIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid contact ID format"})
	}

	groupIDStr := c.Params("group_id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid group ID format"})
	}

	err = h.contactUseCase.AddContactToGroup(c.Context(), uint(contactID), uint(groupID))
	if err != nil {
		if errors.Is(err, contactUseCase.ErrContactNotFound) || errors.Is(err, groupUseCase.ErrGroupNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		if errors.Is(err, contactUseCase.ErrGroupAssociation) { // Ошибка при ассоциации
			return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		h.logger.ErrorContext(c.Context(), "Failed to add contact to group", slog.Uint64("contactID", contactID), slog.Uint64("groupID", groupID), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: "Internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// RemoveContactFromGroup удаляет контакт из группы.
// @Summary Удалить контакт из группы
// @Description Удаляет существующий контакт из существующей группы.
// @Tags contacts
// @Produce json
// @Param contact_id path int true "ID контакта"
// @Param group_id path int true "ID группы"
// @Success 204 "Контакт успешно удален из группы"
// @Failure 400 {object} groupDelivery.ErrorResponse "Некорректный ID контакта или группы, или контакт не в группе"
// @Failure 404 {object} groupDelivery.ErrorResponse "Контакт или группа не найдены"
// @Failure 500 {object} groupDelivery.ErrorResponse "Внутренняя ошибка сервера"
// @Router /contacts/{contact_id}/groups/{group_id} [delete]
func (h *Handler) RemoveContactFromGroup(c *fiber.Ctx) error {
	contactIDStr := c.Params("contact_id")
	contactID, err := strconv.ParseUint(contactIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid contact ID format"})
	}

	groupIDStr := c.Params("group_id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: "Invalid group ID format"})
	}

	err = h.contactUseCase.RemoveContactFromGroup(c.Context(), uint(contactID), uint(groupID))
	if err != nil {
		if errors.Is(err, contactUseCase.ErrContactNotFound) || errors.Is(err, groupUseCase.ErrGroupNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		// Если usecase возвращает ошибку, что контакт не в группе, это BadRequest
		if e, ok := err.(interface{ Error() string }); ok && e.Error() == fmt.Sprintf("contact is not a member of group %d", groupID) {
			return c.Status(fiber.StatusBadRequest).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		if errors.Is(err, contactUseCase.ErrGroupAssociation) { // Ошибка при диссоциации
			return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: err.Error()})
		}
		h.logger.ErrorContext(c.Context(), "Failed to remove contact from group", slog.Uint64("contactID", contactID), slog.Uint64("groupID", groupID), slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(groupDelivery.ErrorResponse{Message: "Internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// toContactResponse преобразует domain.Contact в ContactResponse DTO.
func toContactResponse(contact *domain.Contact) ContactResponse {
	grRes := make([]groupDelivery.GroupResponse, len(contact.Groups))
	for i, g := range contact.Groups {
		grRes[i] = groupDelivery.GroupResponse{
			ID:        g.ID,
			Name:      g.Name,
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
		}
	}
	return ContactResponse{
		ID:        contact.ID,
		Name:      contact.Name,
		Phone:     contact.Phone,
		Email:     contact.Email,
		Transport: contact.Transport,
		Printer:   contact.Printer,
		Allergies: contact.Allergies,
		VK:        contact.VK,
		Telegram:  contact.Telegram,
		Groups:    grRes,
		CreatedAt: contact.CreatedAt,
		UpdatedAt: contact.UpdatedAt,
	}
}
