import { create } from 'zustand'
import { User } from '../types/auth'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  // Actions
  setUser: (user: User) => void
  setToken: (token: string) => void
  setAuthenticated: (isAuthenticated: boolean) => void
  setLoading: (isLoading: boolean) => void
  setError: (error: string | null) => void
  logout: () => void
  reset: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  // WHY: Token is stored in httpOnly cookie by backend for security.
  // httpOnly cookies cannot be accessed by JavaScript, preventing XSS
  // token theft. The frontend only tracks authentication state, not
  // the token itself. Backend validates the cookie on each request.
  token: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,

  setUser: (user) => set({ user }),
  setToken: (token) => {
    // WHY: Token is stored in httpOnly cookie by backend.
    // We set an empty string to indicate "logged in" without
    // storing the actual token, which is in the secure cookie.
    set({ token: '', isAuthenticated: true })
  },
  setAuthenticated: (isAuthenticated) => set({ isAuthenticated }),
  setLoading: (isLoading) => set({ isLoading }),
  setError: (error) => set({ error }),
  logout: () => {
    // WHY: We clear frontend state, but the httpOnly cookie is cleared
    // by the backend's /auth/logout endpoint. This ensures the cookie
    // is properly invalidated on the server side.
    set({
      user: null,
      token: null,
      isAuthenticated: false,
      error: null,
    })
  },
  reset: () => set({
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
  }),
}))
