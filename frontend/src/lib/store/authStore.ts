import { writable } from 'svelte/store';

interface User {
  id: string;
  name: string;
  email: string;
  // Можно добавить другие поля пользователя, которые приходят от бэкенда
  vk?: string;
  telegram?: string;
  transport?: string;
  printer?: string;
  allergies?: string;
  // roles?: string[]; // Если роли будут использоваться на фронте
}

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  error: string | null;
}

const initialAuthState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null, // В реальном приложении токен может быть загружен из localStorage
  error: null,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialAuthState);

  return {
    subscribe,
    login: (userData: User, authToken: string) => {
      update(state => ({
        ...state,
        isAuthenticated: true,
        user: userData,
        token: authToken,
        error: null,
      }));
      // localStorage.setItem('authToken', authToken); // Сохраняем токен
      // localStorage.setItem('user', JSON.stringify(userData)); // Сохраняем пользователя
    },
    logout: () => {
      update(state => ({
        ...state,
        isAuthenticated: false,
        user: null,
        token: null,
        error: null,
      }));
      // localStorage.removeItem('authToken');
      // localStorage.removeItem('user');
    },
    setError: (errorMessage: string) => {
      update(state => ({
        ...state,
        error: errorMessage,
      }));
    },
    // Можно добавить функцию для инициализации store из localStorage при загрузке приложения
    // init: () => {
    //   const token = localStorage.getItem('authToken');
    //   const userString = localStorage.getItem('user');
    //   if (token && userString) {
    //     try {
    //       const user = JSON.parse(userString);
    //       set({ isAuthenticated: true, user, token, error: null });
    //     } catch (e) {
    //       console.error("Failed to parse user from localStorage", e);
    //       // Очистить localStorage, если данные повреждены
    //       localStorage.removeItem('authToken');
    //       localStorage.removeItem('user');
    //       set(initialAuthState);
    //     }
    //   } else {
    //     set(initialAuthState);
    //   }
    // }
  };
}

const AuthStore = createAuthStore();
// AuthStore.init(); // Вызывать при инициализации приложения, например в App.svelte или main.ts

export default AuthStore; 