package delivery

import (
	groupDelivery "rim/internal/group/delivery"
	"time"
)

// CreateContactRequest определяет структуру для запроса на создание контакта.
type CreateContactRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=100"`
	Phone     string `json:"phone" validate:"required,e164"` // Или другой формат телефона
	Email     string `json:"email" validate:"required,email"`
	Transport string `json:"transport,omitempty" validate:"omitempty,oneof='есть машина' 'есть права' 'нет ничего'"`
	Printer   string `json:"printer,omitempty" validate:"omitempty,oneof='цветной' 'обычный' 'нет'"`
	Allergies string `json:"allergies,omitempty" validate:"omitempty,max=255"`
	VK        string `json:"vk,omitempty" validate:"omitempty,url"`            // Или более специфичная валидация для VK/TG
	Telegram  string `json:"telegram,omitempty" validate:"omitempty,alphanum"` // Пример: только буквы и цифры для username
	GroupIDs  []uint `json:"group_ids,omitempty"`
}

// UpdateContactRequest определяет структуру для запроса на обновление контакта.
// Используем указатели, чтобы различать пустые значения от непереданных.
type UpdateContactRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,e164"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	Transport *string `json:"transport,omitempty" validate:"omitempty,oneof='есть машина' 'есть права' 'нет ничего'"`
	Printer   *string `json:"printer,omitempty" validate:"omitempty,oneof='цветной' 'обычный' 'нет'"`
	Allergies *string `json:"allergies,omitempty" validate:"omitempty,max=255"`
	VK        *string `json:"vk,omitempty" validate:"omitempty,url"`
	Telegram  *string `json:"telegram,omitempty" validate:"omitempty,alphanum"`
	GroupIDs  *[]uint `json:"group_ids,omitempty"`
}

// ContactResponse определяет структуру для ответа с информацией о контакте.
type ContactResponse struct {
	ID        uint                          `json:"id"`
	Name      string                        `json:"name"`
	Phone     string                        `json:"phone"`
	Email     string                        `json:"email"`
	Transport string                        `json:"transport,omitempty"`
	Printer   string                        `json:"printer,omitempty"`
	Allergies string                        `json:"allergies,omitempty"`
	VK        string                        `json:"vk,omitempty"`
	Telegram  string                        `json:"telegram,omitempty"`
	Groups    []groupDelivery.GroupResponse `json:"groups,omitempty"`
	CreatedAt time.Time                     `json:"created_at"`
	UpdatedAt time.Time                     `json:"updated_at"`
}

// AddRemoveContactGroupRequest используется для запросов на добавление/удаление контакта из группы.
// Пока не используется, так как ID группы берется из URL.
// type AddRemoveContactGroupRequest struct {
// 	GroupID uint `json:"group_id" validate:"required"`
// }
