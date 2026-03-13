import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EscalationPolicyManager } from './EscalationPolicyManager'

// Mock fetch
global.fetch = vi.fn()

describe('EscalationPolicyManager', () => {
  const mockFetch = global.fetch as any

  const mockPolicies = [
    {
      id: 'policy-1',
      name: 'Critical Alert Policy',
      description: 'Policy for critical alerts',
      is_active: true,
      steps: [
        {
          step_number: 1,
          wait_minutes: 0,
          notification_channel: 'email',
          channel_config: { recipients: 'team@example.com' },
        },
        {
          step_number: 2,
          wait_minutes: 5,
          notification_channel: 'slack',
          channel_config: { channel_id: 'C123456' },
        },
      ],
    },
    {
      id: 'policy-2',
      name: 'Warning Alert Policy',
      description: 'Policy for warning alerts',
      is_active: true,
      steps: [
        {
          step_number: 1,
          wait_minutes: 0,
          notification_channel: 'email',
          channel_config: { recipients: 'team@example.com' },
        },
      ],
    },
    {
      id: 'policy-3',
      name: 'Inactive Policy',
      description: 'Inactive policy',
      is_active: false,
      steps: [],
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    localStorage.setItem('jwt_token', 'test-token')
  })

  it('renders the component', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    expect(screen.getByText('Escalation Policy')).toBeInTheDocument()
    expect(screen.getByText('Select Policy')).toBeInTheDocument()
  })

  it('loads policies from API on mount', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/escalation-policies',
        expect.objectContaining({
          headers: {
            'Authorization': 'Bearer test-token',
          },
        })
      )
    })
  })

  it('displays only active policies in dropdown', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    await waitFor(() => {
      expect(screen.getByText('Critical Alert Policy')).toBeInTheDocument()
      expect(screen.getByText('Warning Alert Policy')).toBeInTheDocument()
      expect(screen.queryByText('Inactive Policy')).not.toBeInTheDocument()
    })
  })

  it('allows policy selection', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    const user = userEvent.setup()
    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    await waitFor(() => {
      expect(screen.getByText('Critical Alert Policy')).toBeInTheDocument()
    })

    const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
    await user.selectOptions(selectElement, 'policy-1')

    expect(selectElement.value).toBe('policy-1')
  })

  it('displays policy details when policy is selected', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    const user = userEvent.setup()
    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    // Wait for options to appear
    await waitFor(() => {
      const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
      expect(selectElement).toBeInTheDocument()
    })

    const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
    await user.selectOptions(selectElement, 'policy-1')

    // After selection, policy details should appear
    await waitFor(() => {
      expect(screen.getByText('Policy for critical alerts')).toBeInTheDocument()
    })

    expect(screen.getByText(/Step 1: email after 0m/)).toBeInTheDocument()
    expect(screen.getByText(/Step 2: slack after 5m/)).toBeInTheDocument()
  })

  it('disables link button when no policy is selected', () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    const linkButton = screen.getByRole('button', { name: /Link Policy/i }) as HTMLButtonElement
    expect(linkButton.disabled).toBe(true)
  })

  it('enables link button when policy is selected', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    const user = userEvent.setup()
    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    await waitFor(() => {
      expect(screen.getByText('Critical Alert Policy')).toBeInTheDocument()
    })

    const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
    await user.selectOptions(selectElement, 'policy-1')

    const linkButton = screen.getByRole('button', { name: /Link Policy/i }) as HTMLButtonElement
    expect(linkButton.disabled).toBe(false)
  })

  it('calls API to link policy when button is clicked', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 'link-1' }),
    })

    const user = userEvent.setup()
    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    await waitFor(() => {
      expect(screen.getByText('Critical Alert Policy')).toBeInTheDocument()
    })

    const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
    await user.selectOptions(selectElement, 'policy-1')

    const linkButton = screen.getByRole('button', { name: /Link Policy/i })
    await user.click(linkButton)

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/alert-rules/rule-1/escalation-policies',
        expect.objectContaining({
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer test-token',
          },
          body: JSON.stringify({
            policy_id: 'policy-1',
          }),
        })
      )
    })
  })

  it('displays error message when policy link fails', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    mockFetch.mockResolvedValueOnce({
      ok: false,
    })

    const user = userEvent.setup()
    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    await waitFor(() => {
      expect(screen.getByText('Critical Alert Policy')).toBeInTheDocument()
    })

    const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
    await user.selectOptions(selectElement, 'policy-1')

    const linkButton = screen.getByRole('button', { name: /Link Policy/i })
    await user.click(linkButton)

    await waitFor(() => {
      expect(screen.getByText('Failed to link policy')).toBeInTheDocument()
    })
  })

  it('clears selection after successful policy link', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 'link-1' }),
    })

    const user = userEvent.setup()
    render(<EscalationPolicyManager alertRuleId="rule-1" />)

    await waitFor(() => {
      expect(screen.getByText('Critical Alert Policy')).toBeInTheDocument()
    })

    const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
    await user.selectOptions(selectElement, 'policy-1')

    const linkButton = screen.getByRole('button', { name: /Link Policy/i })
    await user.click(linkButton)

    await waitFor(() => {
      expect((selectElement as HTMLSelectElement).value).toBe('')
    })
  })

  it('calls onPolicyLinked callback after successful link', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPolicies,
    })

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 'link-1' }),
    })

    const onPolicyLinked = vi.fn()
    const user = userEvent.setup()

    render(
      <EscalationPolicyManager
        alertRuleId="rule-1"
        onPolicyLinked={onPolicyLinked}
      />
    )

    await waitFor(() => {
      expect(screen.getByText('Critical Alert Policy')).toBeInTheDocument()
    })

    const selectElement = screen.getByDisplayValue('Choose a policy...') as HTMLSelectElement
    await user.selectOptions(selectElement, 'policy-1')

    const linkButton = screen.getByRole('button', { name: /Link Policy/i })
    await user.click(linkButton)

    await waitFor(() => {
      expect(onPolicyLinked).toHaveBeenCalled()
    })
  })
})
