<script lang="ts">
  import { Link, navigate } from "svelte-routing";
  // import AuthStore from "$lib/store/authStore"; // Временно не используем
  // import AuthService from "$lib/services/authService"; // Временно не используем
  // import { tick } from 'svelte'; 

  // let isAuthenticated: boolean; // Временно не используем
  // let currentToken: string | null = null; // Временно не используем

  // AuthStore.subscribe(value => { // Временно не используем
  //   isAuthenticated = value.isAuthenticated;
  //   currentToken = value.token;
  // });

  // async function handleLogout() { // Временно не используем
  //   if (!currentToken) {
  //     AuthStore.logout();
  //     await tick();
  //     navigate("/login", { replace: true });
  //     return;
  //   }
  //   try {
  //     await AuthService.logout(currentToken);
  //     console.log("Logout successful on server");
  //   } catch (err: any) {
  //     console.error("Logout failed on server:", err.message);
  //     AuthStore.setError(err.message || 'Ошибка при выходе на сервере.');
  //   } finally {
  //     AuthStore.logout(); 
  //     await tick(); 
  //     navigate("/login", { replace: true });
  //   }
  // }
</script>

<nav>
  <div class="nav-container">
    <img style="margin:20px;" src="src/pictures/logo.png" width="100" height=auto alt="Логотип" class="logo" />
    <Link to="/" class="logo-link">RIM</Link>
    <ul class="nav-links">
      <!-- Временно показываем всегда -->
      <li><Link to="/contacts">Контакты</Link></li>
      <li><Link to="/groups">Группы</Link></li>
      
      <!-- {#if isAuthenticated} Временно скрыто
        <li><Link to="/profile">Профиль</Link></li>
        <li><button on:click={handleLogout}>Выйти</button></li>
      {:else}
        <li><Link to="/login">Войти</Link></li>
        <li><Link to="/register">Регистрация</Link></li>
      {/if} -->
    </ul>
  </div>
</nav>

<style>
  nav {
    background-color: #593494;
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