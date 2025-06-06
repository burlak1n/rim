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
    <Link to="/" class="logo-link">RIM</Link>
    <ul class="nav-links">
      <li><Link to="/contacts">Контакты</Link></li>
      <li><Link to="/groups">Группы</Link></li>
      
      {#if isAuthenticated}
        <li><Link to="/profile">Профиль</Link></li>
        <li><button on:click={handleLogout}>Выйти</button></li>
      {:else}
        <li><Link to="/login">Войти</Link></li>
      {/if}
    </ul>
  </div>
</nav>

<style>
  nav {
    background-color: #001529;
    padding: 0 20px;
    color: white;
  }
  .nav-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    max-width: 1200px;
    margin: 0 auto;
    height: 64px;
  }
  .logo-link {
    font-size: 1.5rem;
    font-weight: bold;
    color: white;
    text-decoration: none;
  }
  .nav-links {
    list-style: none;
    display: flex;
    margin: 0;
    padding: 0;
  }
  .nav-links li {
    margin-left: 20px;
  }
  .nav-links a,
  .nav-links button {
    color: white;
    text-decoration: none;
    background: none;
    border: none;
    font-size: 1rem;
    cursor: pointer;
    padding: 8px 12px;
    border-radius: 4px;
    transition: background-color 0.2s ease-in-out;
  }
  .nav-links a:hover,
  .nav-links button:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
</style> 