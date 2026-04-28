import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CollectorForm } from './CollectorForm'
import { apiClient } from '../services/api'
import { render } from '../test/utils'
import type { CollectorRegisterResponse } from '../types'

// Mock the API module
vi.mock('../services/api', () => ({
  apiClient: {
    testConnection: vi.fn(),
    registerCollector: vi.fn(),
  },
}))

describe('CollectorForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()
  const mockRegistrationSecret = 'test-secret-key'

  const mockSuccessResponse: CollectorRegisterResponse = {
    collector_id: 'new-collector-123',
    status: 'registered',
    token: 'collector-token-abc',
    created_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
    ;(apiClient.testConnection as ReturnType<typeof vi.fn>).mockResolvedValue(true)
    ;(apiClient.registerCollector as ReturnType<typeof vi.fn>).mockResolvedValue(mockSuccessResponse)
  })

  describe('Form rendering', () => {
    it('should render form with hostname field and submit button', () => {
      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Use placeholder since label doesn't have for attribute
      expect(screen.getByPlaceholderText(/prod-db-1/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /register collector/i })).toBeInTheDocument()
    })

    it('should render test connection button', () => {
      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      expect(screen.getByRole('button', { name: /test/i })).toBeInTheDocument()
    })

    it('should render optional fields', () => {
      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Verify all form elements exist via text content
      expect(screen.getByText('Environment')).toBeInTheDocument()
      expect(screen.getByText('Group')).toBeInTheDocument()
      expect(screen.getByText('Description')).toBeInTheDocument()
    })
  })

  describe('Form validation', () => {
    it('should show validation error when hostname is empty on submit', async () => {
      const user = userEvent.setup()
      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Submit without entering hostname
      await user.click(screen.getByRole('button', { name: /register collector/i }))

      // Should show validation error
      await waitFor(() => {
        expect(screen.getByText(/hostname is required/i)).toBeInTheDocument()
      })
    })

    it('should clear validation error when hostname is entered', async () => {
      const user = userEvent.setup()
      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Submit empty to trigger error
      await user.click(screen.getByRole('button', { name: /register collector/i }))
      await waitFor(() => {
        expect(screen.getByText(/hostname is required/i)).toBeInTheDocument()
      })

      // Type hostname using placeholder
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'localhost')

      // Error should be cleared
      await waitFor(() => {
        expect(screen.queryByText(/hostname is required/i)).not.toBeInTheDocument()
      })
    })
  })

  describe('Connection testing', () => {
    it('should test connection successfully', async () => {
      const user = userEvent.setup()
      ;(apiClient.testConnection as ReturnType<typeof vi.fn>).mockResolvedValue(true)

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Enter hostname
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'localhost')

      // Click test button
      await user.click(screen.getByRole('button', { name: /test/i }))

      // Should show success message
      await waitFor(() => {
        expect(screen.getByText(/connection successful/i)).toBeInTheDocument()
      })

      expect(apiClient.testConnection).toHaveBeenCalledWith('localhost')
    })

    it('should display error when connection test fails', async () => {
      const user = userEvent.setup()
      ;(apiClient.testConnection as ReturnType<typeof vi.fn>).mockResolvedValue(false)

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Enter hostname
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'unreachable-host')

      // Click test button
      await user.click(screen.getByRole('button', { name: /test/i }))

      // Should show failure message
      await waitFor(() => {
        expect(screen.getByText(/connection failed/i)).toBeInTheDocument()
      })
    })

    it('should handle connection test error gracefully', async () => {
      const user = userEvent.setup()
      ;(apiClient.testConnection as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Network error'))

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Enter hostname
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'error-host')

      // Click test button
      await user.click(screen.getByRole('button', { name: /test/i }))

      // Should show failure message
      await waitFor(() => {
        expect(screen.getByText(/connection failed/i)).toBeInTheDocument()
      })
    })
  })

  describe('Collector registration', () => {
    it('should submit form with valid data and call onSuccess', async () => {
      const user = userEvent.setup()

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Fill form
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'prod-db-1.example.com')
      await user.selectOptions(screen.getByRole('combobox', { name: '' }), 'production')
      await user.type(screen.getByPlaceholderText(/aws-rds/i), 'AWS-RDS')

      // Submit
      await user.click(screen.getByRole('button', { name: /register collector/i }))

      // Verify API call
      await waitFor(() => {
        expect(apiClient.registerCollector).toHaveBeenCalledWith(
          expect.objectContaining({
            hostname: 'prod-db-1.example.com',
            environment: 'production',
            group: 'AWS-RDS',
          }),
          mockRegistrationSecret
        )
      })

      // Verify onSuccess callback
      await waitFor(() => {
        expect(mockOnSuccess).toHaveBeenCalledWith(mockSuccessResponse)
      })
    })

    it('should handle registration API error and call onError', async () => {
      const user = userEvent.setup()
      const error = new Error('Registration failed')
      ;(apiClient.registerCollector as ReturnType<typeof vi.fn>).mockRejectedValue(error)

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Fill and submit
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'localhost')
      await user.click(screen.getByRole('button', { name: /register collector/i }))

      // Verify onError callback
      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith(error)
      })
    })
  })

  describe('Success state', () => {
    it('should display collector ID and token after successful registration', async () => {
      const user = userEvent.setup()

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Fill and submit
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'localhost')
      await user.click(screen.getByRole('button', { name: /register collector/i }))

      // Verify success screen shows
      await waitFor(() => {
        expect(screen.getByText(/collector registered successfully/i)).toBeInTheDocument()
      })

      // Verify collector ID is displayed (appears multiple times in success UI)
      const collectorIdElements = screen.getAllByText(/new-collector-123/i)
      expect(collectorIdElements.length).toBeGreaterThan(0)

      // Verify token is displayed
      const tokenElements = screen.getAllByText(/collector-token-abc/i)
      expect(tokenElements.length).toBeGreaterThan(0)
    })

    it('should show "Register Another Collector" button after success', async () => {
      const user = userEvent.setup()

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Fill and submit
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'localhost')
      await user.click(screen.getByRole('button', { name: /register collector/i }))

      // Wait for success
      await waitFor(() => {
        expect(screen.getByText(/collector registered successfully/i)).toBeInTheDocument()
      })

      // Verify reset button exists
      expect(screen.getByRole('button', { name: /register another collector/i })).toBeInTheDocument()
    })

    it('should reset form when "Register Another Collector" is clicked', async () => {
      const user = userEvent.setup()

      render(
        <CollectorForm
          registrationSecret={mockRegistrationSecret}
          onSuccess={mockOnSuccess}
          onError={mockOnError}
        />
      )

      // Fill and submit
      await user.type(screen.getByPlaceholderText(/prod-db-1/i), 'localhost')
      await user.click(screen.getByRole('button', { name: /register collector/i }))

      // Wait for success
      await waitFor(() => {
        expect(screen.getByText(/collector registered successfully/i)).toBeInTheDocument()
      })

      // Click reset button
      await user.click(screen.getByRole('button', { name: /register another collector/i }))

      // Form should be back to initial state
      await waitFor(() => {
        expect(screen.getByPlaceholderText(/prod-db-1/i)).toBeInTheDocument()
        expect(screen.getByPlaceholderText(/prod-db-1/i)).toHaveValue('')
      })
    })
  })
})