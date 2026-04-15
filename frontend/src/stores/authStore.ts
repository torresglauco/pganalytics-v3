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
  // ❌ REMOVED: localStorage.getItem('auth_token')
  // ✅ NEW: Token is now in httpOnly cookie, not in localStorage
  token: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,

  setUser: (user) => set({ user }),
  setToken: (token) => {
    // ❌ REMOVED: localStorage.setItem('auth_token', token)
    // ✅ NEW: Token is stored in httpOnly cookie by backend
    // Frontend only tracks that user is authenticated
    set({ token: '', isAuthenticated: true })
  },
  setAuthenticated: (isAuthenticated) => set({ isAuthenticated }),
  setLoading: (isLoading) => set({ isLoading }),
  setError: (error) => set({ error }),
  logout: () => {
    // ❌ REMOVED: localStorage.removeItem('auth_token')
    // ✅ NEW: Backend clears httpOnly cookie on logout
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
