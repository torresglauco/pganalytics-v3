import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LoginForm } from './LoginForm'
import { render } from '../test/utils'

describe('LoginForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()
  const mockOnSwitchToSignup = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render login form with all fields', () => {
    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToSignup={mockOnSwitchToSignup}
      />
    )

    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
    expect(screen.getByText(/don't have an account/i)).toBeInTheDocument()
  })

  it('should validate required fields', async () => {
    const user = userEvent.setup()
    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /log in/i })
    await user.click(submitButton)

    expect(screen.getByText('Username is required')).toBeInTheDocument()
    expect(screen.getByText('Password is required')).toBeInTheDocument()
  })

  it('should clear username error when user types', async () => {
    const user = userEvent.setup()
    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /log in/i })
    await user.click(submitButton)

    expect(screen.getByText('Username is required')).toBeInTheDocument()

    const usernameInput = screen.getByLabelText(/username/i)
    await user.type(usernameInput, 'testuser')

    await waitFor(() => {
      expect(screen.queryByText('Username is required')).not.toBeInTheDocument()
    })
  })

  it('should toggle password visibility', async () => {
    const user = userEvent.setup()
    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const passwordInput = screen.getByLabelText(/password/i)
    expect(passwordInput).toHaveAttribute('type', 'password')

    const toggleButton = screen.getByRole('button', { name: '' }).parentElement?.querySelector('button[type="button"]')
    if (toggleButton) {
      await user.click(toggleButton)
      expect(passwordInput).toHaveAttribute('type', 'text')
    }
  })

  it('should call onSwitchToSignup when link is clicked', async () => {
    const user = userEvent.setup()
    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToSignup={mockOnSwitchToSignup}
      />
    )

    const signupLink = screen.getByRole('button', { name: /sign up here/i })
    await user.click(signupLink)

    expect(mockOnSwitchToSignup).toHaveBeenCalled()
  })

  it('should call external login handler when provided', async () => {
    const user = userEvent.setup()
    const mockExternalLogin = vi.fn().mockResolvedValue(undefined)

    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onLogin={mockExternalLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /log in/i })

    await user.type(usernameInput, 'testuser')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockExternalLogin).toHaveBeenCalledWith('testuser', 'password123')
    })

    expect(mockOnSuccess).toHaveBeenCalledWith('Welcome back, testuser!')
  })

  it('should disable submit button while loading', async () => {
    const user = userEvent.setup()
    const mockExternalLogin = vi.fn(() => new Promise(() => {})) // Never resolves

    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onLogin={mockExternalLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /log in/i })

    await user.type(usernameInput, 'testuser')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    await waitFor(() => {
      expect(submitButton).toBeDisabled()
    })
  })

  it('should handle login error', async () => {
    const user = userEvent.setup()
    const loginError = new Error('Invalid credentials')
    const mockExternalLogin = vi.fn().mockRejectedValue(loginError)

    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onLogin={mockExternalLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /log in/i })

    await user.type(usernameInput, 'testuser')
    await user.type(passwordInput, 'wrongpassword')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockOnError).toHaveBeenCalledWith(loginError)
    })
  })

  it('should clear form after successful login', async () => {
    const user = userEvent.setup()
    const mockExternalLogin = vi.fn().mockResolvedValue(undefined)

    render(
      <LoginForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onLogin={mockExternalLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i) as HTMLInputElement
    const passwordInput = screen.getByLabelText(/password/i) as HTMLInputElement
    const submitButton = screen.getByRole('button', { name: /log in/i })

    await user.type(usernameInput, 'testuser')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    await waitFor(() => {
      expect(usernameInput.value).toBe('')
      expect(passwordInput.value).toBe('')
    })
  })
})
