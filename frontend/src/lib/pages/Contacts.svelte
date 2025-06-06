<script lang="ts">
  import { onMount } from 'svelte';
  import ContactService from '$lib/services/contactService';
  import GroupService from '$lib/services/groupService'; // Импортируем GroupService
  import { authStore } from '$lib/store/authStore';
  import type { Contact, ContactPayload, Group, ContactBasic } from '$lib/types';

  let contacts: (Contact | ContactBasic)[] = [];
  let allGroups: Group[] = []; // Для хранения списка всех групп
  let isLoading = true;
  let error: string | null = null;
  let generalError: string | null = null; // Общая ошибка для страницы
  let isAuthenticated = false;

  let searchTerm: string = '';
  let sortBy: keyof Pick<Contact, 'name' | 'email' | 'phone'> = 'name';

  // Состояние для модального окна создания/редактирования контакта
  let showContactModal = false;
  let currentContactForm: Partial<ContactPayload> = {};
  let editingContactId: string | null = null;

  // Состояние для модального окна управления группами контакта
  let showManageGroupsModal = false;
  let contactToManageGroups: Contact | null = null;
  let selectedGroupIds: Set<string> = new Set(); // ID групп, в которых состоит контакт
  let isSavingGroups = false;
  let groupsModalError: string | null = null;

  // Подписываемся на изменения состояния авторизации
  $: isAuthenticated = $authStore.isAuthenticated;

  // Функция для проверки, является ли контакт полным
  function isFullContact(contact: Contact | ContactBasic): contact is Contact {
    return 'email' in contact;
  }

  onMount(async () => {
    await loadInitialData();
  });

  async function loadInitialData() {
    isLoading = true;
    generalError = null;
    try {
      const [fetchedContacts, fetchedGroups] = await Promise.all([
        ContactService.getAllContacts(),
        GroupService.getAllGroups()
      ]);
      contacts = fetchedContacts || [];
      allGroups = fetchedGroups || [];
    } catch (err: any) {
      console.error("Failed to load initial data:", err);
      generalError = err.message || "Не удалось загрузить данные.";
      contacts = [];
      allGroups = [];
    }
    isLoading = false;
  }

  async function loadContacts() {
    // Эта функция может быть объединена с loadInitialData или вызываться отдельно при необходимости
    isLoading = true;
    generalError = null;
    try {
      contacts = (await ContactService.getAllContacts()) || [];
    } catch (err: any) {
      console.error("Failed to load contacts:", err);
      generalError = err.message || "Не удалось загрузить контакты.";
      contacts = [];
    }
    isLoading = false;
  }

  $: filteredContacts = contacts
    .filter(contact => {
      // Для базовых контактов ищем только по имени
      if (!isFullContact(contact)) {
        return contact.name.toLowerCase().includes(searchTerm.toLowerCase());
      }
      // Для полных контактов ищем по всем полям
      return Object.values(contact).some(value => {
        if (typeof value === 'string') {
          return value.toLowerCase().includes(searchTerm.toLowerCase());
        }
        if (Array.isArray(value)) {
          return value.some(group => group.name.toLowerCase().includes(searchTerm.toLowerCase()));
        }
        return false;
      });
    })
    .sort((a, b) => {
      // Сортировка только по имени для базовых контактов
      if (sortBy === 'name' || !isFullContact(a) || !isFullContact(b)) {
        const valA = a.name;
        const valB = b.name;
        if (valA < valB) return -1;
        if (valA > valB) return 1;
        return 0;
      }
      // Полная сортировка для авторизованных пользователей
      const valA = a[sortBy];
      const valB = b[sortBy];
      if (valA < valB) return -1;
      if (valA > valB) return 1;
      return 0;
    });

  function openCreateContactModal() {
    editingContactId = null;
    currentContactForm = { name: '', email: '', phone: '' };
    error = null; // Сброс ошибки формы
    showContactModal = true;
  }

  function openEditContactModal(contact: Contact) {
    editingContactId = contact.id;
    currentContactForm = { ...contact }; // Копируем данные для редактирования
    error = null; // Сброс ошибки формы
    showContactModal = true;
  }

  async function handleDeleteContact(contactId: string) {
    if (!confirm('Вы уверены, что хотите удалить этот контакт?')) return;
    // isLoading = true; // Можно использовать отдельный флаг загрузки для удаления
    try {
      await ContactService.deleteContact(contactId);
      await loadContacts(); 
    } catch (err: any) {
      console.error("Failed to delete contact:", err);
      generalError = err.message || "Не удалось удалить контакт.";
    }
    // isLoading = false;
  }
  
  async function handleContactFormSubmit() {
    if (!currentContactForm.name || !currentContactForm.email || !currentContactForm.phone) {
        alert("Имя, Email и Телефон обязательны для заполнения.");
        return;
    }
    // isLoading = true;
    error = null;
    try {
      if (editingContactId) {
        await ContactService.updateContact(editingContactId, currentContactForm as ContactPayload);
      } else {
        await ContactService.createContact(currentContactForm as ContactPayload);
      }
      showContactModal = false;
      await loadContacts(); 
    } catch (err: any) {
      console.error("Failed to save contact:", err);
      error = err.message || "Не удалось сохранить контакт."; // Ошибка для модального окна контакта
    }
    // isLoading = false;
  }

  // --- Логика для управления группами контакта ---
  function openManageGroupsModal(contact: Contact) {
    contactToManageGroups = contact;
    selectedGroupIds = new Set(contact.groups?.map(g => g.id) || []);
    groupsModalError = null;
    showManageGroupsModal = true;
  }

  async function handleSaveContactGroups() {
    if (!contactToManageGroups) return;

    isSavingGroups = true;
    groupsModalError = null;
    const contactId = contactToManageGroups.id;
    const initialGroupIds = new Set(contactToManageGroups.groups?.map(g => g.id) || []);
    
    try {
      // Удалить из групп, из которых убрали
      for (const groupId of initialGroupIds) {
        if (!selectedGroupIds.has(groupId)) {
          await ContactService.removeContactFromGroup(contactId, groupId);
        }
      }
      // Добавить в новые группы
      for (const groupId of selectedGroupIds) {
        if (!initialGroupIds.has(groupId)) {
          await ContactService.addContactToGroup(contactId, groupId);
        }
      }
      showManageGroupsModal = false;
      await loadContacts(); // Перезагрузить контакты для отображения изменений
    } catch (err: any) {
      console.error("Failed to update contact groups:", err);
      groupsModalError = err.message || "Не удалось обновить группы контакта.";
    }
    isSavingGroups = false;
  }

