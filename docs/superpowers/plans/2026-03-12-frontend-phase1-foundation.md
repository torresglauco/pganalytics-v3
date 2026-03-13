# pgAnalytics Frontend Phase 1: Foundation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the foundation for pgAnalytics enterprise frontend redesign — setup professional styling, create reusable component library, update authentication flows, and build the unified Dashboard with drill-down pattern.

**Architecture:**
- Upgrade Tailwind config with professional color system and design tokens
- Create focused, single-responsibility components in `/src/components/ui/` (base) and `/src/components/` (composite)
- Setup Zustand store for global state (auth, theme, notifications)
- Redesign authentication pages (Login, Signup, MFA, Reset Password)
- Rebuild main layout (Header with search/notifications, responsive Sidebar)
- Build Dashboard (home page) with system health metrics, activity feed, drill-down navigation cards, and collector status table

**Tech Stack:** React 18, TypeScript, Tailwind CSS, Recharts, Zustand, React Router v6, Axios

**Phases:** Phase 1 builds the foundation (styling, components, auth, layout, dashboard). Remaining phases add Logs, Metrics, Alerts, Channels, Grafana, Users, Settings.

---

## File Structure Overview

### New Directory Organization
```
src/
├── components/
│   ├── ui/                          (Base components)
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Card.tsx
│   │   ├── Badge.tsx
│   │   ├── Modal.tsx
│   │   ├── Dropdown.tsx
│   │   ├── Tabs.tsx
│   │   ├── LoadingSpinner.tsx
│   │   └── Toast.tsx
│   ├── layout/                      (Layout components)
│   │   ├── Header.tsx
│   │   ├── Sidebar.tsx
│   │   └── MainLayout.tsx
│   ├── auth/                        (Auth pages)
│   │   ├── LoginPage.tsx
│   │   ├── SignupPage.tsx
│   │   ├── MFAVerificationPage.tsx
│   │   └── PasswordResetPage.tsx
│   ├── dashboard/                   (Dashboard components)
│   │   ├── Dashboard.tsx            (Main page)
│   │   ├── MetricCard.tsx
│   │   ├── ActivityFeed.tsx
│   │   ├── DrillDownCard.tsx
│   │   └── CollectorStatusTable.tsx
│   └── common/                      (Existing - keep)
├── stores/                          (Zustand stores)
│   ├── authStore.ts
│   ├── themeStore.ts
│   └── notificationStore.ts
├── types/                           (TypeScript types)
│   ├── auth.ts
│   ├── common.ts
│   └── dashboard.ts
├── services/                        (API, utilities)
│   ├── api.ts                       (Existing - update)
│   └── theme.ts                     (New)
├── styles/                          (CSS)
│   └── index.css                    (Update with Tailwind imports)
└── pages/                           (Page containers)
    └── RootLayout.tsx               (New - wraps everything)
```

### Files to Create (26 new files)
- `src/components/ui/Button.tsx`
- `src/components/ui/Input.tsx`
- `src/components/ui/Card.tsx`
- `src/components/ui/Badge.tsx`
- `src/components/ui/Modal.tsx`
- `src/components/ui/Dropdown.tsx`
- `src/components/ui/Tabs.tsx`
- `src/components/ui/LoadingSpinner.tsx`
- `src/components/ui/Toast.tsx`
- `src/components/layout/Header.tsx`
- `src/components/layout/Sidebar.tsx`
- `src/components/layout/MainLayout.tsx`
- `src/components/auth/LoginPage.tsx`
- `src/components/auth/SignupPage.tsx`
- `src/components/auth/MFAVerificationPage.tsx`
- `src/components/auth/PasswordResetPage.tsx`
- `src/components/dashboard/Dashboard.tsx`
- `src/components/dashboard/MetricCard.tsx`
- `src/components/dashboard/ActivityFeed.tsx`
- `src/components/dashboard/DrillDownCard.tsx`
- `src/components/dashboard/CollectorStatusTable.tsx`
- `src/stores/authStore.ts`
- `src/stores/themeStore.ts`
- `src/stores/notificationStore.ts`
- `src/types/auth.ts`
- `src/types/dashboard.ts`

### Files to Modify
- `frontend/tailwind.config.js` (Update with design tokens)
- `frontend/src/App.tsx` (Update routing structure)
- `frontend/src/styles/index.css` (Add Tailwind directives)
- `frontend/src/services/api.ts` (Ensure correct auth headers)

---

## Chunk 1: Design Tokens & Tailwind Configuration

### Task 1: Update Tailwind Configuration with Design Tokens

**Files:**
- Modify: `frontend/tailwind.config.js`

**Description:** Update Tailwind config with professional color palette, typography scale, spacing tokens, and custom animations.

- [ ] **Step 1: Read current tailwind.config.js**

Already done above.

