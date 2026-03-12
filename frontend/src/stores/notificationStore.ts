import { create } from 'zustand'
import { Toast } from '../types/common'

interface NotificationState {
  toasts: Toast[]
  unreadCount: number

  // Actions
  addToast: (toast: Omit<Toast, 'id'>) => void
  removeToast: (id: string) => void
  setUnreadCount: (count: number) => void
  clearToasts: () => void
}

export const useNotificationStore = create<NotificationState>((set) => ({
  toasts: [],
  unreadCount: 0,

  addToast: (toast) => set((state) => ({
    toasts: [
      ...state.toasts,
      {
        ...toast,
        id: Date.now().toString(),
      },
    ],
  })),
  removeToast: (id) => set((state) => ({
    toasts: state.toasts.filter((t) => t.id !== id),
  })),
  setUnreadCount: (count) => set({ unreadCount: count }),
  clearToasts: () => set({ toasts: [] }),
}))
