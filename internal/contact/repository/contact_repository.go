package repository

import (
	"context"
	"log/slog"

	"rim/internal/domain"

	"gorm.io/gorm"
)

// Repository определяет интерфейс для операций с данными контактов.
type Repository interface {
	Create(ctx context.Context, contact *domain.Contact) (*domain.Contact, error)
	GetByID(ctx context.Context, id uint) (*domain.Contact, error)
	GetByEmail(ctx context.Context, email string) (*domain.Contact, error)
	GetByPhone(ctx context.Context, phone string) (*domain.Contact, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Contact, error)
	GetByEmailUnscoped(ctx context.Context, email string) (*domain.Contact, error)
	GetByPhoneUnscoped(ctx context.Context, phone string) (*domain.Contact, error)
	GetAll(ctx context.Context) ([]domain.Contact, error)
	Update(ctx context.Context, contact *domain.Contact) error
	Delete(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error
	AddContactToGroup(ctx context.Context, contact *domain.Contact, group *domain.Group) error
	RemoveContactFromGroup(ctx context.Context, contact *domain.Contact, group *domain.Group) error
}

type sqliteRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewSQLiteRepository создает новый экземпляр sqliteRepository для контактов.
func NewSQLiteRepository(db *gorm.DB, logger *slog.Logger) Repository {
	return &sqliteRepository{
		db:     db,
		logger: logger,
	}
}

func (r *sqliteRepository) Create(ctx context.Context, contact *domain.Contact) (*domain.Contact, error) {
	// Возвращаем к простому созданию. GORM должен сам обработать уникальные индексы.
	// Проверки на существующие активные email/phone теперь полностью в usecase.
	if err := r.db.WithContext(ctx).Create(contact).Error; err != nil {
		r.logger.ErrorContext(ctx, "Error creating contact in DB", slog.Any("error", err), slog.String("contactName", contact.Name))
		return nil, err
	}
	r.logger.InfoContext(ctx, "Successfully created contact in DB", slog.Uint64("contactID", uint64(contact.ID)), slog.String("contactName", contact.Name))
	return contact, nil
}

func (r *sqliteRepository) GetByID(ctx context.Context, id uint) (*domain.Contact, error) {
	var contact domain.Contact
	// Загружаем связанные группы при получении контакта
	if err := r.db.WithContext(ctx).Preload("Groups").First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Contact not found by ID in DB", slog.Uint64("contactID", uint64(id)))
			return nil, err
		}
		r.logger.ErrorContext(ctx, "Error getting contact by ID from DB", slog.Uint64("contactID", uint64(id)), slog.Any("error", err))
		return nil, err
	}
	return &contact, nil
}

func (r *sqliteRepository) GetByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	var contact domain.Contact
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.InfoContext(ctx, "Contact not found by email in DB", slog.String("email", email))
			return nil, err
		}
		r.logger.ErrorContext(ctx, "Error getting contact by email from DB", slog.String("email", email), slog.Any("error", err))
		return nil, err
	}
	return &contact, nil
}

func (r *sqliteRepository) GetByPhone(ctx context.Context, phone string) (*domain.Contact, error) {
	var contact domain.Contact
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.InfoContext(ctx, "Contact not found by phone in DB", slog.String("phone", phone))
			return nil, err
		}
		r.logger.ErrorContext(ctx, "Error getting contact by phone from DB", slog.String("phone", phone), slog.Any("error", err))
		return nil, err
	}
	return &contact, nil
}

func (r *sqliteRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Contact, error) {
	var contact domain.Contact
	if err := r.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.InfoContext(ctx, "Contact not found by telegram ID in DB", slog.Int64("telegram_id", telegramID))
			return nil, err
		}
		r.logger.ErrorContext(ctx, "Error getting contact by telegram ID from DB", slog.Int64("telegram_id", telegramID), slog.Any("error", err))
		return nil, err
	}
	return &contact, nil
}

// GetAll извлекает все контакты (упрощенная версия).
func (r *sqliteRepository) GetAll(ctx context.Context) ([]domain.Contact, error) {
	var contacts []domain.Contact
	// Загружаем связанные группы для каждого контакта
	if err := r.db.WithContext(ctx).Preload("Groups").Find(&contacts).Error; err != nil {
		r.logger.ErrorContext(ctx, "Error getting all contacts from DB", slog.Any("error", err))
		return nil, err
	}
	return contacts, nil
}

