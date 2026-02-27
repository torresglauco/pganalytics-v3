import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CollectorForm } from './CollectorForm'
import { apiClient } from '../services/api'
import { render } from '../test/utils'

vi.mock('../services/api')

describe('CollectorForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()
  const mockRegistrationSecret = 'test-secret-key'

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiClient.testConnection).mockResolvedValue(true)
    vi.mocked(apiClient.registerCollector).mockResolvedValue({
      collector_id: 'new-collector',
      status: 'registered',
      token: 'collector-token',
      created_at: '2024-01-01T00:00:00Z',
    })
  })

  it('should render form with hostname field', () => {
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(screen.getByLabelText(/hostname/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /test connection/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /register/i })).toBeInTheDocument()
  })

  it('should validate hostname is required', async () => {
    const user = userEvent.setup()
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const submitButton = screen.getByRole('button', { name: /register/i })
    await user.click(submitButton)

    expect(screen.getByText('Hostname is required')).toBeInTheDocument()
  })

  it('should test connection successfully', async () => {
    const user = userEvent.setup()
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const hostnameInput = screen.getByLabelText(/hostname/i)
    const testButton = screen.getByRole('button', { name: /test connection/i })

    await user.type(hostnameInput, 'localhost')
    await user.click(testButton)

    await waitFor(() => {
      expect(vi.mocked(apiClient.testConnection)).toHaveBeenCalledWith('localhost')
    })
  })

  it('should handle connection test failure', async () => {
    const user = userEvent.setup()
    vi.mocked(apiClient.testConnection).mockResolvedValue(false)

    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const hostnameInput = screen.getByLabelText(/hostname/i)
    const testButton = screen.getByRole('button', { name: /test connection/i })

    await user.type(hostnameInput, 'invalid-host')
    await user.click(testButton)

    await waitFor(() => {
      expect(vi.mocked(apiClient.testConnection)).toHaveBeenCalled()
    })
  })

  it('should register collector with valid data', async () => {
    const user = userEvent.setup()
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const hostnameInput = screen.getByLabelText(/hostname/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(hostnameInput, 'prod.example.com')
    await user.click(submitButton)

    await waitFor(() => {
      expect(vi.mocked(apiClient.registerCollector)).toHaveBeenCalledWith(
        expect.objectContaining({
          hostname: 'prod.example.com',
        }),
        mockRegistrationSecret
      )
    })

    expect(mockOnSuccess).toHaveBeenCalled()
  })

  it('should show success message after registration', async () => {
    const user = userEvent.setup()
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const hostnameInput = screen.getByLabelText(/hostname/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(hostnameInput, 'new-host')
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('Collector Registered Successfully!')).toBeInTheDocument()
    })
  })

  it('should display collector ID and token on success', async () => {
    const user = userEvent.setup()
    const mockResponse = {
      collector_id: 'collector-abc123',
      status: 'registered',
      token: 'super-secret-token',
      created_at: '2024-01-01T00:00:00Z',
    }

    vi.mocked(apiClient.registerCollector).mockResolvedValue(mockResponse)

    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const hostnameInput = screen.getByLabelText(/hostname/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(hostnameInput, 'new-host')
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText('collector-abc123')).toBeInTheDocument()
      expect(screen.getByText('super-secret-token')).toBeInTheDocument()
    })
  })

  it('should handle registration error', async () => {
    const user = userEvent.setup()
    const error = new Error('Registration failed')
    vi.mocked(apiClient.registerCollector).mockRejectedValue(error)

    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    const hostnameInput = screen.getByLabelText(/hostname/i)
    const submitButton = screen.getByRole('button', { name: /register/i })

    await user.type(hostnameInput, 'bad-host')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockOnError).toHaveBeenCalledWith(error)
    })
  })
})
