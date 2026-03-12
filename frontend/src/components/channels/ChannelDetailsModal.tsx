import { useState } from 'react'
import { Modal } from '../ui/Modal'
import { Button } from '../ui/Button'
import { Badge } from '../ui/Badge'

interface ChannelDetailsModalProps {
  channel: any
  isOpen: boolean
  onClose: () => void
  onEdit: () => void
  onDelete: () => Promise<void>
  onTest: () => Promise<void>
}

export const ChannelDetailsModal: React.FC<ChannelDetailsModalProps> = ({
  channel,
  isOpen,
  onClose,
  onEdit,
  onDelete,
  onTest,
}) => {
  const [loading, setLoading] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [testStatus, setTestStatus] = useState<string | null>(null)

  const handleTest = async () => {
    setLoading(true)
    try {
      await onTest()
      setTestStatus('success')
      setTimeout(() => setTestStatus(null), 3000)
    } catch (err) {
      setTestStatus('error')
      setTimeout(() => setTestStatus(null), 3000)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    setLoading(true)
    try {
      await onDelete()
      onClose()
    } finally {
      setLoading(false)
    }
  }

  const getConfigDisplay = () => {
    switch (channel?.type) {
      case 'email':
        return <div className="text-sm text-slate-600 dark:text-slate-400">{channel.config.recipient}</div>
      case 'slack':
        return <div className="text-sm text-slate-600 dark:text-slate-400">Slack Webhook Configured</div>
      case 'pagerduty':
        return <div className="text-sm text-slate-600 dark:text-slate-400">PagerDuty Integration Key Configured</div>
      case 'webhook':
        return <div className="text-sm text-slate-600 dark:text-slate-400">{channel.config.url}</div>
      default:
        return null
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <div className="space-y-6">
        <div>
          <div className="flex items-start justify-between">
            <div>
              <h2 className="text-xl font-semibold text-slate-900 dark:text-white">
                {channel?.name}
              </h2>
              <div className="mt-2">
                <Badge variant={channel?.type === 'email' ? 'info' : 'success'}>
                  {channel?.type?.toUpperCase()}
                </Badge>
              </div>
            </div>
          </div>
        </div>

        <div>
          <h3 className="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
            Configuration
          </h3>
          {getConfigDisplay()}
        </div>

        {testStatus && (
          <div
            className={`rounded-lg border p-3 ${
              testStatus === 'success'
                ? 'border-emerald-200 bg-emerald-50 dark:border-emerald-900 dark:bg-emerald-900/20'
                : 'border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20'
            }`}
          >
            <div
              className={
                testStatus === 'success'
                  ? 'text-emerald-800 dark:text-emerald-200'
                  : 'text-red-800 dark:text-red-200'
              }
            >
              {testStatus === 'success'
                ? 'Test message sent successfully!'
                : 'Failed to send test message'}
            </div>
          </div>
        )}

        {showDeleteConfirm ? (
          <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-4">
            <p className="text-red-800 dark:text-red-200 mb-4">
              Are you sure you want to delete this channel? Alerts using this channel will need to be updated.
            </p>
            <div className="flex gap-3">
              <Button
                variant="secondary"
                onClick={() => setShowDeleteConfirm(false)}
                disabled={loading}
              >
                Cancel
              </Button>
              <Button
                variant="danger"
                onClick={handleDelete}
                isLoading={loading}
              >
                Delete Channel
              </Button>
            </div>
          </div>
        ) : (
          <div className="flex gap-3 justify-end flex-wrap">
            <Button variant="secondary" onClick={onClose} disabled={loading}>
              Close
            </Button>
            <Button
              variant="secondary"
              onClick={handleTest}
              isLoading={loading}
            >
              Send Test Message
            </Button>
            <Button variant="primary" onClick={onEdit} disabled={loading}>
              Edit
            </Button>
            <Button
              variant="danger"
              onClick={() => setShowDeleteConfirm(true)}
              disabled={loading}
            >
              Delete
            </Button>
          </div>
        )}
      </div>
    </Modal>
  )
}
