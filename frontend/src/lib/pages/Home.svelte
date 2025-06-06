<script lang="ts">
  // Логика для главной страницы (пока пусто)
</script>

<div class="overlay-container">
  <img src="src/pictures/IMG_0933.JPG" class="background-img" />
  <div class="content">
    <h1>Добро пожаловать в РИМ!</h1>
    <p>Корпоративный портал для управления контактами, группами и ресурсами.</p>
  </div>
</div>

<style>
  .overlay-container {
    position: relative;
    width: 100%;
    height: 100vh;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .background-img {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    z-index: 1;
  }

  .content {
    position: relative;
    z-index: 2;
    color: white;
    text-align: center;
    padding: 20px;
    background: rgba(0, 0, 0, 0.5);  /* Полупрозрачный фон для лучшей читаемости */
    border-radius: 10px;
  }
  h1, p {
    text-shadow: 0 0 10px white;
  }
</style>