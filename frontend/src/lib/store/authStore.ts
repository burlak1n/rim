import { writable } from 'svelte/store';
import type { User } from '../types.js';
import AuthService from '../services/authService.js';

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  loading: boolean;
  debugMode: boolean; // Отладочный режим для разрешения редактирования всем
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null,
  loading: false,
  debugMode: false,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  return {
    subscribe,
    
    // Инициализация при загрузке приложения
    init: async () => {
      let debugModeFromServer = false;
      
      // Загружаем состояние отладочного режима с сервера
      try {
        const debugResponse = await AuthService.getDebugMode();
        debugModeFromServer = debugResponse.enabled;
      } catch (error) {
        console.warn('Failed to load debug mode from server:', error);
      }

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
            debugMode: debugModeFromServer,
          });
        } catch (error) {
          console.error('Failed to get user info:', error);
          AuthService.removeToken();
          set({
            ...initialState,
            debugMode: debugModeFromServer,
          });
        }
      } else {
        set({
          ...initialState,
          debugMode: debugModeFromServer,
        });
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
          debugMode: false,
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

    // Обновление данных пользователя
    loadUser: async () => {
      const token = AuthService.getToken();
      if (token) {
        try {
          const user = await AuthService.getMe(token);
          update(state => ({ ...state, user }));
        } catch (error) {
          console.error('Failed to reload user info:', error);
          throw error;
        }
      }
    },

    // Переключение отладочного режима
    toggleDebugMode: async () => {
      const token = AuthService.getToken();
      if (!token) {
        console.error('No token available for debug mode toggle');
        return;
      }

      try {
        // Сначала получаем текущее состояние для переключения
        let currentDebugMode = false;
        update(state => {
          currentDebugMode = state.debugMode;
          return state;
        });

        const newDebugMode = !currentDebugMode;
        
        // Сохраняем на сервере
        await AuthService.setDebugMode(token, newDebugMode);
        
        // Обновляем локальное состояние
        update(state => ({ ...state, debugMode: newDebugMode }));
        
        console.log('Debug mode updated:', newDebugMode);
      } catch (error) {
        console.error('Failed to toggle debug mode:', error);
        // Можно показать уведомление об ошибке
      }
    },

    // Очистка состояния
    clear: () => {
      AuthService.removeToken();
      set(initialState);
    }
  };
}

export const authStore = createAuthStore(); 