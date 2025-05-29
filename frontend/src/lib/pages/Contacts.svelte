<script lang="ts">
  import { onMount } from 'svelte';
  import ContactService from '$lib/services/contactService';
  import type { Contact, ContactPayload, Group } from '$lib/types';

  let contacts: Contact[] = [];
  let isLoading = true;
  let error: string | null = null;

  let searchTerm: string = '';
  let sortBy: keyof Pick<Contact, 'name' | 'email' | 'phone'> = 'name'; // Типизируем sortBy

  // Состояние для модального окна/формы создания/редактирования
  let showModal = false;
  let currentContact: Partial<ContactPayload> = {}; // Для формы создания/редактирования
  let editingContactId: string | null = null; // ID редактируемого контакта

  onMount(async () => {
    await loadContacts();
  });

  async function loadContacts() {
    isLoading = true;
    error = null;
    try {
      const fetchedContacts = await ContactService.getAllContacts();
      contacts = fetchedContacts || []; // Если API вернет null/undefined
    } catch (err: any) {
      console.error("Failed to load contacts:", err);
      error = err.message || "Не удалось загрузить контакты.";
      contacts = []; // Очищаем список в случае ошибки
    }
    isLoading = false;
  }

  $: filteredContacts = contacts
    .filter(contact => 
      Object.values(contact).some(value => {
        if (typeof value === 'string') {
          return value.toLowerCase().includes(searchTerm.toLowerCase());
        }
        if (Array.isArray(value)) { // Для поиска по группам, если они есть в контакте
          return value.some(group => group.name.toLowerCase().includes(searchTerm.toLowerCase()));
        }
        return false;
      })
    )
    .sort((a, b) => {
      const valA = a[sortBy];
      const valB = b[sortBy];
      if (valA < valB) return -1;
      if (valA > valB) return 1;
      return 0;
    });

  function openCreateModal() {
    editingContactId = null;
    currentContact = { name: '', email: '', phone: '' }; // Сброс формы
    showModal = true;
  }

  function openEditModal(contact: Contact) {
    editingContactId = contact.id;
    currentContact = { ...contact }; // Копируем данные для редактирования
    showModal = true;
  }

  async function handleDelete(contactId: string) {
    if (!confirm('Вы уверены, что хотите удалить этот контакт?')) return;
    isLoading = true; // Можно использовать отдельный флаг загрузки для удаления
    try {
      await ContactService.deleteContact(contactId);
      await loadContacts(); // Перезагружаем список
    } catch (err: any) {
      console.error("Failed to delete contact:", err);
      error = err.message || "Не удалось удалить контакт.";
      // Можно показать ошибку пользователю
    }
    isLoading = false;
  }
  
  async function handleFormSubmit() {
    if (!currentContact.name || !currentContact.email || !currentContact.phone) {
        alert("Имя, Email и Телефон обязательны для заполнения.");
        return;
    }
    isLoading = true;
    error = null;
    try {
      if (editingContactId) {
        await ContactService.updateContact(editingContactId, currentContact as ContactPayload);
      } else {
        await ContactService.createContact(currentContact as ContactPayload);
      }
      showModal = false;
      await loadContacts(); // Перезагружаем список
    } catch (err: any) {
      console.error("Failed to save contact:", err);
      error = err.message || "Не удалось сохранить контакт.";
      // Ошибка останется в модальном окне или можно ее показать глобально
    }
    isLoading = false;
  }

  // TODO: Реализовать функции для addContactToGroup / removeContactFromGroup
  // Это потребует UI для выбора группы, возможно, отдельное модальное окно.

</script>

