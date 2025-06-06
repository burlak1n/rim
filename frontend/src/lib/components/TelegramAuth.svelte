<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../store/authStore.js';
  import { config } from '../config.js';
  import type { TelegramAuthData } from '../types.js';

  export let botUsername = config.botUsername;
  export let redirectUrl = window.location.origin;

  let telegramWidgetContainer: HTMLDivElement;
  let isLoading = false;
  let error = '';

  onMount(() => {
    // Загружаем скрипт Telegram Widget
    const script = document.createElement('script');
    script.async = true;
    script.src = 'https://telegram.org/js/telegram-widget.js?22';
    script.setAttribute('data-telegram-login', botUsername);
    script.setAttribute('data-size', 'large');
    script.setAttribute('data-auth-url', redirectUrl);
    script.setAttribute('data-request-access', 'write');
    
    // Создаем глобальную функцию для обработки ответа от Telegram
    (window as any).onTelegramAuth = handleTelegramAuth;
    script.setAttribute('data-onauth', 'onTelegramAuth(user)');

    telegramWidgetContainer.appendChild(script);

    return () => {
      // Очистка при размонтировании компонента
      delete (window as any).onTelegramAuth;
    };
  });

  async function handleTelegramAuth(user: any) {
    try {
      isLoading = true;
      error = '';

      const authData: TelegramAuthData = {
        id: user.id,
        first_name: user.first_name,
        last_name: user.last_name,
        username: user.username,
        photo_url: user.photo_url,
        auth_date: user.auth_date,
        hash: user.hash,
      };

      await authStore.loginWithTelegram(authData);
      
      // Перенаправляем на главную страницу после успешной авторизации
      window.location.href = '/';
    } catch (err) {
      console.error('Telegram auth error:', err);
      error = err instanceof Error ? err.message : 'Ошибка авторизации';
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="telegram-auth">
  <h3>Войти через Telegram</h3>
  
  {#if error}
    <div class="error">
      {error}
    </div>
  {/if}

  {#if isLoading}
    <div class="loading">
      Авторизация...
    </div>
  {/if}

  <div bind:this={telegramWidgetContainer} class="telegram-widget"></div>
  
  <p class="info">
    Для входа в систему используйте Telegram. Нажмите кнопку выше для авторизации.
  </p>
</div>

<style>
  .telegram-auth {
    max-width: 400px;
    margin: 0 auto;
    padding: 2rem;
    text-align: center;
  }

  .telegram-widget {
    margin: 1rem 0;
    display: flex;
    justify-content: center;
  }

  .error {
    background-color: #fee;
    color: #c33;
    padding: 0.75rem;
    border-radius: 4px;
    margin: 1rem 0;
    border: 1px solid #fcc;
  }

  .loading {
    background-color: #e6f3ff;
    color: #0066cc;
    padding: 0.75rem;
    border-radius: 4px;
    margin: 1rem 0;
    border: 1px solid #b3d9ff;
  }

  .info {
    color: #666;
    font-size: 0.9rem;
    margin-top: 1rem;
  }

  h3 {
    margin-bottom: 1rem;
    color: #333;
  }
</style> 