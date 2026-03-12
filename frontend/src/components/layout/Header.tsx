import React, { useState } from 'react'
import clsx from 'clsx'
import { useAuthStore } from '../../stores/authStore'
import { useThemeStore } from '../../stores/themeStore'
import { useNotificationStore } from '../../stores/notificationStore'
import { Button } from '../ui/Button'

export const Header: React.FC = () => {
  const [searchOpen, setSearchOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)

  const { user, logout } = useAuthStore()
  const { theme, toggleTheme } = useThemeStore()
  const { unreadCount } = useNotificationStore()

  return (
    <header className="sticky top-0 z-40 border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-sm">
      <div className="flex items-center justify-between h-16 px-4 md:px-6">
        {/* Logo */}
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-primary-600 rounded-lg" />
          <span className="text-lg font-bold text-slate-900 dark:text-slate-100">
            pgAnalytics
          </span>
        </div>

        {/* Search Bar */}
        <div className="hidden md:flex flex-1 max-w-xs mx-4">
          <div className="relative w-full">
            <input
              type="text"
              placeholder="Search logs, alerts..."
              className="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-slate-50 dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 text-xs">
              /
            </span>
          </div>
        </div>

        {/* Right Section */}
        <div className="flex items-center gap-2 md:gap-4">
          {/* Search Icon (mobile) */}
          <button
            onClick={() => setSearchOpen(!searchOpen)}
            className="md:hidden p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            aria-label="Search"
          >
            🔍
          </button>

          {/* Notifications */}
          <button
            className="relative p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            aria-label="Notifications"
          >
            🔔
            {unreadCount > 0 && (
              <span className="absolute top-1 right-1 w-2 h-2 bg-error rounded-full" />
            )}
          </button>

          {/* Theme Toggle */}
          <button
            onClick={toggleTheme}
            className="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            aria-label="Toggle theme"
          >
            {theme === 'light' ? '🌙' : '☀️'}
          </button>

          {/* User Menu */}
          <div className="relative">
            <button
              onClick={() => setUserMenuOpen(!userMenuOpen)}
              className="flex items-center gap-2 p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            >
              <div className="w-8 h-8 bg-primary-600 rounded-full" />
              <span className="hidden sm:inline text-sm font-medium text-slate-900 dark:text-slate-100">
                {user?.name}
              </span>
            </button>

            {/* User Dropdown */}
            {userMenuOpen && (
              <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg p-2 z-50">
                <div className="px-3 py-2 border-b border-slate-200 dark:border-slate-700">
                  <p className="text-sm font-medium text-slate-900 dark:text-slate-100">
                    {user?.name}
                  </p>
                  <p className="text-xs text-slate-600 dark:text-slate-400">
                    {user?.email}
                  </p>
                </div>
                <button className="w-full text-left px-3 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded transition-colors">
                  Settings
                </button>
                <button
                  onClick={logout}
                  className="w-full text-left px-3 py-2 text-sm text-error hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
                >
                  Logout
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Mobile Search */}
      {searchOpen && (
        <div className="md:hidden border-t border-slate-200 dark:border-slate-700 p-4">
          <input
            type="text"
            placeholder="Search logs, alerts..."
            className="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-slate-50 dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
            autoFocus
          />
        </div>
      )}
    </header>
  )
}

Header.displayName = 'Header'
