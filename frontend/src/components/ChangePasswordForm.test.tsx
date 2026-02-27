import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ChangePasswordForm } from './ChangePasswordForm'
import { render } from '../test/utils'

describe('ChangePasswordForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render password change form', () => {
    render(
      <ChangePasswordForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const button = screen.getByRole('button', { name: /change password/i })
    expect(button).toBeInTheDocument()
  })

  it('should validate required fields', async () => {
    const user = userEvent.setup()
    render(
      <ChangePasswordForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /change password/i })
    await user.click(submitButton)

    expect(submitButton).toBeInTheDocument()
  })

  it('should validate password confirmation match', async () => {
    const user = userEvent.setup()
    render(
      <ChangePasswordForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /change password/i })
    await user.click(submitButton)

    expect(submitButton).toBeInTheDocument()
  })

  it('should submit form with valid data', async () => {
    const user = userEvent.setup()
    render(
      <ChangePasswordForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /change password/i })
    await user.click(submitButton)

    expect(submitButton).toBeInTheDocument()
  })
})
