import React, { useCallback } from 'react'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { LoadingSpinner } from '../ui/LoadingSpinner'
import { ConditionBuilder } from './ConditionBuilder'
import { useAlertRuleBuilder } from '../../hooks/useAlertRuleBuilder'

interface AlertRuleBuilderProps {
  onSuccess?: () => void
  onCancel?: () => void
}

export const AlertRuleBuilder: React.FC<AlertRuleBuilderProps> = ({
  onSuccess,
  onCancel,
}) => {
  const {
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
    clearErrors,
  } = useAlertRuleBuilder()

  const handleSubmit = useCallback(async () => {
    await validateAndSave()
    if (errors.length === 0 && onSuccess) {
      onSuccess()
    }
  }, [validateAndSave, errors, onSuccess])

  const getFieldError = (fieldName: string): string | undefined => {
    return errors.find((e) => e.field === fieldName)?.message
  }

  return (
    <div className="w-full max-w-2xl mx-auto space-y-6 p-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white">
          Create Alert Rule
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
          Set up a new alert rule with conditions and notifications
        </p>
      </div>

      {/* General Error */}
      {getFieldError('form') && (
        <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-4">
          <p className="text-sm text-red-800 dark:text-red-200">
            {getFieldError('form')}
          </p>
        </div>
      )}

      {/* Name Field */}
      <div className="space-y-2">
        <label htmlFor="name" className="block text-sm font-medium text-slate-700 dark:text-slate-300">
          Alert Rule Name
          <span className="text-red-500 ml-1">*</span>
        </label>
        <Input
          id="name"
          type="text"
          placeholder="e.g., High Error Rate Alert"
          value={name}
          onChange={(e) => {
            setName(e.target.value)
            clearErrors()
          }}
          disabled={isSaving || isLoading}
          className={getFieldError('name') ? 'border-red-500' : ''}
        />
        {getFieldError('name') && (
          <p className="text-xs text-red-600 dark:text-red-400">
            {getFieldError('name')}
          </p>
        )}
      </div>

      {/* Description Field */}
      <div className="space-y-2">
        <label htmlFor="description" className="block text-sm font-medium text-slate-700 dark:text-slate-300">
          Description
          <span className="text-slate-400 ml-1">(Optional)</span>
        </label>
        <textarea
          id="description"
          placeholder="Describe what this alert rule monitors"
          value={description}
          onChange={(e) => {
            setDescription(e.target.value)
            clearErrors()
          }}
          disabled={isSaving || isLoading}
          rows={3}
          className="w-full px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
        />
      </div>

      {/* Conditions Section */}
      <div className="space-y-3">
        <ConditionBuilder
          conditions={conditions}
          onAddCondition={addCondition}
          onRemoveCondition={removeCondition}
          onUpdateCondition={updateCondition}
        />
        {getFieldError('conditions') && (
          <p className="text-xs text-red-600 dark:text-red-400">
            {getFieldError('conditions')}
          </p>
        )}
      </div>

      {/* Loading State */}
      {(isLoading || isSaving) && (
        <div className="flex items-center gap-2 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-900">
          <LoadingSpinner size="sm" />
          <span className="text-sm text-blue-800 dark:text-blue-200">
            {isLoading ? 'Validating...' : 'Saving alert rule...'}
          </span>
        </div>
      )}

      {/* Action Buttons */}
      <div className="flex gap-3 justify-end pt-6 border-t border-slate-200 dark:border-slate-700">
        <Button
          variant="secondary"
          onClick={onCancel}
          disabled={isSaving || isLoading}
        >
          Cancel
        </Button>
        <Button
          variant="primary"
          onClick={handleSubmit}
          isLoading={isSaving || isLoading}
          disabled={isSaving || isLoading}
        >
          Create Alert Rule
        </Button>
      </div>
    </div>
  )
}
