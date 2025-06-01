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

// Общий тип для элементов, имеющих ID, полезно для списков
export interface Identifiable {
  id: string | number;
} 