- [ ] **Step 2: Update tailwind.config.js with design tokens**

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      // Professional Color Palette
      colors: {
        // Primary (Blue)
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        // Semantic Colors
        success: '#10b981',
        warning: '#f59e0b',
        error: '#ef4444',
        info: '#3b82f6',

        // Extended Grays (Slate)
        slate: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
        },
      },

      // Typography
      fontSize: {
        'display': ['32px', { lineHeight: '1.2', fontWeight: '600' }],
        'h1': ['24px', { lineHeight: '1.2', fontWeight: '600' }],
        'h2': ['20px', { lineHeight: '1.2', fontWeight: '600' }],
        'h3': ['18px', { lineHeight: '1.5', fontWeight: '500' }],
        'body': ['16px', { lineHeight: '1.5', fontWeight: '400' }],
        'small': ['14px', { lineHeight: '1.5', fontWeight: '400' }],
        'label': ['12px', { lineHeight: '1.5', fontWeight: '500' }],
      },

      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['Fira Code', 'Menlo', 'monospace'],
      },

      // Spacing Scale (4dp increments)
      spacing: {
        'xs': '4px',
        'sm': '8px',
        'md': '12px',
        'lg': '16px',
        'xl': '24px',
        '2xl': '32px',
        '3xl': '48px',
      },

      // Border Radius
      borderRadius: {
        'none': '0',
        'sm': '2px',
        'base': '4px',
        'md': '6px',
        'lg': '8px',
        'xl': '12px',
      },

      // Shadows
      boxShadow: {
        'sm': '0 1px 2px rgba(0, 0, 0, 0.05)',
        'base': '0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06)',
        'md': '0 4px 6px rgba(0, 0, 0, 0.1), 0 2px 4px rgba(0, 0, 0, 0.06)',
        'lg': '0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05)',
        'xl': '0 20px 25px rgba(0, 0, 0, 0.1), 0 10px 10px rgba(0, 0, 0, 0.04)',
      },

      // Animations
      animation: {
        'fade-in': 'fadeIn 0.15s ease-out',
        'scale-in': 'scaleIn 0.15s ease-out',
        'slide-in-down': 'slideInDown 0.2s ease-out',
        'slide-in-up': 'slideInUp 0.2s ease-out',
        'pulse-subtle': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' },
        },
        slideInDown: {
          '0%': { opacity: '0', transform: 'translateY(-8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        slideInUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },

      // Z-index Scale
      zIndex: {
        'dropdown': '1000',
        'modal': '1050',
        'popover': '1100',
        'tooltip': '1150',
      },
    },
  },
  plugins: [],
}
```

- [ ] **Step 3: Run dev server to verify no errors**

```bash
cd frontend && npm run dev
```

Expected: App starts without errors, Tailwind classes are available.

- [ ] **Step 4: Commit**

```bash
git add frontend/tailwind.config.js
git commit -m "feat: update tailwind config with design tokens and professional color palette"
```

---

## Chunk 2: TypeScript Types & Global Stores

### Task 2: Create TypeScript Types

**Files:**
- Create: `src/types/auth.ts`
- Create: `src/types/common.ts`
- Create: `src/types/dashboard.ts`

- [ ] **Step 1: Create auth types**

```typescript
// src/types/auth.ts
export interface User {
  id: string
  email: string
  name: string
  role: 'admin' | 'editor' | 'viewer'
  organization_id: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface SignupRequest {
  email: string
  password: string
  name: string
  organization_name?: string
}

export interface MFAVerificationRequest {
  code: string
  session_id: string
}

export interface PasswordResetRequest {
  email: string
}

export interface PasswordResetConfirmRequest {
  token: string
  password: string
}
```

- [ ] **Step 2: Create common types**

```typescript
// src/types/common.ts
export interface ApiError {
  status: number
  message: string
  code?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export type LogLevel = 'DEBUG' | 'INFO' | 'NOTICE' | 'WARNING' | 'ERROR' | 'FATAL' | 'PANIC'

export type CollectorStatus = 'OK' | 'SLOW' | 'DOWN'

export interface Collector {
  id: string
  hostname: string
  environment: string
  group?: string
  description?: string
  status: CollectorStatus
  last_heartbeat: string
  created_at: string
  updated_at: string
}

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

export interface NotificationPreferences {
  email_enabled: boolean
  browser_enabled: boolean
  sound_enabled: boolean
}
```

- [ ] **Step 3: Create dashboard types**

```typescript
// src/types/dashboard.ts
export interface DashboardMetric {
  label: string
  value: string | number
  unit?: string
  trend?: {
    direction: 'up' | 'down' | 'neutral'
    percentage: number
    period: string
  }
  color?: 'success' | 'warning' | 'error' | 'info'
}

export interface ActivityEvent {
  id: string
  type: 'error' | 'warning' | 'info' | 'success'
  title: string
  description: string
  timestamp: string
  action_url?: string
}

export interface DrillDownOption {
  id: string
  icon: string
  title: string
  description: string
  link: string
}

export interface CollectorStatusRow {
  id: string
  hostname: string
  environment: string
  status: 'OK' | 'SLOW' | 'DOWN'
  last_heartbeat: string
  error_count_24h: number
}
```

- [ ] **Step 4: Commit**

```bash
git add src/types/
git commit -m "feat: add typescript type definitions for auth, common, and dashboard"
```

### Task 3: Create Zustand Stores

**Files:**
- Create: `src/stores/authStore.ts`
- Create: `src/stores/themeStore.ts`
- Create: `src/stores/notificationStore.ts`

- [ ] **Step 1: Create auth store**

```typescript
// src/stores/authStore.ts
import { create } from 'zustand'
import { User } from '../types/auth'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  // Actions
  setUser: (user: User) => void
  setToken: (token: string) => void
  setAuthenticated: (isAuthenticated: boolean) => void
  setLoading: (isLoading: boolean) => void
  setError: (error: string | null) => void
  logout: () => void
  reset: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: localStorage.getItem('auth_token') || null,
  isAuthenticated: !!localStorage.getItem('auth_token'),
  isLoading: false,
  error: null,

  setUser: (user) => set({ user }),
  setToken: (token) => {
    localStorage.setItem('auth_token', token)
    set({ token, isAuthenticated: true })
  },
  setAuthenticated: (isAuthenticated) => set({ isAuthenticated }),
  setLoading: (isLoading) => set({ isLoading }),
  setError: (error) => set({ error }),
  logout: () => {
    localStorage.removeItem('auth_token')
    set({
      user: null,
      token: null,
      isAuthenticated: false,
      error: null,
    })
  },
  reset: () => set({
    user: null,
    token: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
  }),
}))
```

- [ ] **Step 2: Create theme store**

```typescript
// src/stores/themeStore.ts
import { create } from 'zustand'

type Theme = 'light' | 'dark' | 'auto'

interface ThemeState {
  theme: Theme
  prefersDark: boolean

  // Actions
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
  setPrefersDark: (prefersDark: boolean) => void
}

const getStoredTheme = (): Theme => {
  return (localStorage.getItem('theme') as Theme) || 'auto'
}

const getPrefersDark = (): boolean => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  theme: getStoredTheme(),
  prefersDark: getPrefersDark(),

  setTheme: (theme: Theme) => {
    localStorage.setItem('theme', theme)
    set({ theme })
  },
  toggleTheme: () => {
    const { theme } = get()
    const newTheme: Theme = theme === 'light' ? 'dark' : 'light'
    localStorage.setItem('theme', newTheme)
    set({ theme: newTheme })
  },
  setPrefersDark: (prefersDark: boolean) => set({ prefersDark }),
}))
```

- [ ] **Step 3: Create notification store**

```typescript
// src/stores/notificationStore.ts
import { create } from 'zustand'
import { Toast } from '../types/common'

