import React, { useState } from 'react'
import clsx from 'clsx'

interface Tab {
  label: string
  value: string
  disabled?: boolean
  icon?: React.ReactNode
}

interface TabsProps {
  tabs: Tab[]
  defaultValue?: string
  onChange?: (value: string) => void
  variant?: 'default' | 'underline'
  children?: (activeTab: string) => React.ReactNode
}

export const Tabs: React.FC<TabsProps> = ({
  tabs,
  defaultValue,
  onChange,
  variant = 'default',
  children,
}) => {
  const [activeTab, setActiveTab] = useState(defaultValue || tabs[0]?.value || '')

  const handleTabChange = (value: string) => {
    setActiveTab(value)
    onChange?.(value)
  }

  return (
    <div>
      <div
        className={clsx(
          'flex gap-2 border-b border-slate-200 dark:border-slate-700',
          variant === 'underline' && 'border-0',
        )}
        role="tablist"
      >
        {tabs.map((tab) => (
          <button
            key={tab.value}
            onClick={() => handleTabChange(tab.value)}
            disabled={tab.disabled}
            role="tab"
            aria-selected={activeTab === tab.value}
            className={clsx(
              'px-4 py-2.5 font-medium text-sm transition-colors relative whitespace-nowrap',
              activeTab === tab.value
                ? 'text-primary-600 dark:text-primary-400 border-b-2 border-primary-600 dark:border-primary-400'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200',
              tab.disabled && 'opacity-50 cursor-not-allowed',
            )}
          >
            <span className="flex items-center gap-2">
              {tab.icon && <span>{tab.icon}</span>}
              {tab.label}
            </span>
          </button>
        ))}
      </div>

      {children && (
        <div className="mt-4">
          {children(activeTab)}
        </div>
      )}
    </div>
  )
}

Tabs.displayName = 'Tabs'
