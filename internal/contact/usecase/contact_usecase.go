package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	contactRepo "rim/internal/contact/repository"
	"rim/internal/domain"
	groupRepo "rim/internal/group/repository"
	groupUseCase "rim/internal/group/usecase" // Для ошибок ErrGroupNotFound

	"gorm.io/gorm"
)

var (
	ErrContactNotFound    = errors.New("contact not found")
	ErrContactNameEmpty   = errors.New("contact name cannot be empty")
	ErrContactPhoneEmpty  = errors.New("contact phone cannot be empty")
	ErrContactEmailEmpty  = errors.New("contact email cannot be empty")
	ErrContactPhoneExists = errors.New("contact with this phone already exists")
	ErrContactEmailExists = errors.New("contact with this email already exists")
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrInvalidPhoneFormat = errors.New("invalid phone format") // Может понадобиться более сложная валидация
	ErrGroupAssociation   = errors.New("error associating contact with group")
)

// CreateContactData определяет данные для создания нового контакта.
type CreateContactData struct {
	Name      string
	Phone     string
	Email     string
	Transport string
	Printer   string
	Allergies string
	VK        string
	Telegram  string
	GroupIDs  []uint // ID групп, к которым нужно добавить контакт
}

// UpdateContactData определяет данные для обновления существующего контакта.
type UpdateContactData struct {
	Name      *string // Указатели, чтобы различать пустые значения и отсутствующие в запросе
	Phone     *string
	Email     *string
	Transport *string
	Printer   *string
	Allergies *string
	VK        *string
	Telegram  *string
	GroupIDs  *[]uint // Список ID групп для полной замены существующих связей
}

// UseCase определяет интерфейс для бизнес-логики управления контактами.
type UseCase interface {
	CreateContact(ctx context.Context, data CreateContactData) (*domain.Contact, error)
	GetContactByID(ctx context.Context, id uint) (*domain.Contact, error)
	GetAllContacts(ctx context.Context) ([]domain.Contact, error)
	UpdateContact(ctx context.Context, id uint, data UpdateContactData) (*domain.Contact, error)
	DeleteContact(ctx context.Context, id uint) error
	AddContactToGroup(ctx context.Context, contactID uint, groupID uint) error
	RemoveContactFromGroup(ctx context.Context, contactID uint, groupID uint) error
}

type contactUseCase struct {
	contactRepo contactRepo.Repository
	groupRepo   groupRepo.Repository // Нужен для проверки существования групп
	logger      *slog.Logger
}

// NewContactUseCase создает новый экземпляр contactUseCase.
func NewContactUseCase(cr contactRepo.Repository, gr groupRepo.Repository, logger *slog.Logger) UseCase {
	return &contactUseCase{
		contactRepo: cr,
		groupRepo:   gr,
		logger:      logger,
	}
}

func (uc *contactUseCase) CreateContact(ctx context.Context, data CreateContactData) (*domain.Contact, error) {
	data.Name = strings.TrimSpace(data.Name)
	data.Phone = strings.TrimSpace(data.Phone)
	data.Email = strings.TrimSpace(data.Email)

	if data.Name == "" {
		return nil, ErrContactNameEmpty
	}
	if data.Phone == "" {
		return nil, ErrContactPhoneEmpty
	}
	if data.Email == "" {
		return nil, ErrContactEmailEmpty
	}
	// TODO: Добавить более строгую валидацию формата Email и Phone

	// Проверка уникальности Email
	existingByEmail, err := uc.contactRepo.GetByEmail(ctx, data.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking contact email existence", slog.String("email", data.Email), slog.Any("error", err))
		return nil, err
	}
	if existingByEmail != nil {
		return nil, ErrContactEmailExists
	}

	// Проверка уникальности Phone
	existingByPhone, err := uc.contactRepo.GetByPhone(ctx, data.Phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		uc.logger.ErrorContext(ctx, "Error checking contact phone existence", slog.String("phone", data.Phone), slog.Any("error", err))
		return nil, err
	}
	if existingByPhone != nil {
		return nil, ErrContactPhoneExists
	}

	contact := &domain.Contact{
		Name:      data.Name,
		Phone:     data.Phone,
		Email:     data.Email,
		Transport: data.Transport,
		Printer:   data.Printer,
		Allergies: data.Allergies,
		VK:        data.VK,
		Telegram:  data.Telegram,
	}

	// Проверка и подготовка групп
	if len(data.GroupIDs) > 0 {
		groups := make([]*domain.Group, 0, len(data.GroupIDs))
		for _, groupID := range data.GroupIDs {
			group, err := uc.groupRepo.GetByID(ctx, groupID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					uc.logger.WarnContext(ctx, "Group not found for contact association during create", slog.Uint64("groupID", uint64(groupID)))
					return nil, fmt.Errorf("%w: group with id %d not found", groupUseCase.ErrGroupNotFound, groupID)
				}
				uc.logger.ErrorContext(ctx, "Error fetching group for contact association", slog.Uint64("groupID", uint64(groupID)), slog.Any("error", err))
				return nil, err
			}
			groups = append(groups, group)
		}
		contact.Groups = groups
	}

	createdContact, err := uc.contactRepo.Create(ctx, contact)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create contact via repository", slog.String("name", contact.Name), slog.Any("error", err))
		return nil, err
	}

	uc.logger.InfoContext(ctx, "Contact created successfully", slog.Uint64("id", uint64(createdContact.ID)))
	return createdContact, nil
}

