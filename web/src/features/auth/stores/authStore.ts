import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

export interface User {
  id: string;
  name: string;
  email?: string;
  auth_type: 'password' | 'google' | 'guest';
  is_guest: boolean;
}

export interface AuthStore {
  // Auth state
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  
  // Auth methods availability
  ssoEnabled: boolean;
  authMethods: {
    password: boolean;
    google: boolean;
    guest: boolean;
  };

  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  loginWithGoogle: (accessToken: string) => Promise<void>;
  loginAsGuest: (name: string) => Promise<void>;
  logout: () => Promise<void>;
  checkAuthStatus: () => Promise<void>;
  validateToken: () => Promise<boolean>;
  setError: (error: string | null) => void;
  clearError: () => void;
}

const API_BASE = process.env.NODE_ENV === 'production' 
  ? '/api/v1' 
  : 'http://localhost:8080/api/v1';

export const useAuthStore = create<AuthStore>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
        ssoEnabled: true,
        authMethods: {
          password: true,
          google: true,
          guest: true,
        },

        // Actions
        login: async (email: string, password: string) => {
          set({ isLoading: true, error: null }, false, 'auth-login-start');
          
          try {
            const response = await fetch(`${API_BASE}/auth/login`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({ email, password }),
            });

            const data = await response.json();

            if (!response.ok) {
              throw new Error(data.error || 'Login failed');
            }

            set({
              user: data.user,
              token: data.token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            }, false, 'auth-login-success');

          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'Login failed',
              isLoading: false,
            }, false, 'auth-login-error');
            throw error;
          }
        },

        register: async (name: string, email: string, password: string) => {
          set({ isLoading: true, error: null }, false, 'auth-register-start');
          
          try {
            const response = await fetch(`${API_BASE}/auth/register`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({ name, email, password }),
            });

            const data = await response.json();

            if (!response.ok) {
              throw new Error(data.error || 'Registration failed');
            }

            set({
              user: data.user,
              token: data.token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            }, false, 'auth-register-success');

          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'Registration failed',
              isLoading: false,
            }, false, 'auth-register-error');
            throw error;
          }
        },

        loginWithGoogle: async (accessToken: string) => {
          set({ isLoading: true, error: null }, false, 'auth-google-start');
          
          try {
            const response = await fetch(`${API_BASE}/auth/google`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({ access_token: accessToken }),
            });

            const data = await response.json();

            if (!response.ok) {
              throw new Error(data.error || 'Google login failed');
            }

            set({
              user: data.user,
              token: data.token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            }, false, 'auth-google-success');

          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'Google login failed',
              isLoading: false,
            }, false, 'auth-google-error');
            throw error;
          }
        },

        loginAsGuest: async (name: string) => {
          set({ isLoading: true, error: null }, false, 'auth-guest-start');
          
          try {
            const response = await fetch(`${API_BASE}/auth/guest`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({ name }),
            });

            const data = await response.json();

            if (!response.ok) {
              throw new Error(data.error || 'Guest login failed');
            }

            set({
              user: data.user,
              token: data.token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            }, false, 'auth-guest-success');

          } catch (error) {
            set({
              error: error instanceof Error ? error.message : 'Guest login failed',
              isLoading: false,
            }, false, 'auth-guest-error');
            throw error;
          }
        },

        logout: async () => {
          const { token } = get();
          
          if (token) {
            try {
              await fetch(`${API_BASE}/auth/logout`, {
                method: 'POST',
                headers: {
                  'Authorization': `Bearer ${token}`,
                },
              });
            } catch (error) {
              console.error('Logout error:', error);
            }
          }

          set({
            user: null,
            token: null,
            isAuthenticated: false,
            error: null,
          }, false, 'auth-logout');
        },

        checkAuthStatus: async () => {
          try {
            const response = await fetch(`${API_BASE}/auth/status`);
            const data = await response.json();

            set({
              ssoEnabled: data.sso_enabled,
              authMethods: data.methods,
            }, false, 'auth-status-checked');

          } catch (error) {
            console.error('Failed to check auth status:', error);
          }
        },

        validateToken: async () => {
          const { token } = get();
          if (!token) return false;

          try {
            const response = await fetch(`${API_BASE}/auth/validate`, {
              headers: {
                'Authorization': `Bearer ${token}`,
              },
            });

            if (response.ok) {
              const data = await response.json();
              set({ user: data.user }, false, 'token-validated');
              return true;
            } else {
              // Token invalid, clear auth state
              set({
                user: null,
                token: null,
                isAuthenticated: false,
              }, false, 'token-invalid');
              return false;
            }
          } catch (error) {
            console.error('Token validation error:', error);
            return false;
          }
        },

        setError: (error: string | null) => {
          set({ error }, false, 'auth-set-error');
        },

        clearError: () => {
          set({ error: null }, false, 'auth-clear-error');
        },
      }),
      {
        name: 'auth-store',
        partialize: (state) => ({
          user: state.user,
          token: state.token,
          isAuthenticated: state.isAuthenticated,
        }),
      }
    ),
    {
      name: 'auth-store',
    }
  )
);
