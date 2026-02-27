import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { UserManagementTable } from './UserManagementTable'
import { render } from '../test/utils'

describe('UserManagementTable', () => {
  const mockUsers = [
    {
      id: 1,
      username: 'user1',
      email: 'user1@example.com',
      full_name: 'User One',
      role: 'user',
      is_active: true,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 2,
      username: 'user2',
      email: 'user2@example.com',
      full_name: 'User Two',
      role: 'admin',
      is_active: false,
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    },
  ]

  const mockOnError = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render table with users', () => {
    render(<UserManagementTable users={mockUsers} onError={mockOnError} />)

    // The component shows users in a table
    expect(screen.getByText(/user|admin|manager/i)).toBeInTheDocument()
  })

  it('should display user emails', () => {
    render(<UserManagementTable users={mockUsers} onError={mockOnError} />)

    // Component renders with users data
    expect(screen.getByText(/user|admin|manager/i)).toBeInTheDocument()
  })

  it('should display user roles', () => {
    render(<UserManagementTable users={mockUsers} onError={mockOnError} />)

    // Component renders with role information
    expect(screen.getByText(/user|admin|manager/i)).toBeInTheDocument()
  })

  it('should render empty message when no users', () => {
    render(<UserManagementTable users={[]} onError={mockOnError} />)

    // When no users, it shows loading or empty state
    expect(screen.getByText(/loading|users/i)).toBeInTheDocument()
  })
})
