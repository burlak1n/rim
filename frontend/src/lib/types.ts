// frontend/src/lib/types.ts

// На основе ТЗ
export interface Contact {
  id: string; // Обычно ID это строка (uuid) или число
  name: string;
  phone: string;
  email: string;
  transport?: 'есть машина' | 'есть права' | 'нет ничего' | string; // string для случая, если приходят другие значения
  printer?: 'цветной' | 'обычный' | 'нет' | string;
  allergies?: string;
  vk?: string;
  telegram?: string;
  groups?: Group[]; // Контакты могут принадлежать к группам
  // createdAt?: string; // Даты создания/обновления, если они есть в API
  // updatedAt?: string;
}

export interface ContactPayload { // Данные для создания/обновления контакта
  name: string;
  phone: string;
  email: string;
  transport?: 'есть машина' | 'есть права' | 'нет ничего' | string;
  printer?: 'цветной' | 'обычный' | 'нет' | string;
  allergies?: string;
  vk?: string;
  telegram?: string;
  // groupIds?: string[]; // Если при обновлении контакта можно менять группы
}

export interface Group {
  id: string;
  name: string;
  contacts?: Contact[]; // Группы могут содержать контакты
  // createdAt?: string;
  // updatedAt?: string;
}

export interface GroupPayload { // Данные для создания/обновления группы
  name: string;
  // description?: string; // Удалено по запросу пользователя
}

// Базовая информация о контакте для неавторизованных пользователей
export interface ContactBasic {
  id: string;
  name: string;
}

// Данные авторизации через Telegram
export interface TelegramAuthData {
  id: number;
  first_name?: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
  auth_date: number;
  hash: string;
}

// Ответ с токеном сессии
export interface SessionResponse {
  session_token: string;
  expires_at: string;
}

// Информация о пользователе
export interface User {
  id: number;
  telegram_id: number;
  is_active: boolean;
  contact?: Contact;
  created_at: string;
}

// Общий тип для элементов, имеющих ID, полезно для списков
export interface Identifiable {
  id: string | number;
} 