<script lang="ts">
  import { authStore } from "$lib/store/authStore";
  import AuthService from "$lib/services/authService";
  import type { ContactPayload } from "$lib/types";

  // Предполагаем, что данные пользователя хранятся в authStore
  $: currentUser = $authStore.user;
  $: isAuthenticated = $authStore.isAuthenticated;
  $: debugMode = $authStore.debugMode;
  $: isAdmin = currentUser?.is_admin || false;

  let editMode = false;
  let saving = false;
  let errorMessage = '';
  let successMessage = '';

  // Форма данных для редактирования
  let formData = {
    name: '',
    email: '',
    phone: '',
    transport: '',
    printer: '',
    allergies: '',
    vk: '',
    telegram: '',
    telegram_id: ''
  };

  // Инициализация формы данными из контакта
  $: if (currentUser?.contact) {
    formData = {
      name: currentUser.contact.name || '',
      email: currentUser.contact.email || '',
      phone: currentUser.contact.phone || '',
      transport: currentUser.contact.transport || '',
      printer: currentUser.contact.printer || '',
      allergies: currentUser.contact.allergies || '',
      vk: currentUser.contact.vk || '',
      telegram: currentUser.contact.telegram || '',
      telegram_id: currentUser.contact.telegram_id?.toString() || ''
    };
  }

  async function saveContact() {
    if (!currentUser) return;
    
    saving = true;
    errorMessage = '';
    successMessage = '';

    try {
      const token = AuthService.getToken();
      if (!token) {
        throw new Error('Токен авторизации не найден');
      }

      const payload: Partial<ContactPayload> = {
        name: formData.name.trim(),
        email: formData.email.trim(),
        phone: formData.phone.trim(),
        transport: formData.transport || undefined,
        printer: formData.printer || undefined,
        allergies: formData.allergies.trim() || undefined,
        vk: formData.vk.trim() || undefined,
        telegram: formData.telegram.trim() || undefined,
        telegram_id: formData.telegram_id ? parseInt(formData.telegram_id) : undefined
      };

      await AuthService.updateMyContact(token, payload);
      
      // Обновляем данные пользователя
      await authStore.loadUser();
      
      editMode = false;
      successMessage = 'Контакт успешно обновлен';
      
      setTimeout(() => successMessage = '', 3000);
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : 'Произошла ошибка при сохранении';
      console.error('Error saving contact:', error);
    } finally {
      saving = false;
    }
  }

  function cancelEdit() {
    editMode = false;
    errorMessage = '';
    // Сбрасываем форму к текущим данным
    if (currentUser?.contact) {
      formData = {
        name: currentUser.contact.name || '',
        email: currentUser.contact.email || '',
        phone: currentUser.contact.phone || '',
        transport: currentUser.contact.transport || '',
        printer: currentUser.contact.printer || '',
        allergies: currentUser.contact.allergies || '',
        vk: currentUser.contact.vk || '',
        telegram: currentUser.contact.telegram || '',
        telegram_id: currentUser.contact.telegram_id?.toString() || ''
      };
    }
  }

  async function logout() {
    try {
      await authStore.logout();
    } catch (error) {
      console.error('Logout error:', error);
    }
  }
</script>