func (uc *contactUseCase) GetContactByID(ctx context.Context, id uint) (*domain.Contact, error) {
	contact, err := uc.contactRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.WarnContext(ctx, "Contact not found by ID in usecase", slog.Uint64("id", uint64(id)))
			return nil, ErrContactNotFound
		}
		uc.logger.ErrorContext(ctx, "Error getting contact by ID from repository", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return nil, err
	}
	return contact, nil
}

func (uc *contactUseCase) GetAllContacts(ctx context.Context) ([]domain.Contact, error) {
	contacts, err := uc.contactRepo.GetAll(ctx)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Error getting all contacts from repository", slog.Any("error", err))
		return nil, err
	}
	return contacts, nil
}

func (uc *contactUseCase) UpdateContact(ctx context.Context, id uint, data UpdateContactData) (*domain.Contact, error) {
	contactToUpdate, err := uc.contactRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.logger.WarnContext(ctx, "Contact to update not found", slog.Uint64("id", uint64(id)))
			return nil, ErrContactNotFound
		}
		uc.logger.ErrorContext(ctx, "Error fetching contact to update", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return nil, err
	}

	// Обновляем поля, если они переданы
	changed := false
	if data.Name != nil {
		name := strings.TrimSpace(*data.Name)
		if name == "" {
			return nil, ErrContactNameEmpty
		}
		if contactToUpdate.Name != name {
			contactToUpdate.Name = name
			changed = true
		}
	}
	if data.Email != nil {
		email := strings.TrimSpace(*data.Email)
		if email == "" {
			return nil, ErrContactEmailEmpty
		}
		if contactToUpdate.Email != email {
			// Проверка уникальности нового Email
			existingByEmail, err := uc.contactRepo.GetByEmail(ctx, email)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				uc.logger.ErrorContext(ctx, "Error checking new contact email existence", slog.String("email", email), slog.Any("error", err))
				return nil, err
			}
			if existingByEmail != nil && existingByEmail.ID != id {
				return nil, ErrContactEmailExists
			}
			contactToUpdate.Email = email
			changed = true
		}
	}
	if data.Phone != nil {
		phone := strings.TrimSpace(*data.Phone)
		if phone == "" {
			return nil, ErrContactPhoneEmpty
		}
		if contactToUpdate.Phone != phone {
			// Проверка уникальности нового Phone
			existingByPhone, err := uc.contactRepo.GetByPhone(ctx, phone)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				uc.logger.ErrorContext(ctx, "Error checking new contact phone existence", slog.String("phone", phone), slog.Any("error", err))
				return nil, err
			}
			if existingByPhone != nil && existingByPhone.ID != id {
				return nil, ErrContactPhoneExists
			}
			contactToUpdate.Phone = phone
			changed = true
		}
	}

	if data.Transport != nil && contactToUpdate.Transport != *data.Transport {
		contactToUpdate.Transport = *data.Transport
		changed = true
	}
	if data.Printer != nil && contactToUpdate.Printer != *data.Printer {
		contactToUpdate.Printer = *data.Printer
		changed = true
	}
	if data.Allergies != nil && contactToUpdate.Allergies != *data.Allergies {
		contactToUpdate.Allergies = *data.Allergies
		changed = true
	}
	if data.VK != nil && contactToUpdate.VK != *data.VK {
		contactToUpdate.VK = *data.VK
		changed = true
	}
	if data.Telegram != nil && contactToUpdate.Telegram != *data.Telegram {
		contactToUpdate.Telegram = *data.Telegram
		changed = true
	}

	// Обновление групп
	if data.GroupIDs != nil {
		newGroups := make([]*domain.Group, 0, len(*data.GroupIDs))
		for _, groupID := range *data.GroupIDs {
			group, err := uc.groupRepo.GetByID(ctx, groupID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					uc.logger.WarnContext(ctx, "Group not found for contact association during update", slog.Uint64("groupID", uint64(groupID)))
					return nil, fmt.Errorf("%w: group with id %d not found", groupUseCase.ErrGroupNotFound, groupID)
				}
				uc.logger.ErrorContext(ctx, "Error fetching group for contact association update", slog.Uint64("groupID", uint64(groupID)), slog.Any("error", err))
				return nil, err
			}
			newGroups = append(newGroups, group)
		}
		contactToUpdate.Groups = newGroups
		changed = true // Даже если список групп тот же, но передан, считаем изменением для Replace
	}

	if !changed && data.GroupIDs == nil { // Если ничего не изменилось и группы не переданы для обновления
		uc.logger.InfoContext(ctx, "No changes detected for contact update", slog.Uint64("id", uint64(id)))
		return contactToUpdate, nil
	}

	if err := uc.contactRepo.Update(ctx, contactToUpdate); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to update contact via repository", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return nil, err
	}

	uc.logger.InfoContext(ctx, "Contact updated successfully", slog.Uint64("id", uint64(id)))
	// Возвращаем обновленный контакт со всеми ассоциациями
	return uc.contactRepo.GetByID(ctx, id)
}