<div class="contacts-page">
  <h2>Контакты</h2>
  
  <div class="controls">
    <input type="text" placeholder="Поиск..." bind:value={searchTerm} disabled={isLoading}>
    <select bind:value={sortBy} disabled={isLoading}>
      <option value="name">Имя</option>
      <option value="email">Email</option>
      <option value="phone">Телефон</option>
    </select>
    <button on:click={openCreateModal} disabled={isLoading}>Добавить контакт</button>
  </div>

  {#if isLoading && contacts.length === 0}
    <p>Загрузка контактов...</p>
  {:else if error}
    <p class="error-message">Ошибка: {error}</p>
    <button on:click={loadContacts}>Попробовать снова</button>
  {:else if filteredContacts.length > 0}
    <ul class="contact-list">
      {#each filteredContacts as contact (contact.id)}
        <li class="contact-item">
          <h3>{contact.name}</h3>
          <p><strong>Email:</strong> {contact.email}</p>
          <p><strong>Телефон:</strong> {contact.phone}</p>
          {#if contact.transport}<p><strong>Транспорт:</strong> {contact.transport}</p>{/if}
          {#if contact.printer}<p><strong>Принтер:</strong> {contact.printer}</p>{/if}
          {#if contact.allergies}<p><strong>Аллергии:</strong> {contact.allergies}</p>{/if}
          {#if contact.vk}<p><strong>VK:</strong> <a href={contact.vk} target="_blank">{contact.vk}</a></p>{/if}
          {#if contact.telegram}<p><strong>Telegram:</strong> {contact.telegram}</p>{/if}
          {#if contact.groups && contact.groups.length > 0}
            <p><strong>Группы:</strong> {contact.groups.map(g => g.name).join(', ')}</p>
          {/if}
          <div class="actions">
            <button class="edit" on:click={() => openEditModal(contact)}>Редактировать</button>
            <button class="delete" on:click={() => handleDelete(contact.id)}>Удалить</button>
            <!-- TODO: Кнопки для добавления/удаления из групп -->
          </div>
        </li>
      {/each}
    </ul>
  {:else}
    <p>Контакты не найдены. <button on:click={openCreateModal}>Добавить первый контакт?</button></p>
  {/if}
</div>

{#if showModal}
  <div class="modal-backdrop" on:click={() => showModal = false}></div>
  <div class="modal">
    <h3>{editingContactId ? 'Редактировать контакт' : 'Создать контакт'}</h3>
    <form on:submit|preventDefault={handleFormSubmit}>
      <div class="form-group">
        <label for="contactName">Имя*:</label>
        <input type="text" id="contactName" bind:value={currentContact.name} required disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactEmail">Email*:</label>
        <input type="email" id="contactEmail" bind:value={currentContact.email} required disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactPhone">Телефон*:</label>
        <input type="tel" id="contactPhone" bind:value={currentContact.phone} required disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactTransport">Транспорт:</label>
        <select id="contactTransport" bind:value={currentContact.transport} disabled={isLoading}>
          <option value={undefined}>Не указано</option>
          <option value="есть машина">Есть машина</option>
          <option value="есть права">Есть права</option>
          <option value="нет ничего">Нет ничего</option>
        </select>
      </div>
      <div class="form-group">
        <label for="contactPrinter">Принтер:</label>
        <select id="contactPrinter" bind:value={currentContact.printer} disabled={isLoading}>
          <option value={undefined}>Не указано</option>
          <option value="цветной">Цветной</option>
          <option value="обычный">Обычный</option>
          <option value="нет">Нет</option>
        </select>
      </div>
      <div class="form-group">
        <label for="contactAllergies">Аллергии:</label>
        <textarea id="contactAllergies" bind:value={currentContact.allergies} disabled={isLoading}></textarea>
      </div>
      <div class="form-group">
        <label for="contactVK">VK (ссылка):</label>
        <input type="url" id="contactVK" bind:value={currentContact.vk} disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactTelegram">Telegram (username):</label>
        <input type="text" id="contactTelegram" bind:value={currentContact.telegram} disabled={isLoading}>
      </div>
      {#if error && showModal} <!-- Показываем ошибку внутри модального окна -->
        <p class="error-message">{error}</p>
      {/if}
      <div class="modal-actions">
        <button type="submit" class="primary" disabled={isLoading}>{#if isLoading}Сохранение...{:else}Сохранить{/if}</button>
        <button type="button" on:click={() => showModal = false} disabled={isLoading}>Отмена</button>
      </div>
    </form>
  </div>
{/if}

<style>
  .contacts-page {
    padding: 20px;
  }
  h2 {
    margin-bottom: 20px;
  }
  .controls {
    display: flex;
    gap: 10px;
    margin-bottom: 20px;
    align-items: center;
  }
  .controls input[type="text"] {
    flex-grow: 1;
    margin-bottom: 0; /* Убираем стандартный отступ */
  }
   .controls select {
    margin-bottom: 0;
    min-width: 150px;
  }
  .contact-list {
    list-style: none;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 20px;
  }
  .contact-item {
    background-color: #fff;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    display: flex;
    flex-direction: column;
  }
  .contact-item h3 {
    margin-top: 0;
    margin-bottom: 10px;
    color: #1890ff;
  }
  .contact-item p {
    margin: 5px 0;
    font-size: 0.9rem;
    word-break: break-word;
  }
  .actions {
    margin-top: auto; /* Прижимает кнопки вниз */
    padding-top: 10px;
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }
  .actions button {
    padding: 6px 12px;
    font-size: 0.85rem;
  }
  .actions button.edit {
    background-color: #1890ff;
    color: white;
  }
  .actions button.edit:hover {
    background-color: #40a9ff;
  }
  .actions button.delete {
    background-color: #ff4d4f;
    color: white;
  }
  .actions button.delete:hover {
    background-color: #ff7875;
  }

  /* Стили для модального окна */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0,0,0,0.5);
    z-index: 10;
  }
  .modal {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background-color: white;
    padding: 25px;
    border-radius: 8px;
    box-shadow: 0 5px 15px rgba(0,0,0,0.3);
    z-index: 11;
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
  }
  .modal h3 {
    margin-top: 0;
    margin-bottom: 20px;
  }
  .modal .form-group {
    margin-bottom: 15px;
  }
  .modal-actions {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
  .modal-actions button.primary {
    background-color: #1890ff;
    color: white;
  }
  .modal-actions button.primary:hover {
    background-color: #40a9ff;
  }
  .modal-actions button.primary:disabled,
  .modal-actions button:disabled {
    background-color: #ccc;
    cursor: not-allowed;
  }
  .error-message {
    color: red;
    margin-bottom: 10px;
  }
</style> 