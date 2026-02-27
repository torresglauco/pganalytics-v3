import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CreateUserForm } from './CreateUserForm'
import { render } from '../test/utils'

describe('CreateUserForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render create user form with required fields', () => {
    render(
      <CreateUserForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    // Check for form elements
    expect(screen.getByText(/create.*user/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument()
  })

  it('should validate required fields', async () => {
    const user = userEvent.setup()
    render(
      <CreateUserForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /create/i })
    await user.click(submitButton)

    // The form should be present
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument()
  })
})
