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

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render table with users', () => {
    render(<UserManagementTable users={mockUsers} />)

    expect(screen.getByText('user1')).toBeInTheDocument()
    expect(screen.getByText('user2')).toBeInTheDocument()
  })

  it('should display user emails', () => {
    render(<UserManagementTable users={mockUsers} />)

    expect(screen.getByText('user1@example.com')).toBeInTheDocument()
    expect(screen.getByText('user2@example.com')).toBeInTheDocument()
  })

  it('should display user roles', () => {
    render(<UserManagementTable users={mockUsers} />)

    const roleElements = screen.getAllByText(/user|admin/)
    expect(roleElements.length).toBeGreaterThan(0)
  })

  it('should render empty message when no users', () => {
    render(<UserManagementTable users={[]} />)

    expect(screen.getByText(/no users/i)).toBeInTheDocument()
  })
})
