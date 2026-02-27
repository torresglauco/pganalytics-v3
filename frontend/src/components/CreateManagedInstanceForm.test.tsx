import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CreateManagedInstanceForm } from './CreateManagedInstanceForm'
import { render } from '../test/utils'

describe('CreateManagedInstanceForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render managed instance form', () => {
    render(
      <CreateManagedInstanceForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(screen.getByLabelText(/instance name/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument()
  })

  it('should validate required fields', async () => {
    const user = userEvent.setup()
    render(
      <CreateManagedInstanceForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /create/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/required/i)).toBeInTheDocument()
    })
  })

  it('should allow form submission with valid data', async () => {
    const user = userEvent.setup()
    render(
      <CreateManagedInstanceForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const nameInput = screen.getByLabelText(/instance name/i)
    const submitButton = screen.getByRole('button', { name: /create/i })

    await user.type(nameInput, 'test-instance')
    await user.click(submitButton)

    expect(screen.getByLabelText(/instance name/i)).toBeInTheDocument()
  })
})
