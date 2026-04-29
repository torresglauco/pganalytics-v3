import React, { useState } from 'react';
import { ChevronDown } from 'lucide-react';

interface BulkRuleActionsProps {
  selectedCount: number;
  onAction: (action: string, value?: any) => Promise<void>;
}

export const BulkRuleActions: React.FC<BulkRuleActionsProps> = ({
  selectedCount,
  onAction,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAction = async (action: string, value?: any) => {
    try {
      setIsLoading(true);
      setError(null);
      await onAction(action, value);
      setIsOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 space-y-3">
      <div className="flex justify-between items-center">
        <p className="font-medium text-blue-900">
          {selectedCount} rule{selectedCount !== 1 ? 's' : ''} selected
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
              <button
                onClick={() => handleAction('enable')}
                disabled={isLoading}
                className="block w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50"
              >
                Enable
              </button>
              <button
                onClick={() => handleAction('disable')}
                disabled={isLoading}
                className="block w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50"
              >
                Disable
              </button>
              <div className="border-t border-gray-200"></div>

              <button
                onClick={() => {
                  const severity = prompt(
                    'Enter new severity (low/medium/high/critical):'
                  );
                  if (severity) {
                    handleAction('update_severity', severity);
                  }
                }}
                disabled={isLoading}
                className="block w-full text-left px-4 py-2 hover:bg-gray-50 disabled:opacity-50"
              >
                Change Severity
              </button>

              <div className="border-t border-gray-200"></div>
              <button
                onClick={() => {
                  if (
                    window.confirm(
                      `Delete ${selectedCount} rule${selectedCount !== 1 ? 's' : ''}? This cannot be undone.`
                    )
                  ) {
                    handleAction('delete');
                  }
                }}
                disabled={isLoading}
                className="block w-full text-left px-4 py-2 hover:bg-red-50 text-red-600 disabled:opacity-50"
              >
                Delete
              </button>
            </div>
          )}
        </div>
      </div>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700 flex gap-2">
          <AlertCircle size={16} className="flex-shrink-0 mt-0.5" />
          <div>{error}</div>
        </div>
      )}
    </div>
  );
};

export default BulkRuleActions;
