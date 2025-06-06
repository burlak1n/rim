package domain

import (
	"time"

	"gorm.io/gorm"
)

// Contact представляет модель контакта в системе.
// Содержит обязательные и необязательные поля, а также связь с группами.
type Contact struct {
	gorm.Model        // Включает ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string `gorm:"not null"`
	Phone      string `gorm:"not null;uniqueIndex"` // Телефон должен быть уникальным
	Email      string `gorm:"not null;uniqueIndex"` // Email должен быть уникальным

	// Необязательные поля
	Transport  string // "car", "license", "none"
	Printer    string // "color", "plain", "none"
	Allergies  string
	VK         string
	Telegram   string
	TelegramID int64 `gorm:"uniqueIndex"` // ID пользователя в Telegram

	Groups []*Group `gorm:"many2many:contact_groups;"` // Связь многие-ко-многим с группами
}

// User представляет авторизованного пользователя системы
type User struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	TelegramID int64     `json:"telegram_id" gorm:"uniqueIndex;not null"`
	ContactID  *uint     `json:"contact_id" gorm:"index"` // Связь с контактом
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Связь с контактом
	Contact *Contact `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
}

// UserSession представляет сессию пользователя в Redis
type UserSession struct {
	SessionToken string    `json:"session_token"`
	UserID       uint      `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiredAt    time.Time `json:"expired_at"`
}

// Group представляет модель группы контактов.
// Контакты могут принадлежать к нескольким группам.
type Group struct {
	gorm.Model        // Включает ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string `gorm:"not null;uniqueIndex"` // Название группы должно быть уникальным

	Contacts []*Contact `gorm:"many2many:contact_groups;"` // Связь многие-ко-многим с контактами
}

// TODO: Рассмотреть необходимость отдельных типов для Transport и Printer,
// например, enum-подобные константы, для улучшения типобезопасности и валидации.

// TODO: Добавить модель User для аутентификации и ролей.
