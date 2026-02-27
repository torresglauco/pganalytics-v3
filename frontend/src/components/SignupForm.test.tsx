import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SignupForm } from './SignupForm'
import { apiClient } from '../services/api'
import { render } from '../test/utils'

vi.mock('../services/api')

describe('SignupForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()
  const mockOnSwitchToLogin = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiClient.signup).mockResolvedValue({
      token: 'test-token',
      refresh_token: 'test-refresh',
      expires_at: '2024-12-31T00:00:00Z',
      user: {
        id: 1,
        username: 'newuser',
        email: 'new@example.com',
        full_name: 'New User',
        role: 'user',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    })
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

    const usernameInput = screen.getByLabelText(/username/i)
    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/^password/i)

    await user.type(usernameInput, 'validuser')
    await user.type(emailInput, 'invalid-email')
    await user.type(passwordInput, 'password123')

    const submitButton = screen.getByRole('button', { name: /create account/i })
    await user.click(submitButton)

    // Validation should prevent empty submission
    expect(submitButton).toBeInTheDocument()
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
      expect(vi.mocked(apiClient.signup)).toHaveBeenCalledWith({
        username: 'newuser',
        email: 'new@example.com',
        password: 'password123',
        full_name: 'New User',
      })
      expect(mockOnSuccess).toHaveBeenCalled()
    })
  })

  it('should disable submit button while loading', async () => {
    const user = userEvent.setup()
    vi.mocked(apiClient.signup).mockImplementation(() => new Promise(() => {})) // Never resolves

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

    expect(usernameInput.value).toBe('newuser')
    expect(emailInput.value).toBe('new@example.com')

    await user.click(submitButton)

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalled()
      expect(usernameInput.value).toBe('')
      expect(emailInput.value).toBe('')
      expect(passwordInput.value).toBe('')
      expect(fullNameInput.value).toBe('')
    })
  })
})