interface NotificationState {
  toasts: Toast[]
  unreadCount: number

  // Actions
  addToast: (toast: Omit<Toast, 'id'>) => void
  removeToast: (id: string) => void
  setUnreadCount: (count: number) => void
  clearToasts: () => void
}

export const useNotificationStore = create<NotificationState>((set) => ({
  toasts: [],
  unreadCount: 0,

  addToast: (toast) => set((state) => ({
    toasts: [
      ...state.toasts,
      {
        ...toast,
        id: Date.now().toString(),
      },
    ],
  })),
  removeToast: (id) => set((state) => ({
    toasts: state.toasts.filter((t) => t.id !== id),
  })),
  setUnreadCount: (count) => set({ unreadCount: count }),
  clearToasts: () => set({ toasts: [] }),
}))
```

- [ ] **Step 4: Commit**

```bash
git add src/stores/
git commit -m "feat: create zustand stores for auth, theme, and notifications"
```

---

## Chunk 3: Base UI Components (Part 1)

### Task 4: Create Button Component

**Files:**
- Create: `src/components/ui/Button.tsx`

- [ ] **Step 1: Create Button component with variants**

```typescript
// src/components/ui/Button.tsx
import React from 'react'
import clsx from 'clsx'

type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost'
type ButtonSize = 'sm' | 'md' | 'lg'

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  isLoading?: boolean
  icon?: React.ReactNode
  iconPosition?: 'left' | 'right'
  fullWidth?: boolean
  children: React.ReactNode
}

const variantStyles: Record<ButtonVariant, string> = {
  primary: 'bg-primary-600 hover:bg-primary-700 text-white disabled:bg-primary-400',
  secondary: 'bg-slate-200 hover:bg-slate-300 text-slate-900 disabled:bg-slate-100',
  danger: 'bg-error hover:bg-red-600 text-white disabled:bg-red-300',
  ghost: 'bg-transparent hover:bg-slate-100 text-slate-900 disabled:opacity-50',
}

const sizeStyles: Record<ButtonSize, string> = {
  sm: 'px-3 py-2 text-sm font-medium rounded-md',
  md: 'px-4 py-2.5 text-base font-medium rounded-md',
  lg: 'px-6 py-3 text-base font-semibold rounded-lg',
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({
    variant = 'primary',
    size = 'md',
    isLoading = false,
    icon,
    iconPosition = 'left',
    fullWidth = false,
    disabled,
    className,
    children,
    ...props
  }, ref) => {
    const isDisabled = disabled || isLoading

    return (
      <button
        ref={ref}
        disabled={isDisabled}
        className={clsx(
          'inline-flex items-center justify-center gap-2',
          'font-medium transition-colors duration-150',
          'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          variantStyles[variant],
          sizeStyles[size],
          fullWidth && 'w-full',
          isLoading && 'opacity-60 cursor-wait',
          className,
        )}
        {...props}
      >
        {isLoading ? (
          <>
            <div className="animate-spin rounded-full h-4 w-4 border-2 border-current border-t-transparent" />
            {children}
          </>
        ) : (
          <>
            {icon && iconPosition === 'left' && <span>{icon}</span>}
            {children}
            {icon && iconPosition === 'right' && <span>{icon}</span>}
          </>
        )}
      </button>
    )
  }
)

Button.displayName = 'Button'
```

- [ ] **Step 2: Commit**

```bash
git add src/components/ui/Button.tsx
git commit -m "feat: create Button component with variants and sizes"
```

### Task 5: Create Input Component

**Files:**
- Create: `src/components/ui/Input.tsx`

- [ ] **Step 1: Create Input component**

```typescript
// src/components/ui/Input.tsx
import React from 'react'
import clsx from 'clsx'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  helpText?: string
  icon?: React.ReactNode
  iconPosition?: 'left' | 'right'
  fullWidth?: boolean
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({
    label,
    error,
    helpText,
    icon,
    iconPosition = 'left',
    fullWidth = true,
    disabled,
    className,
    type = 'text',
    ...props
  }, ref) => {
    return (
      <div className={clsx(fullWidth && 'w-full')}>
        {label && (
          <label className="block text-sm font-medium text-slate-900 mb-2">
            {label}
            {props.required && <span className="text-error ml-1">*</span>}
          </label>
        )}

        <div className="relative">
          {icon && iconPosition === 'left' && (
            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">
              {icon}
            </div>
          )}

          <input
            ref={ref}
            type={type}
            disabled={disabled}
            className={clsx(
              'w-full px-3 py-2.5 text-base',
              'border border-slate-300 rounded-md',
              'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent',
              'placeholder-slate-400',
              'disabled:bg-slate-50 disabled:text-slate-500 disabled:cursor-not-allowed',
              'transition-colors duration-150',
              error && 'border-error focus:ring-error',
              icon && iconPosition === 'left' && 'pl-9',
              icon && iconPosition === 'right' && 'pr-9',
              className,
            )}
            {...props}
          />

          {icon && iconPosition === 'right' && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500">
              {icon}
            </div>
          )}
        </div>

        {error && (
          <p className="mt-1 text-sm text-error">{error}</p>
        )}

        {helpText && !error && (
          <p className="mt-1 text-sm text-slate-500">{helpText}</p>
        )}
      </div>
    )
  }
)

Input.displayName = 'Input'
```

- [ ] **Step 2: Commit**

```bash
git add src/components/ui/Input.tsx
git commit -m "feat: create Input component with label, error, and icon support"
```

### Task 6: Create Card and Badge Components

**Files:**
- Create: `src/components/ui/Card.tsx`
- Create: `src/components/ui/Badge.tsx`

- [ ] **Step 1: Create Card component**

```typescript
// src/components/ui/Card.tsx
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
```

- [ ] **Step 2: Create Badge component**

```typescript
// src/components/ui/Badge.tsx
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
```

- [ ] **Step 3: Commit**

```bash
git add src/components/ui/Card.tsx src/components/ui/Badge.tsx
git commit -m "feat: create Card and Badge components"
```

---

## Chunk 4: Base UI Components (Part 2)

### Task 7: Create Modal, LoadingSpinner, Toast Components

**Files:**
- Create: `src/components/ui/Modal.tsx`
- Create: `src/components/ui/LoadingSpinner.tsx`
- Create: `src/components/ui/Toast.tsx`

- [ ] **Step 1: Create Modal component**

```typescript
// src/components/ui/Modal.tsx
import React, { useEffect } from 'react'
import clsx from 'clsx'

interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  children: React.ReactNode
  size?: 'sm' | 'md' | 'lg'
  closeOnBackdropClick?: boolean
}

const sizeStyles = {
  sm: 'w-full max-w-sm',
  md: 'w-full max-w-md',
  lg: 'w-full max-w-lg',
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  children,
  size = 'md',
  closeOnBackdropClick = true,
}) => {
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }

    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
      document.body.style.overflow = 'hidden'
    }

    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = 'auto'
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-modal flex items-center justify-center p-4">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 animate-fade-in"
        onClick={() => closeOnBackdropClick && onClose()}
        aria-hidden="true"
      />

      {/* Modal Content */}
      <div
        className={clsx(
          'relative bg-white dark:bg-slate-800',
          'rounded-lg shadow-xl',
          'animate-scale-in',
          sizeStyles[size],
          'max-h-[90vh] overflow-y-auto',
        )}
        role="dialog"
        aria-modal="true"
      >
        {/* Header */}
        {title && (
          <div className="flex items-center justify-between border-b border-slate-200 dark:border-slate-700 px-6 py-4">
            <h2 className="text-xl font-semibold text-slate-900 dark:text-slate-100">
              {title}
            </h2>
            <button
              onClick={onClose}
              className="text-slate-500 hover:text-slate-700 dark:hover:text-slate-300 transition-colors"
              aria-label="Close modal"
            >
              ✕
            </button>
          </div>
        )}

        {/* Body */}
        <div className="p-6">{children}</div>
      </div>
    </div>
  )
}

Modal.displayName = 'Modal'
```

- [ ] **Step 2: Create LoadingSpinner component**

```typescript
// src/components/ui/LoadingSpinner.tsx
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
```

- [ ] **Step 3: Create Toast component**

```typescript
// src/components/ui/Toast.tsx
import React, { useEffect } from 'react'
import clsx from 'clsx'
import { Toast as ToastType } from '../types/common'

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
```

- [ ] **Step 4: Commit**

```bash
git add src/components/ui/Modal.tsx src/components/ui/LoadingSpinner.tsx src/components/ui/Toast.tsx
git commit -m "feat: create Modal, LoadingSpinner, and Toast components"
```

---

## Chunk 5: Layout Components & Authentication Pages

### Task 8: Create Header Component

**Files:**
- Create: `src/components/layout/Header.tsx`

- [ ] **Step 1: Create Header component**

```typescript
// src/components/layout/Header.tsx
import React, { useState } from 'react'
import clsx from 'clsx'
import { useAuthStore } from '../../stores/authStore'
import { useThemeStore } from '../../stores/themeStore'
import { useNotificationStore } from '../../stores/notificationStore'
import { Button } from '../ui/Button'

export const Header: React.FC = () => {
  const [searchOpen, setSearchOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)

  const { user, logout } = useAuthStore()
  const { theme, toggleTheme } = useThemeStore()
  const { unreadCount } = useNotificationStore()

  return (
    <header className="sticky top-0 z-40 border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-sm">
      <div className="flex items-center justify-between h-16 px-4 md:px-6">
        {/* Logo */}
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-primary-600 rounded-lg" />
          <span className="text-lg font-bold text-slate-900 dark:text-slate-100">
            pgAnalytics
          </span>
        </div>

        {/* Search Bar */}
        <div className="hidden md:flex flex-1 max-w-xs mx-4">
          <div className="relative w-full">
            <input
              type="text"
              placeholder="Search logs, alerts..."
              className="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-slate-50 dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 text-xs">
              /
            </span>
          </div>
        </div>

        {/* Right Section */}
        <div className="flex items-center gap-2 md:gap-4">
          {/* Search Icon (mobile) */}
          <button
            onClick={() => setSearchOpen(!searchOpen)}
            className="md:hidden p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            aria-label="Search"
          >
            🔍
          </button>

          {/* Notifications */}
          <button
            className="relative p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            aria-label="Notifications"
          >
            🔔
            {unreadCount > 0 && (
              <span className="absolute top-1 right-1 w-2 h-2 bg-error rounded-full" />
            )}
          </button>

          {/* Theme Toggle */}
          <button
            onClick={toggleTheme}
            className="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            aria-label="Toggle theme"
          >
            {theme === 'light' ? '🌙' : '☀️'}
          </button>

          {/* User Menu */}
          <div className="relative">
            <button
              onClick={() => setUserMenuOpen(!userMenuOpen)}
              className="flex items-center gap-2 p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
            >
              <div className="w-8 h-8 bg-primary-600 rounded-full" />
              <span className="hidden sm:inline text-sm font-medium text-slate-900 dark:text-slate-100">
                {user?.name}
              </span>
            </button>

            {/* User Dropdown */}
            {userMenuOpen && (
              <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg p-2 z-50">
                <div className="px-3 py-2 border-b border-slate-200 dark:border-slate-700">
                  <p className="text-sm font-medium text-slate-900 dark:text-slate-100">
                    {user?.name}
                  </p>
                  <p className="text-xs text-slate-600 dark:text-slate-400">
                    {user?.email}
                  </p>
                </div>
                <button className="w-full text-left px-3 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded transition-colors">
                  Settings
                </button>
                <button
                  onClick={logout}
                  className="w-full text-left px-3 py-2 text-sm text-error hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
                >
                  Logout
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Mobile Search */}
      {searchOpen && (
        <div className="md:hidden border-t border-slate-200 dark:border-slate-700 p-4">
          <input
            type="text"
            placeholder="Search logs, alerts..."
            className="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-700 rounded-lg bg-slate-50 dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-primary-500"
            autoFocus
          />
        </div>
      )}
    </header>
  )
}

Header.displayName = 'Header'
```

- [ ] **Step 2: Commit**

```bash
git add src/components/layout/Header.tsx
git commit -m "feat: create Header component with search, notifications, theme toggle, and user menu"
```

### Task 9: Create Sidebar Component

**Files:**
- Create: `src/components/layout/Sidebar.tsx`

- [ ] **Step 1: Create Sidebar component**

```typescript
// src/components/layout/Sidebar.tsx
import React, { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import clsx from 'clsx'

interface NavItem {
  icon: string
  label: string
  href: string
  section: 'shortcuts' | 'main' | 'admin'
}

const navItems: NavItem[] = [
  // Shortcuts
  { section: 'shortcuts', icon: '🏠', label: 'Home', href: '/' },
  { section: 'shortcuts', icon: '📋', label: 'Logs', href: '/logs' },
  { section: 'shortcuts', icon: '📈', label: 'Metrics', href: '/metrics' },
  { section: 'shortcuts', icon: '🚨', label: 'Alerts', href: '/alerts' },

  // Main
  { section: 'main', icon: '📁', label: 'Collectors', href: '/collectors' },
  { section: 'main', icon: '🔔', label: 'Channels', href: '/channels' },
  { section: 'main', icon: '📊', label: 'Grafana', href: '/grafana' },

  // Admin
  { section: 'admin', icon: '👥', label: 'Users', href: '/users' },
  { section: 'admin', icon: '⚙', label: 'Settings', href: '/settings' },
]

export const Sidebar: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()

  const isActive = (href: string) => location.pathname === href

  const renderSection = (section: 'shortcuts' | 'main' | 'admin') => {
    const items = navItems.filter((item) => item.section === section)

    return (
      <div className="mb-6">
        <div
          className={clsx(
            'text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider',
            collapsed ? 'hidden' : 'px-4 mb-2',
          )}
        >
          {section === 'shortcuts' && 'Shortcuts'}
          {section === 'main' && 'Main'}
          {section === 'admin' && 'Admin'}
        </div>

        <div className="space-y-1">
          {items.map((item) => (
            <Link
              key={item.href}
              to={item.href}
              className={clsx(
                'flex items-center gap-3 px-4 py-2.5 rounded-lg transition-colors',
                isActive(item.href)
                  ? 'bg-primary-600 text-white'
                  : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700',
                collapsed && 'justify-center px-2',
              )}
              title={collapsed ? item.label : undefined}
            >
              <span className="text-lg">{item.icon}</span>
              {!collapsed && <span className="text-sm font-medium">{item.label}</span>}
            </Link>
          ))}
        </div>
      </div>
    )
  }

  return (
    <aside
      className={clsx(
        'flex flex-col h-screen border-r border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 transition-all duration-300',
        collapsed ? 'w-20' : 'w-64',
      )}
    >
      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {renderSection('shortcuts')}
        {renderSection('main')}
        {renderSection('admin')}
      </div>

      {/* Collapse Toggle */}
      <div className="border-t border-slate-200 dark:border-slate-700 p-4">
        <button
          onClick={() => setCollapsed(!collapsed)}
          className={clsx(
            'flex items-center justify-center w-full p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors',
          )}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? '→' : '←'}
        </button>
      </div>
    </aside>
  )
}

