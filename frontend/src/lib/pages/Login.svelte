<script lang="ts">
  import { navigate } from "svelte-routing";
  import AuthStore from "$lib/store/authStore";
  import AuthService from "$lib/services/authService";

  let email = '';
  let password = '';
  let error = '';
  let isLoading = false; // Для отображения состояния загрузки

  async function handleSubmit() {
    error = '';
    isLoading = true;
    if (!email || !password) {
      error = 'Пожалуйста, введите email и пароль.';
      isLoading = false;
      return;
    }
    try {
      const response = await AuthService.login(email, password);
      if (response && response.user && response.token) {
        AuthStore.login(response.user, response.token);
        navigate('/', { replace: true });
      } else {
        // Этот случай маловероятен, если handleResponse в сервисе работает корректно
        // и всегда выбрасывает ошибку или возвращает данные.
        error = 'Получен некорректный ответ от сервера.';
        AuthStore.setError(error);
      }
    } catch (err: any) {
      console.error('Login failed:', err);
      error = err.message || 'Ошибка входа. Пожалуйста, проверьте ваши данные.';
      AuthStore.setError(error);
    }
    isLoading = false;
  }
</script>

<div class="login-container">
  <h2>Вход в систему</h2>
  <form on:submit|preventDefault={handleSubmit}>
    <div class="form-group">
      <label for="email">Email:</label>
      <input type="email" id="email" bind:value={email} required disabled={isLoading}>
    </div>
    <div class="form-group">
      <label for="password">Пароль:</label>
      <input type="password" id="password" bind:value={password} required disabled={isLoading}>
    </div>
    {#if error}
      <p class="error-message">{error}</p>
    {/if}
    <button type="submit" disabled={isLoading}>
      {#if isLoading}Загрузка...{:else}Войти{/if}
    </button>
  </form>
  <p class="register-link">
    Нет аккаунта? <a href="/register">Зарегистрироваться</a>
  </p>
</div>

<style>
  .login-container {
    max-width: 400px;
    margin: 40px auto;
    padding: 30px;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  h2 {
    text-align: center;
    margin-bottom: 25px;
    color: #333;
  }
  .form-group {
    margin-bottom: 20px;
  }
  label {
    display: block;
    margin-bottom: 8px;
    font-weight: 500;
    color: #555;
  }
  button[type="submit"] {
    width: 100%;
    padding: 12px;
    background-color: #1890ff;
    color: white;
    border: none;
    font-size: 1rem;
    font-weight: 500;
  }
  button[type="submit"]:hover {
    background-color: #40a9ff;
  }
  button[type="submit"]:disabled {
    background-color: #a0d9ff;
    cursor: not-allowed;
  }
  .register-link {
    text-align: center;
    margin-top: 20px;
    font-size: 0.9rem;
  }
</style> 