package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"rim/internal/domain"
	"rim/internal/group/repository"

	"gorm.io/gorm"
)

var (
	ErrGroupNameEmpty    = errors.New("group name cannot be empty")
	ErrGroupNotFound     = errors.New("group not found")
	ErrGroupNameExists   = errors.New("group with this name already exists")
	ErrCannotDeleteGroup = errors.New("cannot delete group") // Общая ошибка, может быть детализирована
)

// UseCase определяет интерфейс для бизнес-логики управления группами.
type UseCase interface {
	CreateGroup(ctx context.Context, name string) (*domain.Group, error)
	GetGroupByID(ctx context.Context, id uint) (*domain.Group, error)
	GetAllGroups(ctx context.Context) ([]domain.Group, error)
	UpdateGroup(ctx context.Context, id uint, newName string) (*domain.Group, error)
	DeleteGroup(ctx context.Context, id uint) error
}

type groupUseCase struct {
	groupRepo repository.Repository
	logger    *slog.Logger
}

// NewGroupUseCase создает новый экземпляр groupUseCase.
func NewGroupUseCase(groupRepo repository.Repository, logger *slog.Logger) UseCase {
	return &groupUseCase{
		groupRepo: groupRepo,
		logger:    logger,
	}
}

// CreateGroup создает новую группу.
func (uc *groupUseCase) CreateGroup(ctx context.Context, name string) (*domain.Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		uc.logger.WarnContext(ctx, "Attempt to create group with empty name")
		return nil, ErrGroupNameEmpty
	}

	// Проверяем, не существует ли группа с таким именем
	existingGroup, err := uc.groupRepo.GetByName(ctx, name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking for existing group by name", slog.String("name", name), slog.Any("error", err))
		return nil, err // Внутренняя ошибка сервера
	}
	if existingGroup != nil {
		uc.logger.WarnContext(ctx, "Attempt to create group with existing name", slog.String("name", name))
		return nil, ErrGroupNameExists
	}

	group := &domain.Group{Name: name}
	createdGroup, err := uc.groupRepo.Create(ctx, group)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create group via repository", slog.String("name", name), slog.Any("error", err))
		return nil, err // Внутренняя ошибка сервера
	}

	uc.logger.InfoContext(ctx, "Group created successfully", slog.Uint64("id", uint64(createdGroup.ID)), slog.String("name", createdGroup.Name))
	return createdGroup, nil
}

// GetGroupByID извлекает группу по ID.
func (uc *groupUseCase) GetGroupByID(ctx context.Context, id uint) (*domain.Group, error) {
	group, err := uc.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.WarnContext(ctx, "Group not found by ID", slog.Uint64("id", uint64(id)))
			return nil, ErrGroupNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting group by ID from repository", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return nil, err // Внутренняя ошибка сервера
	}
	return group, nil
}

// GetAllGroups извлекает все группы.
func (uc *groupUseCase) GetAllGroups(ctx context.Context) ([]domain.Group, error) {
	groups, err := uc.groupRepo.GetAll(ctx)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting all groups from repository", slog.Any("error", err))
		return nil, err // Внутренняя ошибка сервера
	}
	return groups, nil
}

// UpdateGroup обновляет существующую группу.
func (uc *groupUseCase) UpdateGroup(ctx context.Context, id uint, newName string) (*domain.Group, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		uc.logger.WarnContext(ctx, "Attempt to update group with empty name", slog.Uint64("id", uint64(id)))
		return nil, ErrGroupNameEmpty
	}

	groupToUpdate, err := uc.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.WarnContext(ctx, "Group to update not found by ID", slog.Uint64("id", uint64(id)))
			return nil, ErrGroupNotFound
		}
		uc.logger.ErrorContext(ctx, "Error fetching group to update from repository", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return nil, err // Внутренняя ошибка сервера
	}

	// Если имя не изменилось, ничего не делаем, возвращаем существующую группу
	if groupToUpdate.Name == newName {
		uc.logger.InfoContext(ctx, "Group name not changed, no update needed", slog.Uint64("id", uint64(id)), slog.String("name", newName))
		return groupToUpdate, nil
	}

	// Проверяем, не занято ли новое имя другой группой
	existingGroupWithNewName, err := uc.groupRepo.GetByName(ctx, newName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking for existing group by new name during update", slog.String("newName", newName), slog.Any("error", err))
		return nil, err // Внутренняя ошибка сервера
	}
	if existingGroupWithNewName != nil && existingGroupWithNewName.ID != id {
		uc.logger.WarnContext(ctx, "Attempt to update group name to an already existing name", slog.Uint64("id", uint64(id)), slog.String("newName", newName))
		return nil, ErrGroupNameExists
	}

	groupToUpdate.Name = newName
	if err := uc.groupRepo.Update(ctx, groupToUpdate); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update group via repository", slog.Uint64("id", uint64(id)), slog.String("newName", newName), slog.Any("error", err))
		return nil, err // Внутренняя ошибка сервера
	}

	uc.logger.InfoContext(ctx, "Group updated successfully", slog.Uint64("id", uint64(id)), slog.String("name", newName))
	return groupToUpdate, nil
}

// DeleteGroup удаляет группу по ID.
// TODO: Добавить логику проверки, что группа не используется (например, нет контактов в группе), если это требуется.
func (uc *groupUseCase) DeleteGroup(ctx context.Context, id uint) error {
	// Сначала проверим, существует ли группа
	_, err := uc.groupRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.WarnContext(ctx, "Group to delete not found by ID", slog.Uint64("id", uint64(id)))
			return ErrGroupNotFound
		}
		uc.logger.ErrorContext(ctx, "Error fetching group to delete from repository", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return err // Внутренняя ошибка сервера
	}

	if err := uc.groupRepo.Delete(ctx, id); err != nil {
		// gorm.ErrRecordNotFound может быть возвращен, если запись уже удалена или не найдена, что мы уже проверили выше.
		// Однако, оставим проверку на всякий случай, если логика репозитория изменится.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.WarnContext(ctx, "Group to delete not found by ID during deletion attempt", slog.Uint64("id", uint64(id)))
			return ErrGroupNotFound // Повторная проверка, но лучше быть уверенным
		}
		uc.logger.ErrorContext(ctx, "Failed to delete group via repository", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return ErrCannotDeleteGroup // Используем нашу общую ошибку или можно вернуть err
	}

	uc.logger.InfoContext(ctx, "Group deleted successfully", slog.Uint64("id", uint64(id)))
	return nil
}
