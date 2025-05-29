// frontend/src/lib/services/groupService.ts
// import AuthStore from "$lib/store/authStore"; // Временно не используется для getAuthHeaders
import type { Group, GroupPayload } from "$lib/types";

const API_BASE_URL = '/api/v1/groups'; // Базовый URL для групп

// Вспомогательная функция для обработки ответов fetch (аналогична той, что в authService)
async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let errorMessage = `HTTP error! status: ${response.status}`;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || errorMessage;
    } catch (e) {
      //
    }
    throw new Error(errorMessage);
  }
  if (response.status === 204) {
    return null as T; // No Content
  }
  return response.json() as Promise<T>;
}

// Функция для получения токена из AuthStore
function getAuthHeaders(): HeadersInit {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  // ВРЕМЕННО УБРАНА ЛОГИКА АВТОРИЗАЦИИ
  // let currentToken: string | null = null;
  // const unsubscribe = AuthStore.subscribe(value => {
  //   currentToken = value.token;
  // });
  // unsubscribe(); 
  // if (currentToken) {
  //   headers['Authorization'] = `Bearer ${currentToken}`;
  // }
  return headers;
}

const GroupService = {
  createGroup: async (groupData: GroupPayload): Promise<Group> => {
    const response = await fetch(API_BASE_URL, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(groupData),
    });
    return handleResponse<Group>(response);
  },

  getAllGroups: async (): Promise<Group[]> => {
    const response = await fetch(API_BASE_URL, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse<Group[]>(response);
  },

  getGroupById: async (id: string): Promise<Group> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse<Group>(response);
  },

  updateGroup: async (id: string, groupData: Partial<GroupPayload>): Promise<Group> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(groupData),
    });
    return handleResponse<Group>(response);
  },

  deleteGroup: async (id: string): Promise<null> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse<null>(response); // Или просто response.ok, если тело ответа пустое
  },
  
  // Если для групп понадобятся специфичные методы (например, получение всех контактов группы),
  // их можно будет добавить сюда.
};

export default GroupService; 