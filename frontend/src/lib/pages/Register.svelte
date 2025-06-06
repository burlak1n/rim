<script lang="ts">
  import TelegramAuth from "$lib/components/TelegramAuth.svelte";
</script>

<div class="register-container">
  <h2>Регистрация</h2>
  <TelegramAuth />
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