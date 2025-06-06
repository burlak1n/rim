import type { TelegramAuthData, SessionResponse, User } from '../types.js';
import { config } from '../config.js';

const API_BASE_URL = `http://localhost:3000/api/v1/auth`; // Базовый URL для эндпоинтов аутентификации

// Глобальное хранение CSRF токена
let csrfToken: string | null = null;

// Получить CSRF токен
async function getCSRFToken(): Promise<string | null> {
  if (csrfToken) {
    return csrfToken;
  }
  
  try {
    const response = await fetch(`${API_BASE_URL}/csrf-token`, {
      method: 'GET',
      credentials: 'include', // Важно для cookies
    });
    
    if (response.ok) {
      const data = await response.json();
      csrfToken = data.csrf_token;
      return csrfToken;
    }
  } catch (error) {
    console.warn('Failed to get CSRF token:', error);
  }
  
  return null;
}

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

// Создать заголовки с CSRF токеном
async function createSecureHeaders(includeCSRF: boolean = true): Promise<HeadersInit> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  
  if (includeCSRF) {
    const token = await getCSRFToken();
    if (token) {
      headers['X-CSRF-Token'] = token;
    }
  }
  
  return headers;
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
      credentials: 'include', // Важно для cookies
      body: JSON.stringify(authData),
    });
    const result = await handleResponse(response);
    
    // После успешной авторизации получаем CSRF токен
    await getCSRFToken();
    
    return result;
  },

  // Получить информацию о текущем пользователе
  getMe: async (token?: string): Promise<User> => {
    console.log('AuthService.getMe called');
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    
    // Поддерживаем обратную совместимость с token
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(`${API_BASE_URL}/me`, {
      method: 'GET',
      headers,
      credentials: 'include', // Используем cookies
    });
    return handleResponse(response);
  },

  // Выход из системы
  logout: async (token?: string): Promise<void> => {
    console.log('AuthService.logout called');
    const headers: Record<string, string> = await createSecureHeaders() as Record<string, string>;
    
    // Поддерживаем обратную совместимость с token
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(`${API_BASE_URL}/logout`, {
      method: 'POST',
      headers,
      credentials: 'include', // Используем cookies
    });
    
    // Очищаем CSRF токен после логаута
    csrfToken = null;
    
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
  },

  // Обновить свой контакт
  updateMyContact: async (token: string, contactData: any): Promise<any> => {
    console.log('AuthService.updateMyContact called', contactData);
    const headers: Record<string, string> = await createSecureHeaders() as Record<string, string>;
    
    // Поддерживаем обратную совместимость с token
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(`${API_BASE_URL}/contact`, {
      method: 'PUT',
      headers,
      credentials: 'include', // Используем cookies
      body: JSON.stringify(contactData),
    });
    return handleResponse(response);
  },

  // Получить состояние отладочного режима
  getDebugMode: async (): Promise<{ enabled: boolean }> => {
    console.log('AuthService.getDebugMode called');
    const response = await fetch('http://localhost:3000/api/v1/system/debug-mode', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include', // Используем cookies
    });
    return handleResponse(response);
  },

  // Установить состояние отладочного режима
  setDebugMode: async (token: string, enabled: boolean): Promise<{ enabled: boolean }> => {
    console.log('AuthService.setDebugMode called', enabled);
    const headers: Record<string, string> = await createSecureHeaders() as Record<string, string>;
    
    // Поддерживаем обратную совместимость с token
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch('http://localhost:3000/api/v1/system/debug-mode', {
      method: 'PUT',
      headers,
      credentials: 'include', // Используем cookies
      body: JSON.stringify({ enabled }),
    });
    return handleResponse(response);
  }
};

export default AuthService; 