import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { AlertRuleBuilder } from './AlertRuleBuilder'
import * as alertRuleBuilderHook from '../../hooks/useAlertRuleBuilder'

// Mock the hook
vi.mock('../../hooks/useAlertRuleBuilder', () => ({
  useAlertRuleBuilder: vi.fn(),
}))

describe('AlertRuleBuilder', () => {
  const mockHook = {
    name: '',
    description: '',
    conditions: [],
    errors: [],
    isLoading: false,
    isSaving: false,
    setName: vi.fn(),
    setDescription: vi.fn(),
    addCondition: vi.fn(),
    removeCondition: vi.fn(),
    updateCondition: vi.fn(),
    validateAndSave: vi.fn().mockResolvedValue(undefined),
    reset: vi.fn(),
    clearErrors: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHook as any)
  })

  it('renders the component', () => {
    render(<AlertRuleBuilder />)
    expect(screen.getByRole('heading', { name: /Create Alert Rule/i })).toBeInTheDocument()
  })

  it('renders input fields for name and description', () => {
    render(<AlertRuleBuilder />)

    const nameInput = screen.getByPlaceholderText('e.g., High Error Rate Alert')
    const descriptionInput = screen.getByPlaceholderText('Describe what this alert rule monitors')

    expect(nameInput).toBeInTheDocument()
    expect(descriptionInput).toBeInTheDocument()
  })

  it('renders condition builder section', () => {
    render(<AlertRuleBuilder />)
    expect(screen.getByText('Alert Conditions')).toBeInTheDocument()
  })

  it('renders save button', () => {
    render(<AlertRuleBuilder />)
    expect(screen.getByRole('button', { name: /Create Alert Rule/i })).toBeInTheDocument()
  })

  it('renders cancel button', () => {
    render(<AlertRuleBuilder />)
    expect(screen.getByRole('button', { name: /Cancel/i })).toBeInTheDocument()
  })

  it('displays validation error when name is empty', async () => {
    const mockHookWithError = {
      ...mockHook,
      errors: [{ field: 'name', message: 'Alert rule name is required' }],
    }
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHookWithError as any)

    render(<AlertRuleBuilder />)

    expect(screen.getByText('Alert rule name is required')).toBeInTheDocument()
  })

  it('displays validation error when conditions are empty', async () => {
    const mockHookWithError = {
      ...mockHook,
      errors: [{ field: 'conditions', message: 'At least one condition is required' }],
    }
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHookWithError as any)

    render(<AlertRuleBuilder />)

    expect(screen.getByText('At least one condition is required')).toBeInTheDocument()
  })

  it('calls validateAndSave when save button is clicked', async () => {
    const mockValidateAndSave = vi.fn().mockResolvedValue(undefined)
    const mockHookWithValidation = {
      ...mockHook,
      name: 'Test Alert',
      conditions: [{ id: '1', metricType: 'error_count', operator: '>', threshold: 5, timeWindow: 10 }],
      validateAndSave: mockValidateAndSave,
    }
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHookWithValidation as any)

    render(<AlertRuleBuilder />)

    const saveButton = screen.getByRole('button', { name: /Create Alert Rule/i })
    fireEvent.click(saveButton)

    await waitFor(() => {
      expect(mockValidateAndSave).toHaveBeenCalled()
    })
  })

  it('displays loading state while saving', () => {
    const mockHookWithLoading = {
      ...mockHook,
      isSaving: true,
    }
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHookWithLoading as any)

    render(<AlertRuleBuilder />)

    expect(screen.getByText('Saving alert rule...')).toBeInTheDocument()
  })

  it('calls onSuccess callback after successful save', async () => {
    const onSuccess = vi.fn()
    const mockValidateAndSave = vi.fn().mockResolvedValue(undefined)
    const mockHookWithValidation = {
      ...mockHook,
      name: 'Test Alert',
      conditions: [{ id: '1', metricType: 'error_count', operator: '>', threshold: 5, timeWindow: 10 }],
      errors: [],
      validateAndSave: mockValidateAndSave,
    }
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHookWithValidation as any)

    render(<AlertRuleBuilder onSuccess={onSuccess} />)

    const saveButton = screen.getByRole('button', { name: /Create Alert Rule/i })
    fireEvent.click(saveButton)

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled()
    })
  })

  it('calls onCancel callback when cancel button is clicked', () => {
    const onCancel = vi.fn()
    render(<AlertRuleBuilder onCancel={onCancel} />)

    const cancelButton = screen.getByRole('button', { name: /Cancel/i })
    fireEvent.click(cancelButton)

    expect(onCancel).toHaveBeenCalled()
  })

  it('disables inputs while saving', () => {
    const mockHookWithLoading = {
      ...mockHook,
      isSaving: true,
    }
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHookWithLoading as any)

    render(<AlertRuleBuilder />)

    const nameInput = screen.getByPlaceholderText('e.g., High Error Rate Alert') as HTMLInputElement
    const cancelButton = screen.getByRole('button', { name: /Cancel/i }) as HTMLButtonElement

    expect(nameInput.disabled).toBe(true)
    expect(cancelButton.disabled).toBe(true)
  })

  it('displays form error message', () => {
    const mockHookWithError = {
      ...mockHook,
      errors: [{ field: 'form', message: 'Failed to create alert rule' }],
    }
    vi.mocked(alertRuleBuilderHook.useAlertRuleBuilder).mockReturnValue(mockHookWithError as any)

    render(<AlertRuleBuilder />)

    expect(screen.getByText('Failed to create alert rule')).toBeInTheDocument()
  })
})
