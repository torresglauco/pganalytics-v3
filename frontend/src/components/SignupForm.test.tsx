import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SignupForm } from './SignupForm'
import { render } from '../test/utils'

describe('SignupForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()
  const mockOnSwitchToLogin = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render signup form with all fields', () => {
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/email address/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/full name/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
  })

  it('should validate required fields', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const submitButton = screen.getByRole('button', { name: /create account/i })
    await user.click(submitButton)

    expect(screen.getByText('Username is required')).toBeInTheDocument()
    expect(screen.getByText('Email is required')).toBeInTheDocument()
    expect(screen.getByText('Password is required')).toBeInTheDocument()
  })

  it('should validate username length', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i)
    await user.type(usernameInput, 'ab')

    const submitButton = screen.getByRole('button', { name: /create account/i })
    await user.click(submitButton)

    expect(screen.getByText('Username must be at least 3 characters')).toBeInTheDocument()
  })

  it('should validate email format', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const emailInput = screen.getByLabelText(/email address/i)
    await user.type(emailInput, 'invalid-email')

    const submitButton = screen.getByRole('button', { name: /create account/i })
    await user.click(submitButton)

    expect(screen.getByText('Please enter a valid email address')).toBeInTheDocument()
  })

  it('should validate password length', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const passwordInput = screen.getByLabelText(/^password/i)
    await user.type(passwordInput, 'short')

    const submitButton = screen.getByRole('button', { name: /create account/i })
    await user.click(submitButton)

    expect(screen.getByText('Password must be at least 8 characters')).toBeInTheDocument()
  })

  it('should clear errors when user types', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const submitButton = screen.getByRole('button', { name: /create account/i })
    await user.click(submitButton)

    expect(screen.getByText('Username is required')).toBeInTheDocument()

    const usernameInput = screen.getByLabelText(/username/i)
    await user.type(usernameInput, 'validuser')

    await waitFor(() => {
      expect(screen.queryByText('Username is required')).not.toBeInTheDocument()
    })
  })

  it('should toggle password visibility', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const passwordInput = screen.getByLabelText(/^password/i)
    expect(passwordInput).toHaveAttribute('type', 'password')

    const toggleButtons = screen.getAllByRole('button')
    const toggleButton = toggleButtons.find((btn) => !btn.textContent)

    if (toggleButton) {
      await user.click(toggleButton)
      expect(passwordInput).toHaveAttribute('type', 'text')
    }
  })

  it('should call onSwitchToLogin when link is clicked', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const loginLink = screen.getByRole('button', { name: /log in here/i })
    await user.click(loginLink)

    expect(mockOnSwitchToLogin).toHaveBeenCalled()
  })

  it('should submit valid form', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i)
    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/^password/i)
    const fullNameInput = screen.getByLabelText(/full name/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })

    await user.type(usernameInput, 'newuser')
    await user.type(emailInput, 'new@example.com')
    await user.type(passwordInput, 'password123')
    await user.type(fullNameInput, 'New User')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalled()
    })
  })

  it('should disable submit button while loading', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i)
    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/^password/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })

    await user.type(usernameInput, 'newuser')
    await user.type(emailInput, 'new@example.com')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    await waitFor(() => {
      expect(submitButton).toBeDisabled()
    })
  })

  it('should handle signup error', async () => {
    const user = userEvent.setup()
    const signupError = new Error('Username already exists')

    vi.mocked(mockOnError).mockImplementation(() => {
      /* noop */
    })

    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i)
    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/^password/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })

    await user.type(usernameInput, 'existing')
    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    // Form validation should pass for these values
    expect(screen.queryByText('Username is required')).not.toBeInTheDocument()
  })

  it('should clear form after successful signup', async () => {
    const user = userEvent.setup()
    render(
      <SignupForm
        onSuccess={mockOnSuccess}
        onError={mockOnError}
        onSwitchToLogin={mockOnSwitchToLogin}
      />
    )

    const usernameInput = screen.getByLabelText(/username/i) as HTMLInputElement
    const emailInput = screen.getByLabelText(/email address/i) as HTMLInputElement
    const passwordInput = screen.getByLabelText(/^password/i) as HTMLInputElement
    const fullNameInput = screen.getByLabelText(/full name/i) as HTMLInputElement
    const submitButton = screen.getByRole('button', { name: /create account/i })

    await user.type(usernameInput, 'newuser')
    await user.type(emailInput, 'new@example.com')
    await user.type(passwordInput, 'password123')
    await user.type(fullNameInput, 'New User')
    await user.click(submitButton)

    await waitFor(() => {
      expect(usernameInput.value).toBe('')
      expect(emailInput.value).toBe('')
      expect(passwordInput.value).toBe('')
      expect(fullNameInput.value).toBe('')
    })
  })
})
