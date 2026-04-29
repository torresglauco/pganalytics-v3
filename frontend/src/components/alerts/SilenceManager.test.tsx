import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SilenceManager } from './SilenceManager'

// Mock fetch
global.fetch = vi.fn()

describe('SilenceManager', () => {
  const mockFetch = global.fetch as any

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    localStorage.setItem('jwt_token', 'test-token')
  })

  it('renders the component', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    expect(screen.getByText('Silence Alert')).toBeInTheDocument()
    expect(screen.getByText('High Error Rate')).toBeInTheDocument()
  })

  it('renders duration dropdown with default options', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    const selectElement = screen.getByDisplayValue('1 hour') as HTMLSelectElement
    expect(selectElement).toBeInTheDocument()
    expect(screen.getByText('5 minutes')).toBeInTheDocument()
    expect(screen.getByText('15 minutes')).toBeInTheDocument()
    expect(screen.getByText('1 hour')).toBeInTheDocument()
  })

  it('allows duration change via dropdown', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    const selectElement = screen.getByDisplayValue('1 hour') as HTMLSelectElement
    fireEvent.change(selectElement, { target: { value: '300' } })

    expect(selectElement.value).toBe('300')
  })

  it('allows reason input', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    const user = userEvent.setup()
    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    const reasonInput = screen.getByPlaceholderText('e.g., Maintenance window, False positive')
    await user.type(reasonInput, 'Maintenance window')

    expect(reasonInput).toHaveValue('Maintenance window')
  })

  it('calls API to create silence when button is clicked', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 'silence-1' }),
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    const user = userEvent.setup()
    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    const reasonInput = screen.getByPlaceholderText('e.g., Maintenance window, False positive')
    await user.type(reasonInput, 'Test silence')

    const button = screen.getByRole('button', { name: /Silence for/ })
    await user.click(button)

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/alert-silences',
        expect.objectContaining({
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer test-token',
          },
        })
      )
    })
  })

  it('displays active silences from API', async () => {
    const mockSilences = [
      {
        id: 'silence-1',
        alert_rule_id: 'rule-1',
        reason: 'Maintenance window',
        expires_at: '2025-12-31T10:00:00Z',
      },
    ]

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockSilences,
    })

    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    await waitFor(() => {
      expect(screen.getByText('Active Silences')).toBeInTheDocument()
      expect(screen.getByText('Maintenance window')).toBeInTheDocument()
    })
  })

  it('calls API to deactivate silence when deactivate button is clicked', async () => {
    const mockSilences = [
      {
        id: 'silence-1',
        alert_rule_id: 'rule-1',
        reason: 'Maintenance window',
        expires_at: '2025-12-31T10:00:00Z',
      },
    ]

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockSilences,
    })

    const user = userEvent.setup()

    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    await waitFor(() => {
      expect(screen.getByText('Maintenance window')).toBeInTheDocument()
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    const deactivateButton = screen.getByRole('button', { name: 'Deactivate' })
    await user.click(deactivateButton)

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/alert-silences/silence-1',
        expect.objectContaining({
          method: 'DELETE',
        })
      )
    })
  })

  it('displays error message when silence creation fails', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    mockFetch.mockResolvedValueOnce({
      ok: false,
    })

    const user = userEvent.setup()
    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    const button = screen.getByRole('button', { name: /Silence for/ })
    await user.click(button)

    await waitFor(() => {
      expect(screen.getByText('Failed to create silence')).toBeInTheDocument()
    })
  })

  it('clears form after successful silence creation', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 'silence-1' }),
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    const user = userEvent.setup()
    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
      />
    )

    const reasonInput = screen.getByPlaceholderText('e.g., Maintenance window, False positive') as HTMLInputElement
    await user.type(reasonInput, 'Test silence')

    const button = screen.getByRole('button', { name: /Silence for/ })
    await user.click(button)

    await waitFor(() => {
      expect(reasonInput.value).toBe('')
    })
  })

  it('calls onSilenceCreated callback after successful silence creation', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 'silence-1' }),
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    })

    const onSilenceCreated = vi.fn()
    const user = userEvent.setup()

    render(
      <SilenceManager
        alertRuleId="rule-1"
        alertRuleName="High Error Rate"
        onSilenceCreated={onSilenceCreated}
      />
    )

    const button = screen.getByRole('button', { name: /Silence for/ })
    await user.click(button)

    await waitFor(() => {
      expect(onSilenceCreated).toHaveBeenCalled()
    })
  })
})
