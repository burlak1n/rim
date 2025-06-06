// Функция для определения базового URL API в зависимости от текущего хоста
function getApiBaseUrl(): string {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL;
  }
  
  // В dev режиме используем прокси (относительные пути)
  if (import.meta.env.DEV) {
    return '';
  }
  
  // В браузере определяем текущий хост для production
  if (typeof window !== 'undefined') {
    const currentHost = window.location.hostname;
    const port = currentHost === 'localhost.local' ? '3000' : '3000';
    return `http://${currentHost}:${port}`;
  }
  
  // Fallback для SSR или других случаев
  return 'http://localhost:3000';
}

// Конфигурация приложения
export const config = {
  // Telegram Bot Username - берется из переменной окружения
  botUsername: import.meta.env.VITE_BOT_USERNAME || 'test_burlak1n_bot',
  
  // API Base URL - динамически определяется
  apiBaseUrl: getApiBaseUrl(),
}; 