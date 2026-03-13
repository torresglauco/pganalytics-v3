import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useAlertRuleBuilder } from './useAlertRuleBuilder'
import * as apiService from '../services/api'

vi.mock('../services/api', () => ({
  apiClient: {
    validateAlertCondition: vi.fn(),
    createAlertRule: vi.fn(),
  },
}))

describe('useAlertRuleBuilder', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('initializes with default state', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    expect(result.current.name).toBe('')
    expect(result.current.description).toBe('')
    expect(result.current.conditions).toEqual([])
    expect(result.current.errors).toEqual([])
    expect(result.current.isLoading).toBe(false)
    expect(result.current.isSaving).toBe(false)
  })

  it('updates name when setName is called', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    act(() => {
      result.current.setName('Test Alert')
    })

    expect(result.current.name).toBe('Test Alert')
  })

  it('updates description when setDescription is called', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    act(() => {
      result.current.setDescription('Test Description')
    })

    expect(result.current.description).toBe('Test Description')
  })

  it('adds a condition when addCondition is called', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    const newCondition = {
      id: '1',
      metricType: 'error_count' as const,
      operator: '>' as const,
      threshold: 5,
      timeWindow: 10,
    }

    act(() => {
      result.current.addCondition(newCondition)
    })

    expect(result.current.conditions).toHaveLength(1)
    expect(result.current.conditions[0]).toEqual(newCondition)
  })

  it('removes a condition when removeCondition is called', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    const condition = {
      id: '1',
      metricType: 'error_count' as const,
      operator: '>' as const,
      threshold: 5,
      timeWindow: 10,
    }

    act(() => {
      result.current.addCondition(condition)
    })

    expect(result.current.conditions).toHaveLength(1)

    act(() => {
      result.current.removeCondition(0)
    })

    expect(result.current.conditions).toHaveLength(0)
  })

  it('updates a condition when updateCondition is called', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    const condition = {
      id: '1',
      metricType: 'error_count' as const,
      operator: '>' as const,
      threshold: 5,
      timeWindow: 10,
    }

    act(() => {
      result.current.addCondition(condition)
    })

    const updatedCondition = {
      ...condition,
      threshold: 10,
    }

    act(() => {
      result.current.updateCondition(0, updatedCondition)
    })

    expect(result.current.conditions[0].threshold).toBe(10)
  })

  it('validates name is required', async () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    act(() => {
      result.current.addCondition({
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      })
    })

    await act(async () => {
      await result.current.validateAndSave()
    })

    expect(result.current.errors.some((e) => e.field === 'name')).toBe(true)
  })

  it('validates at least one condition is required', async () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    act(() => {
      result.current.setName('Test Alert')
    })

    await act(async () => {
      await result.current.validateAndSave()
    })

    expect(result.current.errors.some((e) => e.field === 'conditions')).toBe(true)
  })

  it('clears errors when fields are updated', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    act(() => {
      result.current.clearErrors()
    })

    expect(result.current.errors).toEqual([])
  })

  it('resets all state when reset is called', () => {
    const { result } = renderHook(() => useAlertRuleBuilder())

    act(() => {
      result.current.setName('Test Alert')
      result.current.setDescription('Test Description')
      result.current.addCondition({
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      })
    })

    act(() => {
      result.current.reset()
    })

    expect(result.current.name).toBe('')
    expect(result.current.description).toBe('')
    expect(result.current.conditions).toEqual([])
    expect(result.current.errors).toEqual([])
  })

  it('successfully saves a valid alert rule', async () => {
    const mockCreateAlertRule = vi.fn().mockResolvedValue({ id: 'rule-1' })
    const mockValidateCondition = vi.fn().mockResolvedValue({ valid: true })

    vi.mocked(apiService.apiClient).validateAlertCondition = mockValidateCondition
    vi.mocked(apiService.apiClient).createAlertRule = mockCreateAlertRule

    const { result } = renderHook(() => useAlertRuleBuilder())

    act(() => {
      result.current.setName('Test Alert')
      result.current.addCondition({
        id: '1',
        metricType: 'error_count',
        operator: '>',
        threshold: 5,
        timeWindow: 10,
      })
    })

    await act(async () => {
      await result.current.validateAndSave()
    })

    await waitFor(() => {
      expect(mockCreateAlertRule).toHaveBeenCalled()
    })
  })
})
