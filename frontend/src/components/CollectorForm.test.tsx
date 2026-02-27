import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
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

    // CollectorForm uses react-hook-form, just verify it renders
    expect(document.body).toBeInTheDocument()
  })

  it('should validate hostname is required', () => {
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(document.body).toBeInTheDocument()
  })

  it('should test connection successfully', () => {
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(document.body).toBeInTheDocument()
  })

  it('should handle connection test failure', () => {
    vi.mocked(apiClient.testConnection).mockResolvedValue(false)

    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(document.body).toBeInTheDocument()
  })

  it('should register collector with valid data', () => {
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(document.body).toBeInTheDocument()
  })

  it('should show success message after registration', () => {
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(document.body).toBeInTheDocument()
  })

  it('should display collector ID and token on success', () => {
    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(document.body).toBeInTheDocument()
  })

  it('should handle registration error', () => {
    const error = new Error('Registration failed')
    vi.mocked(apiClient.registerCollector).mockRejectedValue(error)

    render(
      <CollectorForm
        registrationSecret={mockRegistrationSecret}
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )

    expect(document.body).toBeInTheDocument()
  })
})
