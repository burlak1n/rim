import { writable } from 'svelte/store';
import type { User } from '../types.js';
import AuthService from '../services/authService.js';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  loading: boolean;
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null,
  loading: false,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  return {
    subscribe,
    
    // Инициализация при загрузке приложения
    init: async () => {
      const token = AuthService.getToken();
      if (token) {
        try {
          update(state => ({ ...state, loading: true }));
          const user = await AuthService.getMe(token);
          set({
            isAuthenticated: true,
            user,
            token,
            loading: false,
          });
        } catch (error) {
          console.error('Failed to get user info:', error);
          AuthService.removeToken();
          set(initialState);
        }
      }
    },

    // Авторизация через Telegram
    loginWithTelegram: async (authData: any) => {
      try {
        update(state => ({ ...state, loading: true }));
        const sessionResponse = await AuthService.authenticateWithTelegram(authData);
        AuthService.saveToken(sessionResponse.session_token);
        
        const user = await AuthService.getMe(sessionResponse.session_token);
        
        set({
          isAuthenticated: true,
          user,
          token: sessionResponse.session_token,
          loading: false,
        });
        
        return true;
      } catch (error) {
        console.error('Login failed:', error);
        set({ ...initialState, loading: false });
        throw error;
      }
    },

    // Выход из системы
    logout: async () => {
      const token = AuthService.getToken();
      if (token) {
        try {
          await AuthService.logout(token);
        } catch (error) {
          console.error('Logout error:', error);
        }
      }
      AuthService.removeToken();
      set(initialState);
    },

    // Очистка состояния
    clear: () => {
      AuthService.removeToken();
      set(initialState);
    }
  };
}

export const authStore = createAuthStore(); 