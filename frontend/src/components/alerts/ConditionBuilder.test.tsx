import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ConditionBuilder } from './ConditionBuilder'
import type { AlertCondition } from '../../types/alerts'

describe('ConditionBuilder', () => {
  const mockAddCondition = vi.fn()
  const mockRemoveCondition = vi.fn()
  const mockUpdateCondition = vi.fn()

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('renders empty state when no conditions', () => {
    render(
      <ConditionBuilder
        conditions={[]}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    expect(screen.getByText('No conditions added yet')).toBeInTheDocument()
  })

  it('renders Add Condition button', () => {
    render(
      <ConditionBuilder
        conditions={[]}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    expect(screen.getByRole('button', { name: /Add Condition/i })).toBeInTheDocument()
  })

  it('calls onAddCondition when Add Condition button is clicked', () => {
    render(
      <ConditionBuilder
        conditions={[]}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    const addButton = screen.getByRole('button', { name: /Add Condition/i })
    fireEvent.click(addButton)

    expect(mockAddCondition).toHaveBeenCalled()
  })

  it('renders conditions when provided', () => {
    const conditions: AlertCondition[] = [
      {
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      },
    ]

    render(
      <ConditionBuilder
        conditions={conditions}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    expect(screen.getByDisplayValue('5')).toBeInTheDocument()
    expect(screen.getByDisplayValue('10')).toBeInTheDocument()
  })

  it('renders Remove button for each condition', () => {
    const conditions: AlertCondition[] = [
      {
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      },
    ]

    render(
      <ConditionBuilder
        conditions={conditions}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    expect(screen.getByText('Remove')).toBeInTheDocument()
  })

  it('calls onRemoveCondition when Remove button is clicked', () => {
    const conditions: AlertCondition[] = [
      {
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      },
    ]

    render(
      <ConditionBuilder
        conditions={conditions}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    const removeButton = screen.getByText('Remove')
    fireEvent.click(removeButton)

    expect(mockRemoveCondition).toHaveBeenCalledWith(0)
  })

  it('shows AND label for multiple conditions', () => {
    const conditions: AlertCondition[] = [
      {
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      },
      {
        id: '2',
        metricType: 'slow_query_count',
        operator: '>',
        threshold: 10,
        timeWindow: 15,
      },
    ]

    render(
      <ConditionBuilder
        conditions={conditions}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    const andLabels = screen.getAllByText('AND')
    expect(andLabels.length).toBeGreaterThan(0)
  })

  it('displays condition count', () => {
    const conditions: AlertCondition[] = [
      {
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      },
      {
        id: '2',
        metricType: 'slow_query_count',
        operator: '>',
        threshold: 10,
        timeWindow: 15,
      },
    ]

    render(
      <ConditionBuilder
        conditions={conditions}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    expect(screen.getByText('2 conditions')).toBeInTheDocument()
  })

  it('calls onUpdateCondition when condition field is changed', () => {
    const conditions: AlertCondition[] = [
      {
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      },
    ]

    render(
      <ConditionBuilder
        conditions={conditions}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
        onUpdateCondition={mockUpdateCondition}
      />
    )

    const thresholdInput = screen.getByDisplayValue('5') as HTMLInputElement
    fireEvent.change(thresholdInput, { target: { value: '10' } })

    expect(mockUpdateCondition).toHaveBeenCalledWith(0, expect.objectContaining({ threshold: 10 }))
  })

  it('renders condition preview for each condition', () => {
    const conditions: AlertCondition[] = [
      {
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      },
    ]

    render(
      <ConditionBuilder
        conditions={conditions}
        onAddCondition={mockAddCondition}
        onRemoveCondition={mockRemoveCondition}
      />
    )

    const errorCountTexts = screen.getAllByText(/Error Count/)
    expect(errorCountTexts.length).toBeGreaterThan(0)
  })
})
