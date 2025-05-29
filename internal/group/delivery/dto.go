package delivery

import "time"

// CreateGroupRequest определяет структуру для запроса на создание группы.
type CreateGroupRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"` // Добавили валидацию
}

// UpdateGroupRequest определяет структуру для запроса на обновление группы.
type UpdateGroupRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"` // Добавили валидацию
}

// GroupResponse определяет структуру для ответа с информацией о группе.
type GroupResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorResponse определяет общую структуру для ответа с ошибкой.
type ErrorResponse struct {
	Message string `json:"message"`
}
