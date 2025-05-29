package domain

import (
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
	Transport string // "car", "license", "none"
	Printer   string // "color", "plain", "none"
	Allergies string
	VK        string
	Telegram  string

	Groups []*Group `gorm:"many2many:contact_groups;"` // Связь многие-ко-многим с группами
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
