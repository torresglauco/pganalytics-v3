import { create } from 'zustand'

type Theme = 'light' | 'dark' | 'auto'

interface ThemeState {
  theme: Theme
  prefersDark: boolean

  // Actions
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
  setPrefersDark: (prefersDark: boolean) => void
}

const getStoredTheme = (): Theme => {
  return (localStorage.getItem('theme') as Theme) || 'auto'
}

const getPrefersDark = (): boolean => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  theme: getStoredTheme(),
  prefersDark: getPrefersDark(),

  setTheme: (theme: Theme) => {
    localStorage.setItem('theme', theme)
    set({ theme })
  },
  toggleTheme: () => {
    const { theme } = get()
    const newTheme: Theme = theme === 'light' ? 'dark' : 'light'
    localStorage.setItem('theme', newTheme)
    set({ theme: newTheme })
  },
  setPrefersDark: (prefersDark: boolean) => set({ prefersDark }),
}))
