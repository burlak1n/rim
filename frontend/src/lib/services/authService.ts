import type { TelegramAuthData, SessionResponse, User } from '../types.js';
import { config } from '../config.js';

const API_BASE_URL = `${config.apiBaseUrl}/api/v1/auth`; // Базовый URL для эндпоинтов аутентификации

// Вспомогательная функция для обработки ответов fetch
async function handleResponse(response: Response) {
  if (!response.ok) {
    let errorMessage = `HTTP error! status: ${response.status}`;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || errorMessage;
    } catch (e) {
      // Если тело ответа не JSON или пустое, используем стандартное сообщение
    }
    throw new Error(errorMessage);
  }
  // Если ответ 204 No Content, возвращаем null, так как .json() вызовет ошибку
  if (response.status === 204) {
    return null;
  }
  return response.json();
}

const AuthService = {
  // Авторизация через Telegram
  authenticateWithTelegram: async (authData: TelegramAuthData): Promise<SessionResponse> => {
    console.log('AuthService.authenticateWithTelegram called', authData);
    const response = await fetch(`${API_BASE_URL}/telegram`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(authData),
    });
    return handleResponse(response);
  },

  // Получить информацию о текущем пользователе
  getMe: async (token: string): Promise<User> => {
    console.log('AuthService.getMe called');
    const response = await fetch(`${API_BASE_URL}/me`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    });
    return handleResponse(response);
  },

  // Выход из системы
  logout: async (token: string): Promise<void> => {
    console.log('AuthService.logout called');
    const response = await fetch(`${API_BASE_URL}/logout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    });
    return handleResponse(response);
  },

  // Сохранить токен в localStorage
  saveToken: (token: string): void => {
    localStorage.setItem('session_token', token);
  },

  // Получить токен из localStorage
  getToken: (): string | null => {
    return localStorage.getItem('session_token');
  },

  // Удалить токен из localStorage
  removeToken: (): void => {
    localStorage.removeItem('session_token');
  }
};

export default AuthService; 