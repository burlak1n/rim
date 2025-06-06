package domain

import (
	"time"

	"gorm.io/gorm"
)

// SystemSetting представляет настройку системы
type SystemSetting struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Key       string         `gorm:"uniqueIndex;not null" json:"key"`
	Value     string         `gorm:"not null" json:"value"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName возвращает имя таблицы для SystemSetting
func (SystemSetting) TableName() string {
	return "system_settings"
}
