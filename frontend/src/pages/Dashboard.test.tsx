import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { Dashboard } from './Dashboard'
import { render } from '../test/utils'

// Mock the API module
vi.mock('../services/api', () => ({
  apiClient: {
    getCurrentUser: vi.fn(),
  },
}))

// Mock child components to focus on Dashboard behavior
vi.mock('../components/CollectorForm', () => ({
  CollectorForm: () => <div data-testid="collector-form">Collector Form</div>,
}))

vi.mock('../components/CollectorList', () => ({
  CollectorList: () => <div data-testid="collector-list">Collector List</div>,
}))

vi.mock('../components/UserMenuDropdown', () => ({
  UserMenuDropdown: ({ onLogout }: { onLogout: () => void }) => (
    <button onClick={onLogout} data-testid="user-menu">
      User Menu
    </button>
  ),
}))

vi.mock('../components/ManagedInstancesTable', () => ({
  ManagedInstancesTable: () => <div data-testid="managed-instances-table">Managed Instances</div>,
}))

vi.mock('../components/RegistrationSecretsManager', () => ({
  RegistrationSecretsManager: () => <div data-testid="registration-secrets-manager">Secrets Manager</div>,
}))

describe('Dashboard', () => {
  const mockOnLogout = vi.fn()

  const regularUser: User = {
    id: 1,
    username: 'testuser',
    email: 'test@example.com',
    full_name: 'Test User',
    role: 'user',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  const adminUser: User = {
    id: 2,
    username: 'admin',
    email: 'admin@example.com',
    full_name: 'Admin User',
    role: 'admin',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Default to regular user
    ;(apiClient.getCurrentUser as ReturnType<typeof vi.fn>).mockReturnValue(regularUser)
  })

  describe('Rendering with user data', () => {
    it('should render dashboard header with title and version', () => {
      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.getByText('pgAnalytics Collector Manager')).toBeInTheDocument()
      expect(screen.getByText(/v3\.3\.0/)).toBeInTheDocument()
    })

    it('should render user menu dropdown', () => {
      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.getByTestId('user-menu')).toBeInTheDocument()
    })

    it('should render main content area', () => {
      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.getByText('Register Collector')).toBeInTheDocument()
      // CollectorForm is only shown when secret is entered - verify collector-list is present
      expect(screen.getByTestId('collector-list')).toBeInTheDocument()
    })

    it('should render registration secret input section', () => {
      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.getByText('Registration Secret Required')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Enter registration secret...')).toBeInTheDocument()
    })
  })

  describe('Admin features visibility', () => {
    it('should show only Manage Collectors tab for regular users', () => {
      (apiClient.getCurrentUser as ReturnType<typeof vi.fn>).mockReturnValue(regularUser)

      render(<Dashboard onLogout={mockOnLogout} />)

      // Regular users see Manage Collectors tab
      expect(screen.getByRole('tab', { name: /manage collectors/i })).toBeInTheDocument()

      // Admin-only tabs should NOT be visible
      expect(screen.queryByRole('tab', { name: /managed instances/i })).not.toBeInTheDocument()
      expect(screen.queryByRole('tab', { name: /registration secrets/i })).not.toBeInTheDocument()
    })

    it('should show all tabs for admin users', () => {
      (apiClient.getCurrentUser as ReturnType<typeof vi.fn>).mockReturnValue(adminUser)

      render(<Dashboard onLogout={mockOnLogout} />)

      // Admin sees all tabs
      expect(screen.getByRole('tab', { name: /manage collectors/i })).toBeInTheDocument()
      expect(screen.getByRole('tab', { name: /managed instances/i })).toBeInTheDocument()
      expect(screen.getByRole('tab', { name: /registration secrets/i })).toBeInTheDocument()
    })

    it('should not render ManagedInstancesTable for regular users', () => {
      (apiClient.getCurrentUser as ReturnType<typeof vi.fn>).mockReturnValue(regularUser)

      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.queryByTestId('managed-instances-table')).not.toBeInTheDocument()
    })

    it('should not render RegistrationSecretsManager for regular users', () => {
      (apiClient.getCurrentUser as ReturnType<typeof vi.fn>).mockReturnValue(regularUser)

      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.queryByTestId('registration-secrets-manager')).not.toBeInTheDocument()
    })
  })

  describe('Error handling', () => {
    it('should handle null user gracefully', () => {
      (apiClient.getCurrentUser as ReturnType<typeof vi.fn>).mockReturnValue(null)

      // Dashboard should still render without crashing
      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.getByText('pgAnalytics Collector Manager')).toBeInTheDocument()
    })

    it('should treat null user as non-admin', () => {
      (apiClient.getCurrentUser as ReturnType<typeof vi.fn>).mockReturnValue(null)

      render(<Dashboard onLogout={mockOnLogout} />)

      // Should not show admin tabs when user is null
      expect(screen.queryByRole('tab', { name: /managed instances/i })).not.toBeInTheDocument()
      expect(screen.queryByRole('tab', { name: /registration secrets/i })).not.toBeInTheDocument()
    })
  })

  describe('Registration secret functionality', () => {
    it('should show secret required message initially', () => {
      render(<Dashboard onLogout={mockOnLogout} />)

      expect(screen.getByText(/Secret is required to register collectors/i)).toBeInTheDocument()
    })

    it('should show collector form disabled when no secret', () => {
      render(<Dashboard onLogout={mockOnLogout} />)

      // When no secret is entered, show the "Secret Required" message
      expect(screen.getByText('Secret Required')).toBeInTheDocument()
    })
  })
})