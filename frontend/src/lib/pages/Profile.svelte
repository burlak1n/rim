<script lang="ts">
  import AuthStore from "$lib/store/authStore";

  // Предполагаем, что данные пользователя хранятся в AuthStore
  let currentUser: { id: string; name: string; email: string; vk?: string; telegram?: string; transport?: string; printer?: string; allergies?: string; } | null = null;

  AuthStore.subscribe(auth => {
    if (auth.isAuthenticated && auth.user) {
      // Загружаем полные данные пользователя, если это необходимо, или используем то, что есть в store
      currentUser = {
        ...auth.user,
        // Здесь можно добавить заглушки для необязательных полей из ТЗ
        vk: auth.user.vk || 'Не указан',
        telegram: auth.user.telegram || 'Не указан',
        transport: auth.user.transport || 'Не указано',
        printer: auth.user.printer || 'Не указан',
        allergies: auth.user.allergies || 'Нет данных'
      };
    } else {
      currentUser = null;
    }
  });

  // TODO: Добавить форму для редактирования профиля
  // TODO: Обработка сохранения изменений
</script>

<div class="profile-page">
  <h2>Мой профиль</h2>
  {#if currentUser}
    <div class="profile-card">
      <div class="profile-info">
        <h3>{currentUser.name}</h3>
        <p><strong>Email:</strong> {currentUser.email}</p>
        <p><strong>VK:</strong> {currentUser.vk}</p>
        <p><strong>Telegram:</strong> {currentUser.telegram}</p>
        <p><strong>Транспорт:</strong> {currentUser.transport}</p>
        <p><strong>Принтер:</strong> {currentUser.printer}</p>
        <p><strong>Аллергии:</strong> {currentUser.allergies}</p>
      </div>
      <!-- <button on:click={() => console.log('Редактировать профиль')}>Редактировать</button> -->
       <p style="margin-top: 20px; color: #888;"><i>(Редактирование профиля будет доступно позже)</i></p>
    </div>
  {:else}
    <p>Информация о пользователе не загружена. Пожалуйста, войдите в систему.</p>
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