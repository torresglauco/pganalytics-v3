import React, { useEffect } from 'react'
import clsx from 'clsx'
import { Toast as ToastType } from '../../types/common'

interface ToastProps extends ToastType {
  onClose: (id: string) => void
}

const typeStyles = {
  success: 'bg-emerald-50 dark:bg-emerald-900 text-emerald-900 dark:text-emerald-100 border-emerald-200 dark:border-emerald-700',
  error: 'bg-red-50 dark:bg-red-900 text-red-900 dark:text-red-100 border-red-200 dark:border-red-700',
  warning: 'bg-amber-50 dark:bg-amber-900 text-amber-900 dark:text-amber-100 border-amber-200 dark:border-amber-700',
  info: 'bg-blue-50 dark:bg-blue-900 text-blue-900 dark:text-blue-100 border-blue-200 dark:border-blue-700',
}

export const Toast: React.FC<ToastProps> = ({
  id,
  type,
  title,
  message,
  duration = 4000,
  onClose,
}) => {
  useEffect(() => {
    if (duration) {
      const timer = setTimeout(() => onClose(id), duration)
      return () => clearTimeout(timer)
    }
  }, [id, duration, onClose])

  return (
    <div
      className={clsx(
        'border rounded-lg p-4 shadow-lg animate-slide-in-up',
        typeStyles[type],
      )}
      role="alert"
    >
      <div className="flex items-start gap-3">
        <div className="flex-1">
          <h3 className="font-semibold">{title}</h3>
          {message && <p className="text-sm mt-1 opacity-80">{message}</p>}
        </div>
        <button
          onClick={() => onClose(id)}
          className="text-current opacity-50 hover:opacity-100 transition-opacity"
          aria-label="Close notification"
        >
          ✕
        </button>
      </div>
    </div>
  )
}

Toast.displayName = 'Toast'