func (r *sqliteRepository) Update(ctx context.Context, contact *domain.Contact) error {
	// При обновлении контакта важно также обновить его связи с группами.
	// GORM .Save() для структуры с ассоциациями many2many может потребовать явного управления ассоциациями,
	// если мы хотим добавлять/удалять группы в том же вызове Update.
	// Либо, если Groups в contact уже корректно установлены (например, загружены и изменены в usecase),
	// то .Save() может попытаться обновить их. Но это может быть сложно.
	//
	// Более явный подход - обновить поля самого контакта, а затем отдельно управлять ассоциациями.
	// Начнем с обновления только полей самого контакта.
	// Ассоциации будем менеджить через AddContactToGroup/RemoveContactFromGroup.

	tx := r.db.WithContext(ctx).Begin()

	// Обновляем основные поля контакта
	// Используем Select, чтобы обновить только указанные поля, исключая ассоциации из этого шага
	if err := tx.Select("Name", "Phone", "Email", "Transport", "Printer", "Allergies", "VK", "Telegram", "UpdatedAt").Updates(contact).Error; err != nil {
		tx.Rollback()
		r.logger.ErrorContext(ctx, "Error updating contact fields in DB", slog.Uint64("contactID", uint64(contact.ID)), slog.Any("error", err))
		return err
	}

	// Обновляем ассоциации (если переданы группы в contact.Groups)
	// Это заменит все существующие ассоциации на новые.
	if contact.Groups != nil { // Проверяем, переданы ли группы для обновления
		if err := tx.Model(contact).Association("Groups").Replace(contact.Groups); err != nil {
			tx.Rollback()
			r.logger.ErrorContext(ctx, "Error updating contact group associations in DB", slog.Uint64("contactID", uint64(contact.ID)), slog.Any("error", err))
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		r.logger.ErrorContext(ctx, "Error committing transaction for contact update", slog.Uint64("contactID", uint64(contact.ID)), slog.Any("error", err))
		return err
	}

	r.logger.InfoContext(ctx, "Successfully updated contact in DB", slog.Uint64("contactID", uint64(contact.ID)))
	return nil
}

func (r *sqliteRepository) Delete(ctx context.Context, id uint) error {
	// Мягкое удаление, GORM сам обработает DeletedAt
	// Также нужно учесть удаление связей в contact_groups. GORM должен это сделать автоматически при правильной настройке foreign keys и onDelete каскадов, либо это нужно делать явно.
	// Пока что просто удаляем контакт.
	result := r.db.WithContext(ctx).Delete(&domain.Contact{}, id)
	if result.Error != nil {
		r.logger.ErrorContext(ctx, "Error deleting contact from DB", slog.Uint64("contactID", uint64(id)), slog.Any("error", result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.logger.WarnContext(ctx, "Contact not found for deletion in DB", slog.Uint64("contactID", uint64(id)))
		return gorm.ErrRecordNotFound
	}
	r.logger.InfoContext(ctx, "Successfully marked contact as deleted in DB", slog.Uint64("contactID", uint64(id)))
	return nil
}

func (r *sqliteRepository) AddContactToGroup(ctx context.Context, contact *domain.Contact, group *domain.Group) error {
	if err := r.db.WithContext(ctx).Model(contact).Association("Groups").Append(group); err != nil {
		r.logger.ErrorContext(ctx, "Error adding contact to group in DB", slog.Uint64("contactID", uint64(contact.ID)), slog.Uint64("groupID", uint64(group.ID)), slog.Any("error", err))
		return err
	}
	r.logger.InfoContext(ctx, "Successfully added contact to group in DB", slog.Uint64("contactID", uint64(contact.ID)), slog.Uint64("groupID", uint64(group.ID)))
	return nil
}

func (r *sqliteRepository) RemoveContactFromGroup(ctx context.Context, contact *domain.Contact, group *domain.Group) error {
	if err := r.db.WithContext(ctx).Model(contact).Association("Groups").Delete(group); err != nil {
		r.logger.ErrorContext(ctx, "Error removing contact from group in DB", slog.Uint64("contactID", uint64(contact.ID)), slog.Uint64("groupID", uint64(group.ID)), slog.Any("error", err))
		return err
	}
	r.logger.InfoContext(ctx, "Successfully removed contact from group in DB", slog.Uint64("contactID", uint64(contact.ID)), slog.Uint64("groupID", uint64(group.ID)))
	return nil
}

func (r *sqliteRepository) GetByEmailUnscoped(ctx context.Context, email string) (*domain.Contact, error) {
	var contact domain.Contact
	if err := r.db.Unscoped().WithContext(ctx).Where("email = ?", email).First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.InfoContext(ctx, "Contact not found by email (unscoped) in DB", slog.String("email", email))
			return nil, err
		}
		r.logger.ErrorContext(ctx, "Error getting contact by email (unscoped) from DB", slog.String("email", email), slog.Any("error", err))
		return nil, err
	}
	return &contact, nil
}

func (r *sqliteRepository) GetByPhoneUnscoped(ctx context.Context, phone string) (*domain.Contact, error) {
	var contact domain.Contact
	if err := r.db.Unscoped().WithContext(ctx).Where("phone = ?", phone).First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.InfoContext(ctx, "Contact not found by phone (unscoped) in DB", slog.String("phone", phone))
			return nil, err
		}
		r.logger.ErrorContext(ctx, "Error getting contact by phone (unscoped) from DB", slog.String("phone", phone), slog.Any("error", err))
		return nil, err
	}
	return &contact, nil
}

func (r *sqliteRepository) HardDelete(ctx context.Context, id uint) error {
	result := r.db.Unscoped().WithContext(ctx).Delete(&domain.Contact{}, id)
	if result.Error != nil {
		r.logger.ErrorContext(ctx, "Error hard deleting contact from DB", slog.Uint64("contactID", uint64(id)), slog.Any("error", result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.logger.WarnContext(ctx, "Contact not found for hard deletion in DB", slog.Uint64("contactID", uint64(id)))
		// Не возвращаем ErrRecordNotFound, т.к. это не всегда ошибка в контексте hard delete (могли уже удалить)
	}
	r.logger.InfoContext(ctx, "Successfully hard deleted contact from DB", slog.Uint64("contactID", uint64(id)))
	return nil
}
