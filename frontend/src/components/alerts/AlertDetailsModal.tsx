import { useState } from 'react'
import { Modal } from '../ui/Modal'
import { Button } from '../ui/Button'
import { Badge } from '../ui/Badge'

interface AlertDetailsModalProps {
  alert: any
  isOpen: boolean
  onClose: () => void
  onEdit: () => void
  onDelete: () => Promise<void>
  onTest: () => Promise<void>
}

export const AlertDetailsModal: React.FC<AlertDetailsModalProps> = ({
  alert,
  isOpen,
  onClose,
  onEdit,
  onDelete,
  onTest,
}) => {
  const [loading, setLoading] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const handleDelete = async () => {
    setLoading(true)
    try {
      await onDelete()
      onClose()
    } finally {
      setLoading(false)
    }
  }

  const handleTest = async () => {
    setLoading(true)
    try {
      await onTest()
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <div className="space-y-6">
        <div>
          <div className="flex items-start justify-between">
            <div>
              <h2 className="text-xl font-semibold text-slate-900 dark:text-white">
                {alert?.name}
              </h2>
              <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
                {alert?.description}
              </p>
            </div>
            <Badge variant={alert?.enabled ? 'success' : 'default'}>
              {alert?.enabled ? 'Active' : 'Inactive'}
            </Badge>
          </div>
        </div>

        <div>
          <h3 className="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">
            Conditions
          </h3>
          <div className="space-y-2">
            {alert?.conditions?.map((condition: any, index: number) => (
              <div
                key={index}
                className="text-sm text-slate-600 dark:text-slate-400 p-2 bg-slate-50 dark:bg-slate-800 rounded"
              >
                {index > 0 && <span className="font-medium">AND </span>}
                <span>
                  {condition.field} {condition.operator} {condition.value}
                </span>
              </div>
            ))}
          </div>
        </div>

        {showDeleteConfirm ? (
          <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-4">
            <p className="text-red-800 dark:text-red-200 mb-4">
              Are you sure you want to delete this alert? This action cannot be undone.
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
                Delete Alert
              </Button>
            </div>
          </div>
        ) : (
          <div className="flex gap-3 justify-end">
            <Button variant="secondary" onClick={onClose} disabled={loading}>
              Close
            </Button>
            <Button
              variant="secondary"
              onClick={handleTest}
              isLoading={loading}
            >
              Test Alert
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