func (uc *contactUseCase) DeleteContact(ctx context.Context, id uint) error {
	_, err := uc.contactRepo.GetByID(ctx, id) // Проверяем существование
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	if err := uc.contactRepo.Delete(ctx, id); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to delete contact via repository", slog.Uint64("id", uint64(id)), slog.Any("error", err))
		return err
	}
	uc.logger.InfoContext(ctx, "Contact deleted successfully", slog.Uint64("id", uint64(id)))
	return nil
}

func (uc *contactUseCase) AddContactToGroup(ctx context.Context, contactID uint, groupID uint) error {
	contact, err := uc.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return groupUseCase.ErrGroupNotFound
		}
		return err
	}

	// Проверим, не состоит ли контакт уже в этой группе (опционально, Append идемпотентен для связей)
	for _, existingGroup := range contact.Groups {
		if existingGroup.ID == group.ID {
			uc.logger.InfoContext(ctx, "Contact already in group", slog.Uint64("contactID", uint64(contactID)), slog.Uint64("groupID", uint64(groupID)))
			return nil // Уже в группе
		}
	}

	if err := uc.contactRepo.AddContactToGroup(ctx, contact, group); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to add contact to group via repository", slog.Uint64("contactID", uint64(contactID)), slog.Uint64("groupID", uint64(groupID)), slog.Any("error", err))
		return ErrGroupAssociation
	}
	uc.logger.InfoContext(ctx, "Contact added to group successfully", slog.Uint64("contactID", uint64(contactID)), slog.Uint64("groupID", uint64(groupID)))
	return nil
}

func (uc *contactUseCase) RemoveContactFromGroup(ctx context.Context, contactID uint, groupID uint) error {
	contact, err := uc.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return groupUseCase.ErrGroupNotFound
		}
		return err
	}

	// Проверим, состоит ли контакт в этой группе
	found := false
	for _, existingGroup := range contact.Groups {
		if existingGroup.ID == group.ID {
			found = true
			break
		}
	}
	if !found {
		uc.logger.WarnContext(ctx, "Contact not in group, cannot remove", slog.Uint64("contactID", uint64(contactID)), slog.Uint64("groupID", uint64(groupID)))
		return fmt.Errorf("contact is not a member of group %d", groupID) // Или можно просто nil вернуть, если не считать это ошибкой
	}

	if err := uc.contactRepo.RemoveContactFromGroup(ctx, contact, group); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to remove contact from group via repository", slog.Uint64("contactID", uint64(contactID)), slog.Uint64("groupID", uint64(groupID)), slog.Any("error", err))
		return ErrGroupAssociation
	}
	uc.logger.InfoContext(ctx, "Contact removed from group successfully", slog.Uint64("contactID", uint64(contactID)), slog.Uint64("groupID", uint64(groupID)))
	return nil
}
