import React from 'react'
import clsx from 'clsx'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  message?: string
  fullScreen?: boolean
}

const sizeStyles = {
  sm: 'h-6 w-6',
  md: 'h-10 w-10',
  lg: 'h-16 w-16',
}

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'md',
  message,
  fullScreen = false,
}) => {
  const spinner = (
    <div className="flex flex-col items-center gap-3">
      <div
        className={clsx(
          'animate-spin rounded-full border-2 border-slate-300 border-t-primary-600',
          sizeStyles[size],
        )}
        role="status"
        aria-label="Loading"
      />
      {message && (
        <p className="text-sm text-slate-600 dark:text-slate-400">{message}</p>
      )}
    </div>
  )

  if (fullScreen) {
    return (
      <div className="fixed inset-0 flex items-center justify-center bg-white dark:bg-slate-900 z-modal">
        {spinner}
      </div>
    )
  }

  return spinner
}

LoadingSpinner.displayName = 'LoadingSpinner'
