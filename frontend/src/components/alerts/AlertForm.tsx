import { useState } from 'react'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Modal } from '../ui/Modal'
import { ConditionBuilder } from './ConditionBuilder'

interface AlertFormProps {
  alert?: any
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: any) => Promise<void>
}

export const AlertForm: React.FC<AlertFormProps> = ({
  alert,
  isOpen,
  onClose,
  onSubmit,
}) => {
  const [formData, setFormData] = useState({
    name: alert?.name || '',
    description: alert?.description || '',
    enabled: alert?.enabled ?? true,
    conditions: alert?.conditions || [],
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async () => {
    if (!formData.name.trim()) {
      setError('Alert name is required')
      return
    }

    if (formData.conditions.length === 0) {
      setError('At least one condition is required')
      return
    }

    setLoading(true)
    try {
      await onSubmit(formData)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save alert')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <div className="space-y-6">
        <h2 className="text-xl font-semibold text-slate-900 dark:text-white">
          {alert ? 'Edit Alert' : 'Create Alert'}
        </h2>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-3">
            <div className="text-sm text-red-800 dark:text-red-200">{error}</div>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
            Alert Name
          </label>
          <Input
            placeholder="e.g., High Error Rate"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
            Description
          </label>
          <textarea
            placeholder="Describe what this alert monitors"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white"
            rows={3}
          />
        </div>

        <ConditionBuilder
          conditions={formData.conditions}
          onConditionsChange={(conditions) =>
            setFormData({ ...formData, conditions })
          }
        />

        <div className="flex items-center">
          <input
            type="checkbox"
            id="enabled"
            checked={formData.enabled}
            onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
            className="rounded"
          />
          <label htmlFor="enabled" className="ml-2 text-sm text-slate-700 dark:text-slate-300">
            Enable this alert
          </label>
        </div>

        <div className="flex gap-3 justify-end">
          <Button variant="secondary" onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            isLoading={loading}
          >
            {alert ? 'Update' : 'Create'} Alert
          </Button>
        </div>
      </div>
    </Modal>
  )
}