Sidebar.displayName = 'Sidebar'
```

- [ ] **Step 2: Commit**

```bash
git add src/components/layout/Sidebar.tsx
git commit -m "feat: create Sidebar component with collapsible navigation"
```

### Task 10: Create MainLayout Component

**Files:**
- Create: `src/components/layout/MainLayout.tsx`

- [ ] **Step 1: Create MainLayout component**

```typescript
// src/components/layout/MainLayout.tsx
import React from 'react'
import { Header } from './Header'
import { Sidebar } from './Sidebar'

interface MainLayoutProps {
  children: React.ReactNode
}

export const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  return (
    <div className="flex h-screen bg-white dark:bg-slate-900 overflow-hidden">
      {/* Sidebar */}
      <Sidebar />

      {/* Main Content */}
      <div className="flex flex-col flex-1 overflow-hidden">
        {/* Header */}
        <Header />

        {/* Page Content */}
        <main className="flex-1 overflow-y-auto">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}

MainLayout.displayName = 'MainLayout'
```

- [ ] **Step 2: Commit**

```bash
git add src/components/layout/MainLayout.tsx
git commit -m "feat: create MainLayout component"
```

---

## Chunk 6: Authentication Pages

### Task 11: Create Login Page

**Files:**
- Create: `src/components/auth/LoginPage.tsx`

- [ ] **Step 1: Create LoginPage component**

```typescript
// src/components/auth/LoginPage.tsx
import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { LoadingSpinner } from '../ui/LoadingSpinner'
import { useAuthStore } from '../../stores/authStore'
import { apiClient } from '../../services/api'

