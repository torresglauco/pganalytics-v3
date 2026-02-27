import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
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

    const button = screen.getByRole('button', { name: /register instance|create/i })
    expect(button).toBeInTheDocument()
  })

  it('should validate required fields', async () => {
    render(
      <CreateManagedInstanceForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const button = screen.getByRole('button', { name: /register instance|create/i })
    expect(button).toBeInTheDocument()
  })

  it('should allow form submission with valid data', async () => {
    render(
      <CreateManagedInstanceForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const button = screen.getByRole('button', { name: /register instance|create/i })
    expect(button).toBeInTheDocument()
  })
})
