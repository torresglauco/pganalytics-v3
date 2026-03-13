import React, { useState } from 'react';

interface AlertAcknowledgmentProps {
  alertId: string;
  isAcknowledged: boolean;
  onAcknowledged?: () => void;
}

export const AlertAcknowledgment: React.FC<AlertAcknowledgmentProps> = ({
  alertId,
  isAcknowledged,
  onAcknowledged,
}) => {
  const [note, setNote] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [acknowledged, setAcknowledged] = useState(isAcknowledged);

  const handleAcknowledge = async () => {
    setIsLoading(true);
    setError('');

    try {
      const response = await fetch(`/api/v1/alerts/${alertId}/acknowledge`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('jwt_token')}`,
        },
        body: JSON.stringify({
          note: note || 'Acknowledged',
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to acknowledge alert');
      }

      setAcknowledged(true);
      setNote('');
      onAcknowledged?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to acknowledge');
    } finally {
      setIsLoading(false);
    }
  };

  if (acknowledged) {
    return (
      <div className="p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
        <p className="text-sm text-green-700 dark:text-green-400">✓ Alert acknowledged</p>
      </div>
    );
  }

  return (
    <div className="space-y-2 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
      <p className="text-sm font-medium text-yellow-800 dark:text-yellow-400">Unacknowledged</p>
      <input
        type="text"
        value={note}
        onChange={(e) => setNote(e.target.value)}
        placeholder="Add a note..."
        className="w-full px-2 py-1 text-sm border border-yellow-300 rounded dark:bg-yellow-900/30 dark:border-yellow-700 dark:text-white"
      />
      {error && <p className="text-xs text-red-600 dark:text-red-400">{error}</p>}
      <button
        onClick={handleAcknowledge}
        disabled={isLoading}
        className="w-full px-2 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:bg-green-400 disabled:cursor-not-allowed"
      >
        {isLoading ? 'Acknowledging...' : 'Acknowledge'}
      </button>
    </div>
  );
};
