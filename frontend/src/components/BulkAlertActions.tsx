import React, { useState } from 'react';
import { ChevronDown, AlertCircle } from 'lucide-react';

interface BulkAlertActionsProps {
  selectedCount: number;
  onAcknowledge?: (notes?: string) => Promise<void>;
  onResolve?: (notes?: string) => Promise<void>;
  onSnooze?: (minutes: number) => Promise<void>;
}

export const BulkAlertActions: React.FC<BulkAlertActionsProps> = ({
  selectedCount,
  onAcknowledge,
  onResolve,
  onSnooze,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notesInput, setNotesInput] = useState('');
  const [showNotes, setShowNotes] = useState(false);

  const handleAction = async (action: string) => {
    try {
      setIsLoading(true);
      setError(null);

      switch (action) {
        case 'acknowledge':
          await onAcknowledge?.(notesInput || undefined);
          break;
        case 'resolve':
          await onResolve?.(notesInput || undefined);
          break;
        default:
          break;
      }

      setIsOpen(false);
      setNotesInput('');
      setShowNotes(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 space-y-3">
      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700 flex gap-2">
          <AlertCircle size={16} className="flex-shrink-0 mt-0.5" />
          <div>{error}</div>
        </div>
      )}

      <div className="flex justify-between items-start">
        <p className="font-medium text-blue-900">
          {selectedCount} alert{selectedCount !== 1 ? 's' : ''} selected
        </p>

        <div className="relative">
          <button
            onClick={() => setIsOpen(!isOpen)}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg"
          >
            Actions
            <ChevronDown size={18} />
          </button>

          {isOpen && (
            <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-10">
              {onAcknowledge && (
                <button
                  onClick={() => {
                    setShowNotes(true);
                  }}
                  disabled={isLoading}
                  className="block w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50"
                >
                  Acknowledge
                </button>
              )}
              {onResolve && (
                <button
                  onClick={() => {
                    setShowNotes(true);
                  }}
                  disabled={isLoading}
                  className="block w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50"
                >
                  Resolve
                </button>
              )}
              {onSnooze && (
                <>
                  <button
                    onClick={() => onSnooze(5)}
                    disabled={isLoading}
                    className="block w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50"
                  >
                    Snooze 5 minutes
                  </button>
                  <button
                    onClick={() => onSnooze(30)}
                    disabled={isLoading}
                    className="block w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50"
                  >
                    Snooze 30 minutes
                  </button>
                </>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Notes Input */}
      {showNotes && (
        <div className="space-y-2 pt-2 border-t border-blue-200">
          <textarea
            value={notesInput}
            onChange={(e) => setNotesInput(e.target.value)}
            placeholder="Add optional notes (e.g., reason for action, ticket number)"
            rows={2}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
          />
          <div className="flex gap-2">
            <button
              onClick={() => handleAction('acknowledge')}
              disabled={isLoading}
              className="flex-1 px-3 py-2 bg-yellow-600 hover:bg-yellow-700 disabled:bg-gray-400 text-white font-medium rounded-lg text-sm"
            >
              Acknowledge
            </button>
            <button
              onClick={() => handleAction('resolve')}
              disabled={isLoading}
              className="flex-1 px-3 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white font-medium rounded-lg text-sm"
            >
              Resolve
            </button>
            <button
              onClick={() => setShowNotes(false)}
              disabled={isLoading}
              className="flex-1 px-3 py-2 border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium rounded-lg text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default BulkAlertActions;
