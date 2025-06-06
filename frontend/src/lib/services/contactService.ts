// frontend/src/lib/services/contactService.ts
import type { Contact, ContactPayload, ContactBasic } from "$lib/types";
import AuthService from "./authService.js";

const API_BASE_URL = 'http://localhost:3000/api/v1/contacts';

// Получить CSRF токен (аналогично authService)
async function getCSRFToken(): Promise<string | null> {
  try {
    const response = await fetch(`http://localhost:3000/api/v1/auth/csrf-token`, {
      method: 'GET',
      credentials: 'include',
    });
    
    if (response.ok) {
      const data = await response.json();
      return data.csrf_token;
    }
  } catch (error) {
    console.warn('Failed to get CSRF token:', error);
  }
  
  return null;
}

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

// Функция для получения заголовков с CSRF токеном
async function getSecureHeaders(includeCSRF: boolean = true): Promise<Record<string, string>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  
  // Поддержка обратной совместимости с localStorage токенами
  const token = AuthService.getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  // Добавляем CSRF токен для изменяющих операций
  if (includeCSRF) {
    const csrfToken = await getCSRFToken();
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }
  }
  
  return headers;
}

const ContactService = {
  createContact: async (contactData: ContactPayload): Promise<Contact> => {
    const response = await fetch(API_BASE_URL, {
      method: 'POST',
      headers: await getSecureHeaders(),
      credentials: 'include',
      body: JSON.stringify(contactData),
    });
    return handleResponse<Contact>(response);
  },

  getAllContacts: async (): Promise<Contact[] | ContactBasic[]> => {
    const response = await fetch(API_BASE_URL, {
      method: 'GET',
      headers: await getSecureHeaders(false), // GET не требует CSRF
      credentials: 'include',
    });
    return handleResponse<Contact[] | ContactBasic[]>(response);
  },

  getContactById: async (id: string): Promise<Contact> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'GET',
      headers: await getSecureHeaders(false), // GET не требует CSRF
      credentials: 'include',
    });
    return handleResponse<Contact>(response);
  },

  updateContact: async (id: string, contactData: Partial<ContactPayload>): Promise<Contact> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'PUT',
      headers: await getSecureHeaders(),
      credentials: 'include',
      body: JSON.stringify(contactData),
    });
    return handleResponse<Contact>(response);
  },

  deleteContact: async (id: string): Promise<null> => {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'DELETE',
      headers: await getSecureHeaders(),
      credentials: 'include',
    });
    return handleResponse<null>(response);
  },

  addContactToGroup: async (contactId: string, groupId: string): Promise<any> => {
    const response = await fetch(`${API_BASE_URL}/${contactId}/groups/${groupId}`, {
      method: 'POST',
      headers: await getSecureHeaders(),
      credentials: 'include',
    });
    return handleResponse<any>(response);
  },

  removeContactFromGroup: async (contactId: string, groupId: string): Promise<any> => {
    const response = await fetch(`${API_BASE_URL}/${contactId}/groups/${groupId}`, {
      method: 'DELETE',
      headers: await getSecureHeaders(),
      credentials: 'include',
    });
    return handleResponse<any>(response);
  },
};

export default ContactService; 