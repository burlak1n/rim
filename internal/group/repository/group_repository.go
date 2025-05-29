package repository

import (
	"context"
	"log/slog"

	"rim/internal/domain"

	"gorm.io/gorm"
)

// Repository определяет интерфейс для операций с данными групп.
// Это позволяет абстрагироваться от конкретной реализации хранилища.
type Repository interface {
	Create(ctx context.Context, group *domain.Group) (*domain.Group, error)
	GetByID(ctx context.Context, id uint) (*domain.Group, error)
	GetByName(ctx context.Context, name string) (*domain.Group, error)
	GetAll(ctx context.Context) ([]domain.Group, error)
	Update(ctx context.Context, group *domain.Group) error
	Delete(ctx context.Context, id uint) error
}

// sqliteRepository реализует Repository для работы с SQLite через GORM.
type sqliteRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewSQLiteRepository создает новый экземпляр sqliteRepository.
func NewSQLiteRepository(db *gorm.DB, logger *slog.Logger) Repository {
	return &sqliteRepository{
		db:     db,
		logger: logger,
	}
}

// Create создает новую группу в базе данных.
func (r *sqliteRepository) Create(ctx context.Context, group *domain.Group) (*domain.Group, error) {
	if err := r.db.WithContext(ctx).Create(group).Error; err != nil {
		r.logger.ErrorContext(ctx, "Error creating group in DB", slog.Any("error", err), slog.String("groupName", group.Name))
		return nil, err
	}
	r.logger.InfoContext(ctx, "Successfully created group in DB", slog.Uint64("groupID", uint64(group.ID)), slog.String("groupName", group.Name))
	return group, nil
}

// GetByID извлекает группу по ее ID.
func (r *sqliteRepository) GetByID(ctx context.Context, id uint) (*domain.Group, error) {
	var group domain.Group
	if err := r.db.WithContext(ctx).First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnContext(ctx, "Group not found by ID in DB", slog.Uint64("groupID", uint64(id)))
			return nil, err // Возвращаем gorm.ErrRecordNotFound как есть
		}
		r.logger.ErrorContext(ctx, "Error getting group by ID from DB", slog.Uint64("groupID", uint64(id)), slog.Any("error", err))
		return nil, err
	}
	return &group, nil
}

// GetByName извлекает группу по ее имени.
func (r *sqliteRepository) GetByName(ctx context.Context, name string) (*domain.Group, error) {
	var group domain.Group
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.InfoContext(ctx, "Group not found by name in DB", slog.String("groupName", name)) // Info, т.к. это ожидаемое поведение при проверке уникальности
			return nil, err                                                                            // Возвращаем gorm.ErrRecordNotFound как есть
		}
		r.logger.ErrorContext(ctx, "Error getting group by name from DB", slog.String("groupName", name), slog.Any("error", err))
		return nil, err
	}
	return &group, nil
}

// GetAll извлекает все группы из базы данных.
func (r *sqliteRepository) GetAll(ctx context.Context) ([]domain.Group, error) {
	var groups []domain.Group
	if err := r.db.WithContext(ctx).Find(&groups).Error; err != nil {
		r.logger.ErrorContext(ctx, "Error getting all groups from DB", slog.Any("error", err))
		return nil, err
	}
	return groups, nil
}

// Update обновляет данные существующей группы.
func (r *sqliteRepository) Update(ctx context.Context, group *domain.Group) error {
	// Убедимся, что группа существует перед обновлением
	// GORM Save обновит все поля или создаст новую запись, если ID 0. Нам нужно именно обновление.
	// Используем Model(&domain.Group{}).Where("id = ?", group.ID).Updates(group) для частичного обновления
	// или просто Save, если хотим обновлять все поля структуры.
	// Для простоты начнем с Save, но учитываем, что он обновит все поля, включая CreatedAt, если не обработать это.
	// Правильнее было бы использовать Updates с мапой или структурой только обновляемых полей.
	// Пока для простоты оставим Save, предполагая, что передается полная обновленная модель.
	result := r.db.WithContext(ctx).Save(group)
	if result.Error != nil {
		r.logger.ErrorContext(ctx, "Error updating group in DB", slog.Uint64("groupID", uint64(group.ID)), slog.Any("error", result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.logger.WarnContext(ctx, "Group not found for update in DB or no changes made", slog.Uint64("groupID", uint64(group.ID)))
		return gorm.ErrRecordNotFound // Или другая специфичная ошибка, что запись не найдена для обновления
	}
	r.logger.InfoContext(ctx, "Successfully updated group in DB", slog.Uint64("groupID", uint64(group.ID)))
	return nil
}

// Delete удаляет группу по ее ID.
func (r *sqliteRepository) Delete(ctx context.Context, id uint) error {
	// GORM использует мягкое удаление по умолчанию, если в модели есть gorm.DeletedAt
	// Это установит поле DeletedAt, а не удалит запись физически.
	result := r.db.WithContext(ctx).Delete(&domain.Group{}, id)
	if result.Error != nil {
		r.logger.ErrorContext(ctx, "Error deleting group from DB", slog.Uint64("groupID", uint64(id)), slog.Any("error", result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.logger.WarnContext(ctx, "Group not found for deletion in DB", slog.Uint64("groupID", uint64(id)))
		return gorm.ErrRecordNotFound // Запись не найдена для удаления
	}
	r.logger.InfoContext(ctx, "Successfully marked group as deleted in DB", slog.Uint64("groupID", uint64(id)))
	return nil
}
