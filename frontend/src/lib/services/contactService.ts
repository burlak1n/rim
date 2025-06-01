// frontend/src/lib/services/contactService.ts
// import AuthStore from "$lib/store/authStore"; // Временно не используется для getAuthHeaders
import type { Contact, ContactPayload, Group } from "$lib/types";

const API_BASE_URL = 'http://localhost:3000/api/v1/contacts';

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

const ContactService = {
  createContact: async (contactData: ContactPayload): Promise<Contact> => {
    const response = await fetch(API_BASE_URL, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(contactData),
    });
    return handleResponse<Contact>(response);
  },

  getAllContacts: async (): Promise<Contact[]> => {
    const response = await fetch(API_BASE_URL, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse<Contact[]>(response);
  },

  getContactById: async (id: string): Promise<Contact> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse<Contact>(response);
  },

  updateContact: async (id: string, contactData: Partial<ContactPayload>): Promise<Contact> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(contactData),
    });
    return handleResponse<Contact>(response);
  },

  deleteContact: async (id: string): Promise<null> => { // Обычно DELETE не возвращает тело
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse<null>(response);
  },

  addContactToGroup: async (contactId: string, groupId: string): Promise<any> => { // Тип ответа зависит от API
    const response = await fetch(`${API_BASE_URL}/${contactId}/groups/${groupId}`, {
      method: 'POST',
      headers: getAuthHeaders(),
    });
    return handleResponse<any>(response);
  },

  removeContactFromGroup: async (contactId: string, groupId: string): Promise<any> => { // Тип ответа зависит от API
    const response = await fetch(`${API_BASE_URL}/${contactId}/groups/${groupId}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    return handleResponse<any>(response);
  },
};

export default ContactService; 