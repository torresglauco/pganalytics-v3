import React, { createContext, useState, useCallback } from 'react';

/**
 * Toast notification types
 */
export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
  id: string;
  type: ToastType;
  message: string;
  title?: string;
  duration?: number; // ms, null = persistent
  action?: {
    label: string;
    onClick: () => void;
  };
}

export interface ToastContextType {
  toasts: Toast[];
  addToast: (toast: Omit<Toast, 'id'>) => string;
  removeToast: (id: string) => void;
  clearToasts: () => void;
  success: (message: string, title?: string) => string;
  error: (message: string, title?: string) => string;
  warning: (message: string, title?: string) => string;
  info: (message: string, title?: string) => string;
}

export const ToastContext = createContext<ToastContextType | undefined>(undefined);

interface ToastProviderProps {
  children: React.ReactNode;
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const [toasts, setToasts] = useState<Toast[]>([]);

  /**
   * Add a new toast notification
   */
  const addToast = useCallback(
    (toast: Omit<Toast, 'id'>): string => {
      const id = `${Date.now()}-${Math.random()}`;
      const newToast: Toast = {
        ...toast,
        id,
        duration: toast.duration ?? 5000, // Default 5 seconds
      };

      setToasts((prev) => [...prev, newToast]);

      // Auto-remove after duration
      if (newToast.duration && newToast.duration > 0) {
        setTimeout(() => removeToast(id), newToast.duration);
      }

      return id;
    },
    []
  );

  /**
   * Remove a toast notification
   */
  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  /**
   * Clear all toasts
   */
  const clearToasts = useCallback(() => {
    setToasts([]);
  }, []);

  /**
   * Convenience methods for different toast types
   */
  const success = useCallback(
    (message: string, title?: string) =>
      addToast({
        type: 'success',
        message,
        title: title || 'Success',
      }),
    [addToast]
  );

  const error = useCallback(
    (message: string, title?: string) =>
      addToast({
        type: 'error',
        message,
        title: title || 'Error',
        duration: null, // Persist until closed
      }),
    [addToast]
  );

  const warning = useCallback(
    (message: string, title?: string) =>
      addToast({
        type: 'warning',
        message,
        title: title || 'Warning',
      }),
    [addToast]
  );

  const info = useCallback(
    (message: string, title?: string) =>
      addToast({
        type: 'info',
        message,
        title: title || 'Info',
      }),
    [addToast]
  );

  return (
    <ToastContext.Provider
      value={{
        toasts,
        addToast,
        removeToast,
        clearToasts,
        success,
        error,
        warning,
        info,
      }}
    >
      {children}
    </ToastContext.Provider>
  );
};

/**
 * Custom hook for accessing toast context
 */
export const useToast = (): ToastContextType => {
  const context = React.useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
};

export default ToastProvider;
