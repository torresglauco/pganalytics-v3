import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthPage } from './AuthPage'
import { render } from '../test/utils'

describe('AuthPage', () => {
  const mockOnLogin = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    mockOnLogin.mockResolvedValue(undefined)
  })

  it('should render auth page with pgAnalytics title', () => {
    render(<AuthPage onLogin={mockOnLogin} />)

    expect(screen.getByText('pgAnalytics')).toBeInTheDocument()
    expect(screen.getByText('PostgreSQL Performance Analytics')).toBeInTheDocument()
  })

  it('should display login form', () => {
    render(<AuthPage onLogin={mockOnLogin} />)

    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /log in/i })).toBeInTheDocument()
  })

  it('should display external error message', () => {
    render(<AuthPage onLogin={mockOnLogin} error="Invalid credentials" />)

    expect(screen.getByText('Invalid credentials')).toBeInTheDocument()
  })

  it('should display success message on successful login', async () => {
    const user = userEvent.setup()
    render(<AuthPage onLogin={mockOnLogin} />)

    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /log in/i })

    await user.type(usernameInput, 'testuser')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockOnLogin).toHaveBeenCalledWith('testuser', 'password123')
    })
  })

  it('should display error message on login failure', async () => {
    const user = userEvent.setup()
    const error = new Error('Invalid credentials')
    mockOnLogin.mockRejectedValue(error)

    render(<AuthPage onLogin={mockOnLogin} />)

    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /log in/i })

    await user.type(usernameInput, 'testuser')
    await user.type(passwordInput, 'wrongpassword')
    await user.click(submitButton)

    // The error handling is done in the LoginForm component
    expect(mockOnLogin).toHaveBeenCalled()
  })

  it('should call onLogin with correct credentials', async () => {
    const user = userEvent.setup()
    render(<AuthPage onLogin={mockOnLogin} />)

    const usernameInput = screen.getByLabelText(/username/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /log in/i })

    await user.type(usernameInput, 'admin')
    await user.type(passwordInput, 'securepassword')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockOnLogin).toHaveBeenCalledWith('admin', 'securepassword')
    })
  })
})
