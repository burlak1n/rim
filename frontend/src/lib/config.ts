// Конфигурация приложения
export const config = {
  // Telegram Bot Username - берется из переменной окружения
  botUsername: import.meta.env.VITE_BOT_USERNAME || 'test_burlak1n_bot',
  
  // API Base URL
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:3000',
}; 