import { Modal } from '../ui/Modal'

interface LogDetailsModalProps {
  log: any
  onClose: () => void
}

export const LogDetailsModal: React.FC<LogDetailsModalProps> = ({ log, onClose }) => {
  return (
    <Modal isOpen={true} onClose={onClose} size="lg">
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-slate-900 dark:text-white">Log Details</h2>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
              Timestamp
            </label>
            <div className="mt-1 text-sm text-slate-600 dark:text-slate-400">
              {new Date(log.log_timestamp).toLocaleString()}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
              Level
            </label>
            <div className="mt-1 text-sm text-slate-600 dark:text-slate-400">
              {log.log_level}
            </div>
          </div>

          <div className="col-span-2">
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
              Message
            </label>
            <div className="mt-1 p-3 bg-slate-100 dark:bg-slate-800 rounded text-sm text-slate-700 dark:text-slate-300 break-words max-h-32 overflow-y-auto">
              {log.log_message}
            </div>
          </div>

          {log.source_location && (
            <div className="col-span-2">
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
                Source Location
              </label>
              <div className="mt-1 text-sm text-slate-600 dark:text-slate-400 font-mono">
                {log.source_location}
              </div>
            </div>
          )}

          {log.user_name && (
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
                User
              </label>
              <div className="mt-1 text-sm text-slate-600 dark:text-slate-400">
                {log.user_name}
              </div>
            </div>
          )}
        </div>
      </div>
    </Modal>
  )
}
