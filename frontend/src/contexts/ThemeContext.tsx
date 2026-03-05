import React, { createContext, useState, useEffect } from 'react';

/**
 * Theme context for dark/light mode support
 */
export type Theme = 'light' | 'dark' | 'system';

export interface ThemeContextType {
  theme: Theme;
  isDark: boolean;
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

export const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

interface ThemeProviderProps {
  children: React.ReactNode;
  defaultTheme?: Theme;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({
  children,
  defaultTheme = 'system',
}) => {
  const [theme, setThemeState] = useState<Theme>(defaultTheme);
  const [isDark, setIsDark] = useState(false);

  /**
   * Initialize theme from localStorage and system preference
   */
  useEffect(() => {
    // Check localStorage for saved theme
    const savedTheme = localStorage.getItem('theme') as Theme | null;
    if (savedTheme) {
      setThemeState(savedTheme);
    }

    // Setup system preference listener
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

    const updateTheme = () => {
      const activeTheme = savedTheme || defaultTheme;
      if (activeTheme === 'system') {
        setIsDark(mediaQuery.matches);
        document.documentElement.classList.toggle('dark', mediaQuery.matches);
      } else {
        const isDarkMode = activeTheme === 'dark';
        setIsDark(isDarkMode);
        document.documentElement.classList.toggle('dark', isDarkMode);
      }
    };

    updateTheme();
    mediaQuery.addEventListener('change', updateTheme);

    return () => mediaQuery.removeEventListener('change', updateTheme);
  }, [defaultTheme]);

  /**
   * Handle theme change
   */
  const handleSetTheme = (newTheme: Theme) => {
    setThemeState(newTheme);
    localStorage.setItem('theme', newTheme);

    // Update DOM
    if (newTheme === 'system') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      const isDarkMode = mediaQuery.matches;
      setIsDark(isDarkMode);
      document.documentElement.classList.toggle('dark', isDarkMode);
    } else {
      const isDarkMode = newTheme === 'dark';
      setIsDark(isDarkMode);
      document.documentElement.classList.toggle('dark', isDarkMode);
    }
  };

  /**
   * Toggle between light and dark
   */
  const toggleTheme = () => {
    if (theme === 'light') {
      handleSetTheme('dark');
    } else if (theme === 'dark') {
      handleSetTheme('light');
    } else {
      // If system, toggle based on current system preference
      handleSetTheme(isDark ? 'light' : 'dark');
    }
  };

  return (
    <ThemeContext.Provider
      value={{
        theme,
        isDark,
        setTheme: handleSetTheme,
        toggleTheme,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
};

/**
 * Custom hook for accessing theme context
 */
export const useTheme = (): ThemeContextType => {
  const context = React.useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
};

export default ThemeProvider;
