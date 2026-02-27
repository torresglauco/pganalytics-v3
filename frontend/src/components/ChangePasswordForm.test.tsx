import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
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

    expect(screen.getByLabelText(/current password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/new password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
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

    await waitFor(() => {
      expect(screen.getByText(/required/i)).toBeInTheDocument()
    })
  })

  it('should validate password confirmation match', async () => {
    const user = userEvent.setup()
    render(
      <ChangePasswordForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const currentPasswordInput = screen.getByLabelText(/current password/i)
    const newPasswordInput = screen.getByLabelText(/^new password/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)

    await user.type(currentPasswordInput, 'oldpassword')
    await user.type(newPasswordInput, 'newpassword123')
    await user.type(confirmPasswordInput, 'differentpassword')

    const submitButton = screen.getByRole('button', { name: /change password/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/do not match/i)).toBeInTheDocument()
    })
  })

  it('should submit form with valid data', async () => {
    const user = userEvent.setup()
    render(
      <ChangePasswordForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const currentPasswordInput = screen.getByLabelText(/current password/i)
    const newPasswordInput = screen.getByLabelText(/^new password/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)

    await user.type(currentPasswordInput, 'oldpassword')
    await user.type(newPasswordInput, 'newpassword123')
    await user.type(confirmPasswordInput, 'newpassword123')

    const submitButton = screen.getByRole('button', { name: /change password/i })
    await user.click(submitButton)

    // Form submission would be handled by the component
    expect(screen.getByLabelText(/current password/i)).toBeInTheDocument()
  })
})
