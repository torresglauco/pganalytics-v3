import { create } from 'zustand';

interface UIStore {
  sidebarExpanded: boolean;
  theme: 'light' | 'dark';
  notificationCount: number;

  // Actions
  toggleSidebar: () => void;
  setSidebarExpanded: (expanded: boolean) => void;
  toggleTheme: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  setNotificationCount: (count: number) => void;
  incrementNotifications: () => void;
  decrementNotifications: () => void;
}

export const useUIStore = create<UIStore>((set) => ({
  sidebarExpanded: true,
  theme: 'light',
  notificationCount: 0,

  toggleSidebar: () => set((state) => ({
    sidebarExpanded: !state.sidebarExpanded,
  })),

  setSidebarExpanded: (expanded) => set({ sidebarExpanded: expanded }),

  toggleTheme: () => set((state) => ({
    theme: state.theme === 'light' ? 'dark' : 'light',
  })),

  setTheme: (theme) => set({ theme }),

  setNotificationCount: (count) => set({ notificationCount: count }),

  incrementNotifications: () => set((state) => ({
    notificationCount: state.notificationCount + 1,
  })),

  decrementNotifications: () => set((state) => ({
    notificationCount: Math.max(0, state.notificationCount - 1),
  })),
}));