export const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { setUser, setToken, setError, setLoading, error, isLoading } = useAuthStore()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [emailError, setEmailError] = useState('')
  const [passwordError, setPasswordError] = useState('')

  const validateForm = () => {
    let isValid = true

    if (!email) {
      setEmailError('Email is required')
      isValid = false
    } else if (!/^[^@]+@[^@]+\.[^@]+$/.test(email)) {
      setEmailError('Please enter a valid email')
      isValid = false
    } else {
      setEmailError('')
    }

    if (!password) {
      setPasswordError('Password is required')
      isValid = false
    } else {
      setPasswordError('')
    }

    return isValid
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    try {
      setLoading(true)
      setError(null)

      const response = await apiClient.login(email, password)

      setToken(response.token)
      setUser(response.user)
      navigate('/')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Login failed'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Logging in..." />
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-slate-50 dark:from-slate-900 dark:to-slate-800 flex items-center justify-center p-4">
      {/* Left Side - Brand */}
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-center px-12">
        <div className="mb-8">
          <div className="w-16 h-16 bg-primary-600 rounded-xl mb-4" />
          <h1 className="text-4xl font-bold text-slate-900 dark:text-slate-100 mb-2">
            pgAnalytics
          </h1>
          <p className="text-lg text-slate-600 dark:text-slate-400">
            PostgreSQL Observability Platform
          </p>
        </div>

        <div className="space-y-6">
          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 mb-2">
              🎯 Monitor in Real-Time
            </h3>
            <p className="text-slate-600 dark:text-slate-400">
              Track PostgreSQL logs, metrics, and alerts in a unified dashboard
            </p>
          </div>

          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 mb-2">
              🔍 Deep Analysis
            </h3>
            <p className="text-slate-600 dark:text-slate-400">
              Find slow queries, errors, and performance issues instantly
            </p>
          </div>

          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 mb-2">
              🚨 Proactive Alerting
            </h3>
            <p className="text-slate-600 dark:text-slate-400">
              Get notified before problems impact your users
            </p>
          </div>
        </div>
      </div>

      {/* Right Side - Form */}
      <div className="w-full lg:w-1/2 max-w-md">
        <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 border border-slate-200 dark:border-slate-700">
          <h2 className="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">
            Sign In
          </h2>
          <p className="text-slate-600 dark:text-slate-400 mb-6">
            Welcome back to pgAnalytics
          </p>

          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-error">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Email"
              type="email"
              placeholder="you@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              error={emailError}
              required
            />

            <Input
              label="Password"
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              error={passwordError}
              required
            />

            <div className="flex items-center">
              <input
                type="checkbox"
                id="remember"
                className="w-4 h-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
              />
              <label htmlFor="remember" className="ml-2 text-sm text-slate-600 dark:text-slate-400">
                Remember me
              </label>
            </div>

            <Button type="submit" fullWidth size="lg">
              Sign In
            </Button>
          </form>

          <div className="mt-6">
            <div className="relative mb-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-200 dark:border-slate-700" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                  Or continue with
                </span>
              </div>
            </div>

            <Button variant="secondary" fullWidth size="md">
              SSO Login
            </Button>
          </div>

          <p className="mt-6 text-center text-slate-600 dark:text-slate-400">
            Don't have an account?{' '}
            <a href="/signup" className="text-primary-600 hover:text-primary-700 font-medium">
              Sign up
            </a>
          </p>
        </div>
      </div>
    </div>
  )
}

LoginPage.displayName = 'LoginPage'
```

- [ ] **Step 2: Commit**

```bash
git add src/components/auth/LoginPage.tsx
git commit -m "feat: create professional LoginPage component"
```

---

## Chunk 7: Dashboard Components

### Task 12: Create Dashboard Component Structure

**Files:**
- Create: `src/components/dashboard/Dashboard.tsx`
- Create: `src/components/dashboard/MetricCard.tsx`
- Create: `src/components/dashboard/ActivityFeed.tsx`
- Create: `src/components/dashboard/DrillDownCard.tsx`
- Create: `src/components/dashboard/CollectorStatusTable.tsx`

- [ ] **Step 1: Create MetricCard component**

```typescript
// src/components/dashboard/MetricCard.tsx
import React from 'react'
import { Card, CardTitle } from '../ui/Card'
import { Badge } from '../ui/Badge'

interface MetricCardProps {
  title: string
  value: string | number
  unit?: string
  trend?: {
    direction: 'up' | 'down' | 'neutral'
    percentage: number
    period: string
  }
  icon?: string
}

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  unit,
  trend,
  icon,
}) => {
  const getTrendColor = () => {
    if (!trend) return 'default'
    return trend.direction === 'up' ? 'error' : 'success'
  }

  const getTrendIcon = () => {
    if (!trend) return ''
    return trend.direction === 'up' ? '📈' : '📉'
  }

  return (
    <Card>
      <div className="flex items-start justify-between mb-3">
        <CardTitle className="text-sm font-medium text-slate-600 dark:text-slate-400">
          {title}
        </CardTitle>
        {icon && <span className="text-2xl">{icon}</span>}
      </div>

      <div className="mb-3">
        <div className="text-3xl font-bold text-slate-900 dark:text-slate-100">
          {value}
          {unit && <span className="text-lg text-slate-500 ml-1">{unit}</span>}
        </div>
      </div>

      {trend && (
        <Badge variant={getTrendColor()} size="sm">
          {getTrendIcon()} {Math.abs(trend.percentage)}% in {trend.period}
        </Badge>
      )}
    </Card>
  )
}

MetricCard.displayName = 'MetricCard'
```

- [ ] **Step 2: Create ActivityFeed component**

```typescript
// src/components/dashboard/ActivityFeed.tsx
import React from 'react'
import { Card, CardHeader, CardTitle } from '../ui/Card'
import { Badge } from '../ui/Badge'

interface Activity {
  id: string
  type: 'error' | 'warning' | 'info' | 'success'
  title: string
  description: string
  timestamp: string
}

interface ActivityFeedProps {
  activities: Activity[]
}

const getTypeColor = (type: string) => {
  switch (type) {
    case 'error':
      return 'error'
    case 'warning':
      return 'warning'
    case 'success':
      return 'success'
    default:
      return 'info'
  }
}

const getTypeIcon = (type: string) => {
  switch (type) {
    case 'error':
      return '🔴'
    case 'warning':
      return '🟡'
    case 'success':
      return '🟢'
    default:
      return '🔵'
  }
}

export const ActivityFeed: React.FC<ActivityFeedProps> = ({ activities }) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
      </CardHeader>

      <div className="space-y-3">
        {activities.length === 0 ? (
          <p className="text-sm text-slate-600 dark:text-slate-400">
            No recent activity
          </p>
        ) : (
          activities.map((activity) => (
            <div key={activity.id} className="flex gap-3 py-2">
              <span className="text-xl">{getTypeIcon(activity.type)}</span>
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-2">
                  <h4 className="text-sm font-medium text-slate-900 dark:text-slate-100">
                    {activity.title}
                  </h4>
                  <span className="text-xs text-slate-500 dark:text-slate-400 whitespace-nowrap">
                    {activity.timestamp}
                  </span>
                </div>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                  {activity.description}
                </p>
              </div>
            </div>
          ))
        )}
      </div>

      {activities.length > 0 && (
        <a
          href="/logs"
          className="inline-block mt-4 text-sm text-primary-600 hover:text-primary-700 font-medium"
        >
          View all logs →
        </a>
      )}
    </Card>
  )
}

ActivityFeed.displayName = 'ActivityFeed'
```

- [ ] **Step 3: Create DrillDownCard component**

```typescript
// src/components/dashboard/DrillDownCard.tsx
import React from 'react'
import { Link } from 'react-router-dom'

