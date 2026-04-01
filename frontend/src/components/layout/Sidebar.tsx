import React, { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import clsx from 'clsx'

interface NavItem {
  icon: string
  label: string
  href: string
  section: 'shortcuts' | 'main' | 'admin'
}

const navItems: NavItem[] = [
  // Shortcuts
  { section: 'shortcuts', icon: '🏠', label: 'Home', href: '/' },
  { section: 'shortcuts', icon: '📋', label: 'Logs', href: '/logs' },
  { section: 'shortcuts', icon: '📈', label: 'Metrics', href: '/metrics' },
  { section: 'shortcuts', icon: '🚨', label: 'Alerts', href: '/alerts' },

  // Main
  { section: 'main', icon: '📁', label: 'Collectors', href: '/collectors' },
  { section: 'main', icon: '🔔', label: 'Channels', href: '/channels' },
  { section: 'main', icon: '🔍', label: 'Index Advisor', href: '/index-advisor' },
  { section: 'main', icon: '📊', label: 'Grafana', href: '/grafana' },

  // Admin
  { section: 'admin', icon: '👥', label: 'Users', href: '/users' },
  { section: 'admin', icon: '⚙', label: 'Settings', href: '/settings' },
]

export const Sidebar: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()

  const isActive = (href: string) => location.pathname === href

  const renderSection = (section: 'shortcuts' | 'main' | 'admin') => {
    const items = navItems.filter((item) => item.section === section)

    return (
      <div className="mb-6">
        <div
          className={clsx(
            'text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider',
            collapsed ? 'hidden' : 'px-4 mb-2',
          )}
        >
          {section === 'shortcuts' && 'Shortcuts'}
          {section === 'main' && 'Main'}
          {section === 'admin' && 'Admin'}
        </div>

        <div className="space-y-1">
          {items.map((item) => (
            <Link
              key={item.href}
              to={item.href}
              className={clsx(
                'flex items-center gap-3 px-4 py-2.5 rounded-lg transition-colors',
                isActive(item.href)
                  ? 'bg-primary-600 text-white'
                  : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700',
                collapsed && 'justify-center px-2',
              )}
              title={collapsed ? item.label : undefined}
            >
              <span className="text-lg">{item.icon}</span>
              {!collapsed && <span className="text-sm font-medium">{item.label}</span>}
            </Link>
          ))}
        </div>
      </div>
    )
  }

  return (
    <aside
      className={clsx(
        'flex flex-col h-screen border-r border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 transition-all duration-300',
        collapsed ? 'w-20' : 'w-64',
      )}
    >
      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {renderSection('shortcuts')}
        {renderSection('main')}
        {renderSection('admin')}
      </div>

      {/* Collapse Toggle */}
      <div className="border-t border-slate-200 dark:border-slate-700 p-4">
        <button
          onClick={() => setCollapsed(!collapsed)}
          className={clsx(
            'flex items-center justify-center w-full p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors',
          )}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? '→' : '←'}
        </button>
      </div>
    </aside>
  )
}

Sidebar.displayName = 'Sidebar'
