import React from 'react';
import { Moon, Sun, Monitor } from 'lucide-react';
import { useTheme, type Theme } from '../contexts/ThemeContext';

export const ThemeToggle: React.FC = () => {
  const { theme, setTheme } = useTheme();

  const themes: Array<{ value: Theme; label: string; icon: React.ReactNode }> = [
    { value: 'light', label: 'Light', icon: <Sun size={18} /> },
    { value: 'dark', label: 'Dark', icon: <Moon size={18} /> },
    { value: 'system', label: 'System', icon: <Monitor size={18} /> },
  ];

  return (
    <div className="flex items-center gap-1 p-1 bg-gray-100 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      {themes.map(({ value, label, icon }) => (
        <button
          key={value}
          onClick={() => setTheme(value)}
          className={`p-2 rounded transition flex items-center gap-1 ${
            theme === value
              ? 'bg-white dark:bg-gray-700 text-blue-600 shadow-sm'
              : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
          }`}
          title={label}
        >
          {icon}
          <span className="text-xs font-medium hidden sm:inline">{label}</span>
        </button>
      ))}
    </div>
  );
};

export default ThemeToggle;
