import { useState } from 'react'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Modal } from '../ui/Modal'

interface ChannelFormProps {
  channel?: any
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: any) => Promise<void>
}

export const ChannelForm: React.FC<ChannelFormProps> = ({
  channel,
  isOpen,
  onClose,
  onSubmit,
}) => {
  const [type, setType] = useState(channel?.type || 'email')
  const [formData, setFormData] = useState({
    name: channel?.name || '',
    type: channel?.type || 'email',
    config: channel?.config || {},
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async () => {
    if (!formData.name.trim()) {
      setError('Channel name is required')
      return
    }

    setLoading(true)
    try {
      await onSubmit(formData)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save channel')
    } finally {
      setLoading(false)
    }
  }

  const updateConfig = (key: string, value: string) => {
    setFormData({
      ...formData,
      config: { ...formData.config, [key]: value },
    })
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <div className="space-y-6">
        <h2 className="text-xl font-semibold text-slate-900 dark:text-white">
          {channel ? 'Edit Channel' : 'Create Channel'}
        </h2>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-3">
            <div className="text-sm text-red-800 dark:text-red-200">{error}</div>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
            Channel Name
          </label>
          <Input
            placeholder="e.g., Production Alerts"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
            Channel Type
          </label>
          <select
            value={formData.type}
            onChange={(e) => {
              setFormData({ ...formData, type: e.target.value, config: {} })
              setType(e.target.value)
            }}
            disabled={!!channel}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white disabled:opacity-50"
          >
            <option value="email">Email</option>
            <option value="slack">Slack</option>
            <option value="pagerduty">PagerDuty</option>
            <option value="webhook">Webhook</option>
          </select>
        </div>

        {/* Email Configuration */}
        {type === 'email' && (
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Recipient Email
            </label>
            <Input
              type="email"
              placeholder="alerts@example.com"
              value={formData.config.recipient || ''}
              onChange={(e) => updateConfig('recipient', e.target.value)}
            />
          </div>
        )}

        {/* Slack Configuration */}
        {type === 'slack' && (
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Webhook URL
            </label>
            <Input
              placeholder="https://hooks.slack.com/services/..."
              value={formData.config.webhook_url || ''}
              onChange={(e) => updateConfig('webhook_url', e.target.value)}
            />
          </div>
        )}

        {/* PagerDuty Configuration */}
        {type === 'pagerduty' && (
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Integration Key
            </label>
            <Input
              placeholder="Your PagerDuty integration key"
              type="password"
              value={formData.config.integration_key || ''}
              onChange={(e) => updateConfig('integration_key', e.target.value)}
            />
          </div>
        )}

        {/* Webhook Configuration */}
        {type === 'webhook' && (
          <>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                Webhook URL
              </label>
              <Input
                placeholder="https://example.com/webhook"
                value={formData.config.url || ''}
                onChange={(e) => updateConfig('url', e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                Authorization Header (optional)
              </label>
              <Input
                placeholder="Bearer token or Authorization header"
                value={formData.config.auth_header || ''}
                onChange={(e) => updateConfig('auth_header', e.target.value)}
              />
            </div>
          </>
        )}

        <div className="flex gap-3 justify-end">
          <Button variant="secondary" onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            isLoading={loading}
          >
            {channel ? 'Update' : 'Create'} Channel
          </Button>
        </div>
      </div>
    </Modal>
  )
}
