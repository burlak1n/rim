// frontend/src/lib/services/contactService.ts
import type { Contact, ContactPayload, ContactBasic } from "$lib/types";
import AuthService from "./authService.js";
import { config } from "../config.js";

const API_BASE_URL = `${config.apiBaseUrl}/api/v1/contacts`;

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

// Функция для получения заголовков с токеном авторизации
function getAuthHeaders(): HeadersInit {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  const token = AuthService.getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
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

  getAllContacts: async (): Promise<Contact[] | ContactBasic[]> => {
    const response = await fetch(API_BASE_URL, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    return handleResponse<Contact[] | ContactBasic[]>(response);
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