<script lang="ts">
  import { Link, navigate } from "svelte-routing";
  import { authStore } from "$lib/store/authStore";
  import { onMount } from 'svelte';

  // Реактивные переменные для отслеживания состояния аутентификации
  $: authState = $authStore;
  $: isAuthenticated = authState.isAuthenticated;
  $: currentUser = authState.user;

  onMount(() => {
    authStore.init();
  });

  async function handleLogout() {
    try {
      await authStore.logout();
      navigate("/", { replace: true });
    } catch (error) {
      console.error("Logout error:", error);
    }
  }
</script>

<nav>
  <div class="nav-container">
    <a href="https://ingroupctc.ru/">
      <img style="margin:20px;" src="src/pictures/logo.png" width="100" height=auto alt="Логотип" class="logo" />
    </a>
    <Link to="/" class="logo-link">РИМ</Link>
    <nav class="flex gap-4 items-center">
      <Link to="/contacts" class="hover:text-blue-600 transition-colors">Контакты</Link>
      <Link to="/groups" class="hover:text-blue-600 transition-colors">Группы</Link>
      <Link to="/profile" class="hover:text-blue-600 transition-colors">Профиль</Link>
    </nav>
  </div>
</nav>

<style>
  nav {
    background-color: #593494;
    padding: 0 20px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  }
  .nav-container {
    display: flex;
    align-items: center;
    justify-content: space-between;
    max-width: 1200px;
    margin: 0 auto;
  }
  .logo-link {
    font-size: 1.5rem;
    font-weight: bold;
    color: white;
    text-decoration: none;
  }
  .nav-links {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 20px;
  }
  .nav-links li {
    display: flex;
    align-items: center;
  }
  .nav-links :global(a) {
    color: white;
    text-decoration: none;
    padding: 15px 10px;
    transition: background-color 0.3s;
  }
  .nav-links :global(a:hover) {
    background-color: rgba(255,255,255,0.1);
    border-radius: 4px;
  }
  
  /* Стилизация кнопки выйти под навигационные ссылки */
  .logout-button {
    background: none;
    border: none;
    color: white;
    text-decoration: none;
    padding: 15px 10px;
    transition: background-color 0.3s;
    cursor: pointer;
    font-size: inherit;
    font-family: inherit;
  }
  .logout-button:hover {
    background-color: rgba(255,255,255,0.1);
    border-radius: 4px;
  }
</style> 