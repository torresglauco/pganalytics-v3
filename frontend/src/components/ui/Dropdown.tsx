import React, { useState, useRef, useEffect } from 'react'
import clsx from 'clsx'

interface DropdownItem {
  label: string
  value: string
  icon?: React.ReactNode
  disabled?: boolean
  divider?: boolean
}

interface DropdownProps {
  items: DropdownItem[]
  onSelect: (value: string) => void
  trigger: React.ReactNode
  align?: 'left' | 'right'
  className?: string
}

export const Dropdown: React.FC<DropdownProps> = ({
  items,
  onSelect,
  trigger,
  align = 'left',
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen])

  const handleSelect = (value: string) => {
    onSelect(value)
    setIsOpen(false)
  }

  return (
    <div ref={ref} className={clsx('relative inline-block', className)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 transition-colors"
      >
        {trigger}
      </button>

      {isOpen && (
        <div
          className={clsx(
            'absolute top-full mt-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg py-2 z-dropdown min-w-48',
            align === 'right' ? 'right-0' : 'left-0',
          )}
        >
          {items.map((item, idx) => (
            <React.Fragment key={idx}>
              {item.divider ? (
                <div className="h-px bg-slate-200 dark:bg-slate-700 my-2" />
              ) : (
                <button
                  onClick={() => handleSelect(item.value)}
                  disabled={item.disabled}
                  className={clsx(
                    'w-full text-left px-4 py-2 text-sm flex items-center gap-2 transition-colors',
                    item.disabled
                      ? 'text-slate-400 cursor-not-allowed'
                      : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700',
                  )}
                >
                  {item.icon && <span>{item.icon}</span>}
                  {item.label}
                </button>
              )}
            </React.Fragment>
          ))}
        </div>
      )}
    </div>
  )
}

Dropdown.displayName = 'Dropdown'
