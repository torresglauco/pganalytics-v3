import React from 'react'
import clsx from 'clsx'

type BadgeVariant = 'default' | 'success' | 'warning' | 'error' | 'info'
type BadgeSize = 'sm' | 'md'

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant
  size?: BadgeSize
  children: React.ReactNode
}

const variantStyles: Record<BadgeVariant, string> = {
  default: 'bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100',
  success: 'bg-emerald-100 dark:bg-emerald-900 text-emerald-800 dark:text-emerald-200',
  warning: 'bg-amber-100 dark:bg-amber-900 text-amber-800 dark:text-amber-200',
  error: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200',
  info: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
}

const sizeStyles: Record<BadgeSize, string> = {
  sm: 'px-2 py-1 text-xs font-medium rounded',
  md: 'px-3 py-1.5 text-sm font-medium rounded',
}

export const Badge: React.FC<BadgeProps> = ({
  variant = 'default',
  size = 'md',
  className,
  children,
  ...props
}) => {
  return (
    <span
      className={clsx(
        'inline-flex items-center',
        variantStyles[variant],
        sizeStyles[size],
        className,
      )}
      {...props}
    >
      {children}
    </span>
  )
}

Badge.displayName = 'Badge'