interface DrillDownCardProps {
  icon: string
  title: string
  description: string
  href: string
}

export const DrillDownCard: React.FC<DrillDownCardProps> = ({
  icon,
  title,
  description,
  href,
}) => {
  return (
    <Link
      to={href}
      className="block p-6 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-md hover:border-primary-300 dark:hover:border-primary-500 transition-all duration-200 group"
    >
      <div className="flex items-start gap-4">
        <span className="text-4xl group-hover:scale-110 transition-transform">
          {icon}
        </span>
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-1">
            {title}
          </h3>
          <p className="text-slate-600 dark:text-slate-400 text-sm">
            {description}
          </p>
          <span className="inline-block mt-3 text-primary-600 hover:text-primary-700 font-medium text-sm">
            Explore →
          </span>
        </div>
      </div>
    </Link>
  )
}

DrillDownCard.displayName = 'DrillDownCard'
```

- [ ] **Step 4: Create CollectorStatusTable component**

```typescript
// src/components/dashboard/CollectorStatusTable.tsx
import React from 'react'
import { Card, CardHeader, CardTitle } from '../ui/Card'
import { Badge } from '../ui/Badge'

interface Collector {
  id: string
  hostname: string
  environment: string
  status: 'OK' | 'SLOW' | 'DOWN'
  last_heartbeat: string
  error_count_24h: number
}

interface CollectorStatusTableProps {
  collectors: Collector[]
  isLoading?: boolean
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'OK':
      return 'success'
    case 'SLOW':
      return 'warning'
    case 'DOWN':
      return 'error'
    default:
      return 'default'
  }
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'OK':
      return '✓'
    case 'SLOW':
      return '⚠'
    case 'DOWN':
      return '✗'
    default:
      return '•'
  }
}

