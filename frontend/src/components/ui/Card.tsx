import React from 'react'
import clsx from 'clsx'

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  hover?: boolean
  children: React.ReactNode
}

export const Card: React.FC<CardProps> = ({ hover = false, className, children, ...props }) => {
  return (
    <div
      className={clsx(
        'bg-white dark:bg-slate-800',
        'border border-slate-200 dark:border-slate-700',
        'rounded-lg',
        'p-4 md:p-6',
        'shadow-sm',
        'transition-all duration-150',
        hover && 'hover:shadow-md hover:border-primary-300',
        className,
      )}
      {...props}
    >
      {children}
    </div>
  )
}

Card.displayName = 'Card'

interface CardHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
}

export const CardHeader: React.FC<CardHeaderProps> = ({ className, children, ...props }) => (
  <div className={clsx('border-b border-slate-200 dark:border-slate-700 pb-4 mb-4', className)} {...props}>
    {children}
  </div>
)

CardHeader.displayName = 'CardHeader'

interface CardTitleProps extends React.HTMLAttributes<HTMLHeadingElement> {
  children: React.ReactNode
}

export const CardTitle: React.FC<CardTitleProps> = ({ className, children, ...props }) => (
  <h2 className={clsx('text-lg font-semibold text-slate-900 dark:text-slate-100', className)} {...props}>
    {children}
  </h2>
)

CardTitle.displayName = 'CardTitle'
