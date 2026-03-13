import { useState, useCallback } from 'react'
import { apiClient } from '../services/api'
import type { AlertCondition, AlertRuleBuilderData } from '../types/alerts'

interface ValidationError {
  field: string
  message: string
}

interface UseAlertRuleBuilderReturn {
  name: string
  description: string
  conditions: AlertCondition[]
  errors: ValidationError[]
  isLoading: boolean
  isSaving: boolean
  setName: (value: string) => void
  setDescription: (value: string) => void
  addCondition: (condition: AlertCondition) => void
  removeCondition: (index: number) => void
  updateCondition: (index: number, condition: AlertCondition) => void
  validateAndSave: () => Promise<void>
  reset: () => void
  clearErrors: () => void
}

export const useAlertRuleBuilder = (): UseAlertRuleBuilderReturn => {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [conditions, setConditions] = useState<AlertCondition[]>([])
  const [errors, setErrors] = useState<ValidationError[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  const clearErrors = useCallback(() => {
    setErrors([])
  }, [])

  const addCondition = useCallback((condition: AlertCondition) => {
    setConditions((prev) => [...prev, condition])
    clearErrors()
  }, [clearErrors])

  const removeCondition = useCallback((index: number) => {
    setConditions((prev) => prev.filter((_, i) => i !== index))
    clearErrors()
  }, [clearErrors])

  const updateCondition = useCallback((index: number, condition: AlertCondition) => {
    setConditions((prev) => {
      const updated = [...prev]
      updated[index] = condition
      return updated
    })
    clearErrors()
  }, [clearErrors])

  const validate = (): boolean => {
    const newErrors: ValidationError[] = []

    if (!name.trim()) {
      newErrors.push({
        field: 'name',
        message: 'Alert rule name is required',
      })
    }

    if (conditions.length === 0) {
      newErrors.push({
        field: 'conditions',
        message: 'At least one condition is required',
      })
    }

    // Validate each condition
    conditions.forEach((condition, index) => {
      if (condition.threshold === undefined || condition.threshold === null) {
        newErrors.push({
          field: `condition_${index}_threshold`,
          message: 'Threshold value is required',
        })
      }

      if (condition.timeWindow === undefined || condition.timeWindow === null || condition.timeWindow <= 0) {
        newErrors.push({
          field: `condition_${index}_timeWindow`,
          message: 'Time window must be greater than 0',
        })
      }
    })

    setErrors(newErrors)
    return newErrors.length === 0
  }

  const validateAndSave = useCallback(async () => {
    if (!validate()) {
      return
    }

    setIsSaving(true)
    try {
      // Validate each condition with the API
      setIsLoading(true)
      for (let i = 0; i < conditions.length; i++) {
        const condition = conditions[i]
        try {
          await apiClient.validateAlertCondition(condition)
        } catch (err) {
          setErrors((prev) => [
            ...prev,
            {
              field: `condition_${i}`,
              message: `Condition validation failed: ${err instanceof Error ? err.message : 'Unknown error'}`,
            },
          ])
          setIsSaving(false)
          return
        }
      }

      // Save the rule
      const ruleData: AlertRuleBuilderData = {
        name: name.trim(),
        description: description.trim(),
        conditions,
      }

      await apiClient.createAlertRule(ruleData)

      // Reset on success
      reset()
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to save alert rule'
      setErrors([
        {
          field: 'form',
          message: errorMessage,
        },
      ])
    } finally {
      setIsLoading(false)
      setIsSaving(false)
    }
  }, [conditions, name, description])

  const reset = useCallback(() => {
    setName('')
    setDescription('')
    setConditions([])
    setErrors([])
  }, [])

  return {
    name,
    description,
    conditions,
    errors,
    isLoading,
    isSaving,
    setName,
    setDescription,
    addCondition,
    removeCondition,
    updateCondition,
    validateAndSave,
    reset,
    clearErrors,
  }
}