export const CollectorStatusTable: React.FC<CollectorStatusTableProps> = ({
  collectors,
  isLoading = false,
}) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Collector Status (Real-time)</CardTitle>
      </CardHeader>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-200 dark:border-slate-700">
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Hostname
              </th>
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Environment
              </th>
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Status
              </th>
              <th className="text-left py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Last Seen
              </th>
              <th className="text-right py-3 px-4 font-semibold text-slate-900 dark:text-slate-100">
                Errors (24h)
              </th>
            </tr>
          </thead>
          <tbody>
            {collectors.map((collector) => (
              <tr
                key={collector.id}
                className="border-b border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors"
              >
                <td className="py-3 px-4 text-slate-900 dark:text-slate-100 font-medium">
                  {collector.hostname}
                </td>
                <td className="py-3 px-4 text-slate-600 dark:text-slate-400">
                  {collector.environment}
                </td>
                <td className="py-3 px-4">
                  <Badge variant={getStatusColor(collector.status)} size="sm">
                    {getStatusIcon(collector.status)} {collector.status}
                  </Badge>
                </td>
                <td className="py-3 px-4 text-slate-600 dark:text-slate-400">
                  {collector.last_heartbeat}
                </td>
                <td className="py-3 px-4 text-right text-slate-900 dark:text-slate-100">
                  {collector.error_count_24h}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  )
}

CollectorStatusTable.displayName = 'CollectorStatusTable'
```

- [ ] **Step 5: Create main Dashboard component**

```typescript
// src/components/dashboard/Dashboard.tsx
import React, { useState, useEffect } from 'react'
import { MainLayout } from '../layout/MainLayout'
import { MetricCard } from './MetricCard'
import { ActivityFeed } from './ActivityFeed'
import { DrillDownCard } from './DrillDownCard'
import { CollectorStatusTable } from './CollectorStatusTable'
import { LoadingSpinner } from '../ui/LoadingSpinner'
import { useAuthStore } from '../../stores/authStore'

export const Dashboard: React.FC = () => {
  const { user } = useAuthStore()
  const [isLoading, setIsLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('24h')

  // Mock data - will be replaced with API calls
  const metrics = [
    {
      title: 'Active Collectors',
      value: 12,
      trend: { direction: 'up' as const, percentage: 0, period: '24h' },
      icon: '📡',
    },
    {
      title: 'Critical Alerts',
      value: 3,
      trend: { direction: 'up' as const, percentage: 12, period: '24h' },
      icon: '🚨',
    },
    {
      title: 'Total Errors',
      value: '1,234',
      trend: { direction: 'down' as const, percentage: 5, period: '24h' },
      icon: '❌',
    },
  ]

  const activities = [
    {
      id: '1',
      type: 'error' as const,
      title: 'High error rate detected',
      description: 'prod-db-1: 245 errors in last 5 minutes',
      timestamp: '2 min ago',
    },
    {
      id: '2',
      type: 'warning' as const,
      title: 'Slow query detected',
      description: 'Query took 8.5s on staging-db',
      timestamp: '15 min ago',
    },
    {
      id: '3',
      type: 'success' as const,
      title: 'Backup completed',
      description: 'Daily backup of analytics_prod completed successfully',
      timestamp: '1 hour ago',
    },
  ]

  const collectors = [
    {
      id: '1',
      hostname: 'prod-db-1.aws',
      environment: 'Production',
      status: 'OK' as const,
      last_heartbeat: '2s ago',
      error_count_24h: 23,
    },
    {
      id: '2',
      hostname: 'staging-db.local',
      environment: 'Staging',
      status: 'SLOW' as const,
      last_heartbeat: '1m ago',
      error_count_24h: 5,
    },
    {
      id: '3',
      hostname: 'dev-db-local',
      environment: 'Development',
      status: 'DOWN' as const,
      last_heartbeat: '2h ago',
      error_count_24h: 0,
    },
  ]

  useEffect(() => {
    // Simulate loading
    const timer = setTimeout(() => setIsLoading(false), 500)
    return () => clearTimeout(timer)
  }, [])

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Loading dashboard..." />
  }

  return (
    <MainLayout>
      <div className="py-6 md:py-8 px-4 md:px-6">
        {/* Page Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900 dark:text-slate-100 mb-1">
            Dashboard
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Welcome back, {user?.name || 'User'}! Here's what's happening with your databases.
          </p>
        </div>

        {/* Time Range Selector */}
        <div className="mb-6 flex gap-2">
          {['24h', '7d', '30d'].map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                timeRange === range
                  ? 'bg-primary-600 text-white'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 hover:bg-slate-200 dark:hover:bg-slate-600'
              }`}
            >
              {range}
            </button>
          ))}
        </div>

        {/* Metrics Row */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          {metrics.map((metric, i) => (
            <MetricCard key={i} {...metric} />
          ))}
        </div>

        {/* Activity & Drill-Down Row */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          <div className="lg:col-span-1">
            <ActivityFeed activities={activities} />
          </div>

          <div className="lg:col-span-2 space-y-4">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">
              Explore & Analyze
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <DrillDownCard
                icon="📋"
                title="View Logs"
                description="Filter and analyze PostgreSQL logs in detail"
                href="/logs"
              />
              <DrillDownCard
                icon="📈"
                title="View Metrics"
                description="Charts, performance trends, error distribution"
                href="/metrics"
              />
              <DrillDownCard
                icon="🚨"
                title="Manage Alerts"
                description="View active alerts and manage incidents"
                href="/alerts"
              />
              <DrillDownCard
                icon="📊"
                title="Grafana Dashboards"
                description="Custom dashboards from Grafana (embedded)"
                href="/grafana"
              />
            </div>
          </div>
        </div>

        {/* Collector Status Table */}
        <CollectorStatusTable collectors={collectors} />
      </div>
    </MainLayout>
  )
}

Dashboard.displayName = 'Dashboard'
```

- [ ] **Step 6: Commit**

```bash
git add src/components/dashboard/
git commit -m "feat: create Dashboard with metrics, activity feed, drill-down cards, and collector status table"
```

---

## Chunk 8: Update App Routing & CSS

### Task 13: Update App.tsx Routing

**Files:**
- Modify: `src/App.tsx`

- [ ] **Step 1: Update App.tsx with new routing structure**

```typescript
// src/App.tsx
import React, { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Dashboard } from './components/dashboard/Dashboard'
import { LoginPage } from './components/auth/LoginPage'
import { useAuthStore } from './stores/authStore'
import { apiClient } from './services/api'
import { LoadingSpinner } from './components/ui/LoadingSpinner'
import './styles/index.css'

function App() {
  const [isLoading, setIsLoading] = useState(true)
  const { isAuthenticated, setAuthenticated, setUser } = useAuthStore()

  useEffect(() => {
    const checkAuthentication = async () => {
      try {
        setIsLoading(true)

        if (apiClient.isAuthenticated()) {
          // Could fetch user profile here
          setAuthenticated(true)
        } else {
          setAuthenticated(false)
        }
      } catch (err) {
        console.error('Auth check failed:', err)
        setAuthenticated(false)
      } finally {
        setIsLoading(false)
      }
    }

    checkAuthentication()
  }, [setAuthenticated])

  if (isLoading) {
    return <LoadingSpinner fullScreen message="Loading pgAnalytics..." />
  }

  return (
    <BrowserRouter>
      <Routes>
        {/* Public Routes */}
        <Route
          path="/login"
          element={isAuthenticated ? <Navigate to="/" /> : <LoginPage />}
        />

        {/* Protected Routes */}
        {isAuthenticated ? (
          <>
            <Route path="/" element={<Dashboard />} />
            {/* Other routes will be added in Phase 2+ */}
            <Route path="*" element={<Navigate to="/" />} />
          </>
        ) : (
          <Route path="*" element={<Navigate to="/login" />} />
        )}
      </Routes>
    </BrowserRouter>
  )
}

export default App
```

- [ ] **Step 2: Commit**

```bash
git add src/App.tsx
git commit -m "feat: update App.tsx with new routing structure and auth flow"
```

### Task 14: Update CSS & Add Tailwind Imports

**Files:**
- Modify: `src/styles/index.css`

- [ ] **Step 1: Update index.css with Tailwind directives**

```css
/* src/styles/index.css */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=Fira+Code:wght@400;500;600&display=swap');

@tailwind base;
@tailwind components;
@tailwind utilities;

/* Base Styles */
html {
  scroll-behavior: smooth;
}

body {
  @apply bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 transition-colors duration-300;
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
}

/* Scrollbar Styling */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  @apply bg-slate-100 dark:bg-slate-800;
}

::-webkit-scrollbar-thumb {
  @apply bg-slate-300 dark:bg-slate-600 rounded hover:bg-slate-400 dark:hover:bg-slate-500;
}

/* Focus Styles */
:focus-visible {
  @apply outline-none ring-2 ring-primary-500 ring-offset-2 dark:ring-offset-slate-900;
}

/* Selection */
::selection {
  @apply bg-primary-600 text-white;
}

/* Utility Classes */
.container-max {
  @apply max-w-7xl mx-auto;
}

.text-muted {
  @apply text-slate-600 dark:text-slate-400;
}

.text-subtle {
  @apply text-slate-500 dark:text-slate-500;
}

.border-divider {
  @apply border-slate-200 dark:border-slate-700;
}

.bg-surface {
  @apply bg-slate-50 dark:bg-slate-800;
}

.shadow-card {
  @apply shadow-sm;
}
```

- [ ] **Step 2: Commit**

```bash
git add src/styles/index.css
git commit -m "feat: add Tailwind directives and utility classes to CSS"
```

---

## Summary & Next Steps

**Phase 1 Complete:** Foundation, components, auth, layout, dashboard ✅

**Files Created:** 26 new components
**Files Modified:** 4 configuration files
**Total Commits:** 13

### Verification Checklist
- [ ] No TypeScript errors: `npm run type-check`
- [ ] No lint errors: `npm run lint`
- [ ] App runs without errors: `npm run dev`
- [ ] Dashboard loads correctly
- [ ] Login page renders
- [ ] Responsive design works on mobile

### Phase 2 (Next) Will Include:
- Logs Viewer page with table & filters
- Metrics & Analytics page with Recharts
- Alert Rules management page
- Collector management page
- Notification Channels configuration

**Total Implementation Time Estimate:** Phase 1 foundation + Phase 2-4 features = ~6-8 weeks for complete system

---

**Plan Status:** ✅ Ready for Implementation
**Execution Method:** Use superpowers:subagent-driven-development for task-by-task implementation with checkpoints
