<script lang="ts">
  import { authStore } from "$lib/store/authStore";

  // Предполагаем, что данные пользователя хранятся в authStore
  $: currentUser = $authStore.user;
  $: isAuthenticated = $authStore.isAuthenticated;

  // TODO: Добавить форму для редактирования профиля
  // TODO: Обработка сохранения изменений
</script>

<div class="profile-page">
  <h2>Мой профиль</h2>
  {#if isAuthenticated && currentUser}
    <div class="profile-card">
      <div class="profile-info">
        <h3>Профиль пользователя</h3>
        <p><strong>Telegram ID:</strong> {currentUser.telegram_id}</p>
        <p><strong>Статус:</strong> {currentUser.is_active ? 'Активен' : 'Неактивен'}</p>
        <p><strong>Дата регистрации:</strong> {new Date(currentUser.created_at).toLocaleDateString('ru-RU')}</p>
        
        {#if currentUser.contact}
          <hr style="margin: 20px 0; border: none; border-top: 1px solid #eee;">
          <h4>Информация о контакте</h4>
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
        {:else}
          <hr style="margin: 20px 0; border: none; border-top: 1px solid #eee;">
          <p style="color: #666; font-style: italic;">Контакт не привязан к аккаунту</p>
        {/if}
        
        <button on:click={() => authStore.logout()} style="margin-top: 20px; background-color: #ff4d4f; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer;">
          Выйти из системы
        </button>
      </div>
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
  .profile-info h3 {
    margin-top: 0;
    color: #1890ff;
    margin-bottom: 15px;
  }
  .profile-info p {
    margin: 8px 0;
    font-size: 1rem;
    line-height: 1.6;
  }
  .profile-info p strong {
    color: #555;
    min-width: 100px; /* Для выравнивания */
    display: inline-block;
  }
  /* Стили для кнопки редактирования (если будет добавлена) */
</style> 