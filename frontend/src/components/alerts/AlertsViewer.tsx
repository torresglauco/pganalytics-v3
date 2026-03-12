import { useState } from 'react'
import { useAlerts } from '../../hooks/useAlerts'
import { Button } from '../ui/Button'
import { AlertForm } from './AlertForm'
import { AlertDetailsModal } from './AlertDetailsModal'
import { AlertsTable } from './AlertsTable'

export const AlertsViewer: React.FC = () => {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [selectedAlert, setSelectedAlert] = useState<any>(null)
  const [editingAlert, setEditingAlert] = useState<any>(null)

  const { data, loading, error, createAlert, updateAlert, deleteAlert } = useAlerts()

  const alerts = data?.alerts || []

  const handleCreateAlert = async (formData: any) => {
    await createAlert(formData)
    setShowCreateForm(false)
  }

  const handleUpdateAlert = async (formData: any) => {
    if (editingAlert) {
      await updateAlert(editingAlert.id, formData)
      setEditingAlert(null)
    }
  }

  const handleDeleteAlert = async () => {
    if (selectedAlert) {
      await deleteAlert(selectedAlert.id)
      setSelectedAlert(null)
    }
  }

  const handleTestAlert = async () => {
    if (selectedAlert) {
      // Test alert via API
      console.log('Testing alert:', selectedAlert.id)
    }
  }

  if (error) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-4">
        <div className="text-red-800 dark:text-red-200">Error: {error}</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-end">
        <Button variant="primary" onClick={() => setShowCreateForm(true)}>
          + Create Alert
        </Button>
      </div>

      <AlertsTable
        alerts={alerts}
        loading={loading}
        onView={(alert) => setSelectedAlert(alert)}
      />

      {showCreateForm && (
        <AlertForm
          isOpen={showCreateForm}
          onClose={() => setShowCreateForm(false)}
          onSubmit={handleCreateAlert}
        />
      )}

      {editingAlert && (
        <AlertForm
          alert={editingAlert}
          isOpen={!!editingAlert}
          onClose={() => setEditingAlert(null)}
          onSubmit={handleUpdateAlert}
        />
      )}

      {selectedAlert && (
        <AlertDetailsModal
          alert={selectedAlert}
          isOpen={!!selectedAlert}
          onClose={() => setSelectedAlert(null)}
          onEdit={() => {
            setEditingAlert(selectedAlert)
            setSelectedAlert(null)
          }}
          onDelete={handleDeleteAlert}
          onTest={handleTestAlert}
        />
      )}
    </div>
  )
}
