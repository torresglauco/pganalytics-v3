import React from 'react';
import { Menu, Bell, User, LogOut } from 'lucide-react';
import { useUIStore } from '../../store/uiStore';

interface HeaderProps {
  onLogout: () => void;
}

export const Header: React.FC<HeaderProps> = ({ onLogout }) => {
  const toggleSidebar = useUIStore((state) => state.toggleSidebar);
  const notifications = useUIStore((state) => state.notifications);

  return (
    <header className="bg-white border-b border-pg-slate/10 sticky top-0 z-40 shadow-sm">
      <div className="flex items-center justify-between px-6 py-4">
        {/* Left side */}
        <div className="flex items-center gap-4">
          <button
            onClick={toggleSidebar}
            className="p-2 hover:bg-pg-slate/10 rounded-lg transition"
            title="Toggle sidebar"
          >
            <Menu className="w-5 h-5 text-pg-dark" />
          </button>

          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-pg-blue to-pg-cyan flex items-center justify-center">
              <span className="text-white font-bold text-sm">PG</span>
            </div>
            <div>
              <h1 className="font-bold text-pg-dark">pgAnalytics</h1>
              <p className="text-xs text-pg-slate">v3.3.0</p>
            </div>
          </div>
        </div>

        {/* Center - Search (placeholder for future) */}
        <div className="flex-1 max-w-md mx-8 hidden md:block">
          <input
            type="text"
            placeholder="Search databases, alerts, queries..."
            className="w-full px-4 py-2 border border-pg-slate/20 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pg-cyan"
          />
        </div>

        {/* Right side */}
        <div className="flex items-center gap-4">
          {/* Notifications */}
          <button className="relative p-2 hover:bg-pg-slate/10 rounded-lg transition">
            <Bell className="w-5 h-5 text-pg-dark" />
            {notifications.length > 0 && (
              <span className="absolute top-0 right-0 w-5 h-5 bg-pg-danger text-white text-xs rounded-full flex items-center justify-center font-bold">
                {notifications.length}
              </span>
            )}
          </button>

          {/* User Menu */}
          <div className="flex items-center gap-2 pl-4 border-l border-pg-slate/10">
            <div className="hidden sm:block text-right">
              <p className="text-sm font-medium text-pg-dark">Admin</p>
              <p className="text-xs text-pg-slate">Administrator</p>
            </div>
            <div className="w-8 h-8 rounded-full bg-pg-blue text-white flex items-center justify-center">
              <User className="w-4 h-4" />
            </div>
            <button
              onClick={onLogout}
              className="p-2 hover:bg-pg-slate/10 rounded-lg transition"
              title="Logout"
            >
              <LogOut className="w-4 h-4 text-pg-dark" />
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};
