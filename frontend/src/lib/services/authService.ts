// Временная заглушка для AuthService
// В будущем здесь будет логика взаимодействия с API бэкенда для аутентификации

const API_BASE_URL = '/api/auth'; // Базовый URL для эндпоинтов аутентификации

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
  login: async (email, password) => {
    console.log('AuthService.login called with', email);
    const response = await fetch(`${API_BASE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });
    return handleResponse(response);
  },

  register: async (userData) => {
    console.log('AuthService.register called with', userData);
    const response = await fetch(`${API_BASE_URL}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(userData),
    });
    return handleResponse(response);
  },

  logout: async (token: string | null) => { // Logout может потребовать токен
    console.log('AuthService.logout called');
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE_URL}/logout`, {
      method: 'POST',
      headers: headers,
      // Тело запроса для logout обычно не требуется, но это зависит от API
    });
    return handleResponse(response);
  }
};

export default AuthService; 