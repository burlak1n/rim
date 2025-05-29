<script lang="ts">
  import { navigate } from "svelte-routing";
  import AuthService from "$lib/services/authService";
  import AuthStore from "$lib/store/authStore"; // Импортируем для setError

  let name = '';
  let email = '';
  let password = '';
  let confirmPassword = '';
  let error = '';
  let isLoading = false;

  async function handleSubmit() {
    error = '';
    isLoading = true;
    if (!name || !email || !password || !confirmPassword) {
      error = 'Пожалуйста, заполните все поля.';
      isLoading = false;
      return;
    }
    if (password !== confirmPassword) {
      error = 'Пароли не совпадают.';
      isLoading = false;
      return;
    }
    try {
      const response = await AuthService.register({ name, email, password });
      console.log('Registration successful', response);
      // Предполагаем, что API регистрации не возвращает токен и не логинит пользователя автоматически
      // А просто создает аккаунт. После успешной регистрации перенаправляем на логин.
      // Если API ведет себя иначе (например, сразу логинит), эту логику нужно будет изменить.
      navigate('/login', { replace: true }); 
      // Можно показать сообщение об успехе перед редиректом, если это необходимо
    } catch (err: any) {
      console.error('Registration failed:', err);
      error = err.message || 'Ошибка регистрации. Пожалуйста, попробуйте еще раз.';
      AuthStore.setError(error);
    }
    isLoading = false;
  }
</script>

<div class="register-container">
  <h2>Регистрация</h2>
  <form on:submit|preventDefault={handleSubmit}>
    <div class="form-group">
      <label for="name">Имя:</label>
      <input type="text" id="name" bind:value={name} required disabled={isLoading}>
    </div>
    <div class="form-group">
      <label for="email">Email:</label>
      <input type="email" id="email" bind:value={email} required disabled={isLoading}>
    </div>
    <div class="form-group">
      <label for="password">Пароль:</label>
      <input type="password" id="password" bind:value={password} required disabled={isLoading}>
    </div>
    <div class="form-group">
      <label for="confirmPassword">Подтвердите пароль:</label>
      <input type="password" id="confirmPassword" bind:value={confirmPassword} required disabled={isLoading}>
    </div>
    {#if error}
      <p class="error-message">{error}</p>
    {/if}
    <button type="submit" disabled={isLoading}>
      {#if isLoading}Регистрация...{:else}Зарегистрироваться{/if}
    </button>
  </form>
  <p class="login-link">
    Уже есть аккаунт? <a href="/login">Войти</a>
  </p>
</div>

<style>
  .register-container {
    max-width: 450px;
    margin: 40px auto;
    padding: 30px;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  h2 {
    text-align: center;
    margin-bottom: 25px;
  }
  .form-group {
    margin-bottom: 20px;
  }
  label {
    display: block;
    margin-bottom: 8px;
    font-weight: 500;
  }
  button[type="submit"] {
    width: 100%;
    padding: 12px;
    background-color: #52c41a;
    color: white;
    border: none;
    font-size: 1rem;
    font-weight: 500;
  }
  button[type="submit"]:hover {
    background-color: #73d13d;
  }
  button[type="submit"]:disabled {
    background-color: #b7eb8f;
    cursor: not-allowed;
  }
  .login-link {
    text-align: center;
    margin-top: 20px;
    font-size: 0.9rem;
  }
</style> 