<div class="profile-page">
  <h2>Мой профиль</h2>
  {#if isAuthenticated && currentUser}
    <div class="profile-card">
      {#if currentUser.contact}
        <div class="profile-header">
          <h4>Информация о контакте</h4>
          {#if !editMode}
            <button on:click={() => editMode = true} class="edit-btn">
              Редактировать
            </button>
          {/if}
        </div>

        {#if successMessage}
          <div class="success-message">{successMessage}</div>
        {/if}

        {#if errorMessage}
          <div class="error-message">{errorMessage}</div>
        {/if}

        {#if editMode}
          <form on:submit|preventDefault={saveContact} class="edit-form">
            <div class="form-group">
              <label for="name">Имя:</label>
              <input 
                type="text" 
                id="name" 
                bind:value={formData.name} 
                required 
                disabled={saving}
              />
            </div>

            <div class="form-group">
              <label for="email">Email:</label>
              <input 
                type="email" 
                id="email" 
                bind:value={formData.email} 
                required 
                disabled={saving}
              />
            </div>

            <div class="form-group">
              <label for="phone">Телефон:</label>
              <input 
                type="tel" 
                id="phone" 
                bind:value={formData.phone} 
                required 
                disabled={saving}
              />
            </div>

            <div class="form-group">
              <label for="transport">Транспорт:</label>
              <select id="transport" bind:value={formData.transport} disabled={saving}>
                <option value="">Не указано</option>
                <option value="есть машина">Есть машина</option>
                <option value="есть права">Есть права</option>
                <option value="нет ничего">Нет ничего</option>
              </select>
            </div>

            <div class="form-group">
              <label for="printer">Принтер:</label>
              <select id="printer" bind:value={formData.printer} disabled={saving}>
                <option value="">Не указано</option>
                <option value="цветной">Цветной</option>
                <option value="обычный">Обычный</option>
                <option value="нет">Нет</option>
              </select>
            </div>

            <div class="form-group">
              <label for="allergies">Аллергии:</label>
              <textarea 
                id="allergies" 
                bind:value={formData.allergies} 
                disabled={saving}
                rows="3"
              ></textarea>
            </div>

            <div class="form-group">
              <label for="vk">VK:</label>
              <input 
                type="url" 
                id="vk" 
                bind:value={formData.vk} 
                disabled={saving}
                placeholder="https://vk.com/username"
              />
            </div>

            <div class="form-group">
              <label for="telegram">Telegram:</label>
              <input 
                type="text" 
                id="telegram" 
                bind:value={formData.telegram} 
                disabled={saving}
                placeholder="@username"
              />
            </div>

            <div class="form-group">
              <label for="telegram_id">Telegram ID:</label>
              <input 
                type="number" 
                id="telegram_id" 
                bind:value={formData.telegram_id} 
                disabled={saving}
                placeholder="123456789"
              />
            </div>

            <div class="form-actions">
              <button type="submit" disabled={saving} class="save-btn">
                {saving ? 'Сохранение...' : 'Сохранить'}
              </button>
              <button type="button" on:click={cancelEdit} disabled={saving} class="cancel-btn">
                Отмена
              </button>
            </div>
          </form>
        {:else}
          <div class="profile-info">
            <p><strong>Имя:</strong> {currentUser.contact.name}</p>
            <p><strong>Email:</strong> {currentUser.contact.email}</p>
            <p><strong>Телефон:</strong> {currentUser.contact.phone}</p>
            {#if currentUser.contact.transport}
              <p><strong>Транспорт:</strong> {currentUser.contact.transport}</p>
            {/if}
            {#if currentUser.contact.printer}
              <p><strong>Принтер:</strong> {currentUser.contact.printer}</p>
            {/if}
            {#if currentUser.contact.allergies}
              <p><strong>Аллергии:</strong> {currentUser.contact.allergies}</p>
            {/if}
            {#if currentUser.contact.vk}
              <p><strong>VK:</strong> <a href={currentUser.contact.vk} target="_blank" rel="noopener noreferrer">{currentUser.contact.vk}</a></p>
            {/if}
            {#if currentUser.contact.telegram}
              <p><strong>Telegram:</strong> {currentUser.contact.telegram}</p>
            {/if}
            {#if currentUser.contact.telegram_id}
              <p><strong>Telegram ID:</strong> {currentUser.contact.telegram_id}</p>
            {/if}
          </div>
        {/if}
      {:else}
        <p style="color: #666; font-style: italic;">Контакт не привязан к аккаунту</p>
      {/if}
      
      {#if isAdmin}
        <div class="debug-section">
          <label class="debug-toggle">
            <input 
              type="checkbox" 
              bind:checked={debugMode}
              on:change={() => authStore.toggleDebugMode()}
            />
            Отладочный режим (разрешить редактирование всем)
          </label>
        </div>
      {/if}
      
      <button on:click={logout} class="logout-btn">
        Выйти из системы
      </button>
    </div>
  {:else}
    <div class="auth-required">
      <p>Для просмотра профиля необходимо войти в систему.</p>
      <a href="/login" style="color: #1890ff; text-decoration: none; font-weight: 500;">Войти через Telegram</a>
    </div>
  {/if}
</div>

<style>
  .profile-page {
    padding: 20px;
  }

  .profile-card {
    background-color: #fff;
    padding: 30px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    max-width: 600px;
    margin: 20px auto;
  }

  .profile-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .profile-header h4 {
    margin: 0;
    color: #1890ff;
  }

  .edit-btn {
    background-color: #1890ff;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
  }

  .edit-btn:hover {
    background-color: #40a9ff;
  }

  .profile-info p {
    margin: 8px 0;
    font-size: 1rem;
    line-height: 1.6;
  }

  .profile-info p strong {
    color: #555;
    min-width: 100px;
    display: inline-block;
  }

  .edit-form {
    margin-top: 20px;
  }

  .form-group {
    margin-bottom: 16px;
  }

  .form-group label {
    display: block;
    margin-bottom: 4px;
    font-weight: 500;
    color: #333;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #d9d9d9;
    border-radius: 4px;
    font-size: 14px;
    box-sizing: border-box;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #1890ff;
    box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
  }

  .form-actions {
    display: flex;
    gap: 12px;
    margin-top: 24px;
  }

  .save-btn {
    background-color: #52c41a;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
  }

  .save-btn:hover:not(:disabled) {
    background-color: #73d13d;
  }

  .save-btn:disabled {
    background-color: #d9d9d9;
    cursor: not-allowed;
  }

  .cancel-btn {
    background-color: #d9d9d9;
    color: #333;
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
  }

  .cancel-btn:hover:not(:disabled) {
    background-color: #bfbfbf;
  }

  .logout-btn {
    margin-top: 20px;
    background-color: #ff4d4f;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
  }

  .logout-btn:hover {
    background-color: #ff7875;
  }

  .success-message {
    background-color: #f6ffed;
    border: 1px solid #b7eb8f;
    color: #52c41a;
    padding: 12px;
    border-radius: 4px;
    margin-bottom: 16px;
  }

  .error-message {
    background-color: #fff2f0;
    border: 1px solid #ffccc7;
    color: #ff4d4f;
    padding: 12px;
    border-radius: 4px;
    margin-bottom: 16px;
  }

  .debug-section {
    margin: 20px 0;
    padding: 16px;
    background-color: #fff7e6;
    border: 1px solid #ffd591;
    border-radius: 4px;
  }

  .debug-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: #d46b08;
    cursor: pointer;
  }

  .debug-toggle input[type="checkbox"] {
    width: auto;
    margin: 0;
  }

  .auth-required {
    text-align: center;
    padding: 40px 20px;
  }
</style> 