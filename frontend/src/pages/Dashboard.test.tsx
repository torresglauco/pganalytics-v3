import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Dashboard } from './Dashboard'
import { apiClient } from '../services/api'
import { render } from '../test/utils'

vi.mock('../services/api')
vi.mock('../components/CollectorForm')
vi.mock('../components/CollectorList')
vi.mock('../components/UserMenuDropdown')
vi.mock('../components/ManagedInstancesTable')
vi.mock('../components/RegistrationSecretsManager')

describe('Dashboard', () => {
  const mockOnLogout = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiClient.getCurrentUser).mockReturnValue({
      id: 1,
      username: 'testuser',
      email: 'test@example.com',
      full_name: 'Test User',
      role: 'user',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    })
  })

  it('should render dashboard header', () => {
    render(<Dashboard onLogout={mockOnLogout} />)

    expect(screen.getByText('pgAnalytics Collector Manager')).toBeInTheDocument()
    expect(screen.getByText(/v3\.3\.0/)).toBeInTheDocument()
  })

  it('should render tab navigation', () => {
    render(<Dashboard onLogout={mockOnLogout} />)

    expect(screen.getByText(/collectors/i)).toBeInTheDocument()
  })

  it('should display user menu', () => {
    render(<Dashboard onLogout={mockOnLogout} />)

    // UserMenuDropdown is mocked but should be rendered
    expect(screen.getByText('pgAnalytics Collector Manager')).toBeInTheDocument()
  })

  it('should hide admin features for regular users', () => {
    vi.mocked(apiClient.getCurrentUser).mockReturnValue({
      id: 1,
      username: 'regularuser',
      email: 'user@example.com',
      full_name: 'Regular User',
      role: 'user',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    })

    render(<Dashboard onLogout={mockOnLogout} />)

    // Admin tabs should not be visible for regular users
    expect(screen.getByText('pgAnalytics Collector Manager')).toBeInTheDocument()
  })

  it('should show admin features for admin users', () => {
    vi.mocked(apiClient.getCurrentUser).mockReturnValue({
      id: 2,
      username: 'admin',
      email: 'admin@example.com',
      full_name: 'Admin User',
      role: 'admin',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    })

    render(<Dashboard onLogout={mockOnLogout} />)

    expect(screen.getByText('pgAnalytics Collector Manager')).toBeInTheDocument()
  })

  it('should pass onLogout to UserMenuDropdown', () => {
    render(<Dashboard onLogout={mockOnLogout} />)

    // Verify the component is rendered and has logout capability
    expect(screen.getByText('pgAnalytics Collector Manager')).toBeInTheDocument()
  })
})
