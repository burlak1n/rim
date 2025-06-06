<script lang="ts">
  import { onMount } from "svelte";
  import { Router, Route, Link } from "svelte-routing";
  import Home from "$lib/pages/Home.svelte";
  import Login from "$lib/pages/Login.svelte";
  import Register from "$lib/pages/Register.svelte";
  import Contacts from "$lib/pages/Contacts.svelte";
  import Groups from "$lib/pages/Groups.svelte";
  import Profile from "$lib/pages/Profile.svelte";
  import Navbar from "$lib/components/Navbar.svelte";
  import { authStore } from "$lib/store/authStore";

  export let url = ""; // Required by svelte-routing

  // Инициализируем auth store при загрузке приложения
  onMount(() => {
    authStore.init();
  });
</script>

<Router {url}>
  <Navbar />
  <div class="container">
    <Route path="/login" component={Login} />
    <Route path="/register" component={Register} />
    <Route path="/contacts" component={Contacts} />
    <Route path="/groups" component={Groups} />
    <Route path="/profile" component={Profile} />
    <Route path="/" component={Home} />
    <!-- Можно добавить Route path="*" для страницы 404 -->
  </div>
</Router>

<style>
  /* Стили для App.svelte, если нужны */
  .container {
    padding-top: 20px; /* Отступ от навбара */
  }
</style>