</script>

<div class="contacts-page">
  <h2>Контакты</h2>
  
  {#if generalError}
    <p class="error-message global-error">Ошибка: {generalError} <button on:click={loadInitialData}>Попробовать снова</button></p>
  {/if}

  <div class="controls">
    <input type="text" placeholder="Поиск..." bind:value={searchTerm} disabled={isLoading && contacts.length === 0}>
    {#if isAuthenticated}
      <select bind:value={sortBy} disabled={isLoading && contacts.length === 0}>
        <option value="name">Имя</option>
        <option value="email">Email</option>
        <option value="phone">Телефон</option>
      </select>
      <button on:click={openCreateContactModal} disabled={isLoading && contacts.length === 0}>Добавить контакт</button>
    {:else}
      <p class="auth-notice">Для управления контактами необходимо <a href="/login">войти в систему</a></p>
    {/if}
  </div>

  {#if isLoading && contacts.length === 0}
    <p>Загрузка данных...</p>
  {:else if !generalError && filteredContacts.length === 0 && searchTerm}
     <p>Контакты по вашему запросу не найдены.</p>
  {:else if !generalError && contacts.length === 0}
    <p>Список контактов пуст. <button on:click={openCreateContactModal}>Добавить первый контакт?</button></p>
  {:else if filteredContacts.length > 0}
    <ul class="contact-list">
      {#each filteredContacts as contact (contact.id)}
        <li class="contact-item">
          <h3>{contact.name}</h3>
          
          {#if isFullContact(contact)}
            <!-- Полная информация для авторизованных пользователей -->
            <p><strong>Email:</strong> {contact.email}</p>
            <p><strong>Телефон:</strong> {contact.phone}</p>
            {#if contact.transport}<p><strong>Транспорт:</strong> {contact.transport}</p>{/if}
            {#if contact.printer}<p><strong>Принтер:</strong> {contact.printer}</p>{/if}
            {#if contact.allergies}<p><strong>Аллергии:</strong> {contact.allergies}</p>{/if}
            {#if contact.vk}<p><strong>VK:</strong> <a href={contact.vk} target="_blank" rel="noopener noreferrer">{contact.vk}</a></p>{/if}
            {#if contact.telegram}<p><strong>Telegram:</strong> {contact.telegram}</p>{/if}
            {#if contact.groups && contact.groups.length > 0}
              <p><strong>Группы:</strong> {contact.groups.map(g => g.name).join(', ')}</p>
            {:else}
              <p><strong>Группы:</strong> <em>не состоит в группах</em></p>
            {/if}
            
            <div class="actions">
              <button class="edit" on:click={() => openEditContactModal(contact)}>Редактировать</button>
              <button class="manage-groups" on:click={() => openManageGroupsModal(contact)}>Упр. группами</button>
              <button class="delete" on:click={() => handleDeleteContact(contact.id)}>Удалить</button>
            </div>
          {:else}
            <!-- Ограниченная информация для неавторизованных пользователей -->
            <p class="limited-info">Для просмотра полной информации необходимо <a href="/login">войти в систему</a></p>
          {/if}
        </li>
      {/each}
    </ul>
  {:else}
    <p>Список контактов пуст или произошла ошибка при фильтрации.</p> <!-- Резервное сообщение -->
  {/if}
</div>

<!-- Модальное окно для создания/редактирования контакта -->
{#if showContactModal}
  <div class="modal-backdrop" on:click={() => showContactModal = false}></div>
  <div class="modal contact-modal">
    <h3>{editingContactId ? 'Редактировать контакт' : 'Создать контакт'}</h3>
    <form on:submit|preventDefault={handleContactFormSubmit}>
      <div class="form-group">
        <label for="contactName">Имя*:</label>
        <input type="text" id="contactName" bind:value={currentContactForm.name} required disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactEmail">Email*:</label>
        <input type="email" id="contactEmail" bind:value={currentContactForm.email} required disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactPhone">Телефон*:</label>
        <input type="tel" id="contactPhone" bind:value={currentContactForm.phone} required disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactTransport">Транспорт:</label>
        <select id="contactTransport" bind:value={currentContactForm.transport} disabled={isLoading}>
          <option value={undefined}>Не указано</option>
          <option value="есть машина">Есть машина</option>
          <option value="есть права">Есть права</option>
          <option value="нет ничего">Нет ничего</option>
        </select>
      </div>
      <div class="form-group">
        <label for="contactPrinter">Принтер:</label>
        <select id="contactPrinter" bind:value={currentContactForm.printer} disabled={isLoading}>
          <option value={undefined}>Не указано</option>
          <option value="цветной">Цветной</option>
          <option value="обычный">Обычный</option>
          <option value="нет">Нет</option>
        </select>
      </div>
      <div class="form-group">
        <label for="contactAllergies">Аллергии:</label>
        <textarea id="contactAllergies" bind:value={currentContactForm.allergies} disabled={isLoading}></textarea>
      </div>
      <div class="form-group">
        <label for="contactVK">VK (ссылка):</label>
        <input type="url" id="contactVK" bind:value={currentContactForm.vk} disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="contactTelegram">Telegram (username):</label>
        <input type="text" id="contactTelegram" bind:value={currentContactForm.telegram} disabled={isLoading}>
      </div>
      {#if error} <!-- Ошибка формы контакта -->
        <p class="error-message">{error}</p>
      {/if}
      <div class="modal-actions">
        <button type="submit" class="primary" disabled={isLoading}>{#if isLoading}Сохранение...{:else}Сохранить{/if}</button>
        <button type="button" on:click={() => showContactModal = false} disabled={isLoading}>Отмена</button>
      </div>
    </form>
  </div>
{/if}

<!-- Модальное окно для управления группами контакта -->
{#if showManageGroupsModal && contactToManageGroups}
  <div class="modal-backdrop" on:click={() => showManageGroupsModal = false}></div>
  <div class="modal groups-modal">
    <h3>Управление группами для: {contactToManageGroups.name}</h3>
    {#if allGroups.length === 0}
      <p>Список групп пуст. Сначала <a href="/groups">создайте группы</a>.</p>
    {:else}
      <form on:submit|preventDefault={handleSaveContactGroups}>
        <div class="groups-checkbox-list">
          {#each allGroups as group (group.id)}
            <label>
              <input 
                type="checkbox" 
                value={group.id} 
                checked={selectedGroupIds.has(group.id)}
                on:change={() => {
                  if (selectedGroupIds.has(group.id)) {
                    selectedGroupIds.delete(group.id);
                  } else {
                    selectedGroupIds.add(group.id);
                  }
                  selectedGroupIds = selectedGroupIds; // Для реактивности Svelte
                }}
                disabled={isSavingGroups}
              />
              {group.name}
            </label>
          {/each}
        </div>
        {#if groupsModalError}
          <p class="error-message">{groupsModalError}</p>
        {/if}
        <div class="modal-actions">
          <button type="submit" class="primary" disabled={isSavingGroups || allGroups.length === 0}>{#if isSavingGroups}Сохранение...{:else}Сохранить группы{/if}</button>
          <button type="button" on:click={() => showManageGroupsModal = false} disabled={isSavingGroups}>Отмена</button>
        </div>
      </form>
    {/if}
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
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); /* Немного увеличил ширину */
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
  .limited-info {
    color: #666;
    font-style: italic;
    text-align: center;
    margin: 20px 0;
  }
  
  .auth-notice {
    color: #666;
    font-size: 0.9rem;
    margin: 0;
  }
  
  .auth-notice a {
    color: #1890ff;
    text-decoration: none;
  }
  
  .auth-notice a:hover {
    text-decoration: underline;
  }
  
  .actions {
    margin-top: auto; 
    padding-top: 15px;
    border-top: 1px solid #f0f0f0;
    display: flex;
    gap: 8px; /* Уменьшил gap */
    justify-content: flex-end;
    flex-wrap: wrap; /* Для переноса кнопок на маленьких экранах */
  }
  .actions button {
    padding: 6px 10px; /* Немного уменьшил padding */
    font-size: 0.8rem; /* Уменьшил шрифт */
  }
  .actions button.edit {
    background-color: #1890ff;
    color: white;
  }
  .actions button.edit:hover {
    background-color: #40a9ff;
  }
   .actions button.manage-groups {
    background-color: #52c41a; /* Зеленый */
    color: white;
  }
  .actions button.manage-groups:hover {
    background-color: #73d13d;
  }
  .actions button.delete {
    background-color: #ff4d4f;
    color: white;
  }
  .actions button.delete:hover {
    background-color: #ff7875;
  }

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
    max-width: 500px; /* Общий для всех модалок */
    max-height: 90vh;
    overflow-y: auto;
  }
  .modal.groups-modal { /* Специфичный стиль для модалки групп */
     max-width: 400px;
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
  .global-error {
    border: 1px solid red;
    padding: 10px;
    margin-bottom: 15px;
    background-color: #ffeeee;
  }
  .groups-checkbox-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-height: 250px; /* Ограничение высоты списка */
    overflow-y: auto; /* Прокрутка, если групп много */
    margin-bottom: 20px;
    border: 1px solid #eee;
    padding: 10px;
    border-radius: 4px;
  }
  .groups-checkbox-list label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }
</style> 