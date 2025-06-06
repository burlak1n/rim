<script lang="ts">
  import TelegramAuth from "$lib/components/TelegramAuth.svelte";
</script>

<div class="login-container">
  <h2>Вход в систему</h2>
  <TelegramAuth />
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
    margin-bottom: 0px;
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