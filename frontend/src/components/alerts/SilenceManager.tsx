import React, { useState, useEffect } from 'react';

interface Silence {
  id: string;
  alert_rule_id: string;
  reason: string;
  expires_at: string;
}

interface SilenceManagerProps {
  alertRuleId: string;
  alertRuleName: string;
  onSilenceCreated?: () => void;
}

export const SilenceManager: React.FC<SilenceManagerProps> = ({
  alertRuleId,
  alertRuleName,
  onSilenceCreated,
}) => {
  const [duration, setDuration] = useState(3600); // 1 hour default
  const [reason, setReason] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [activeSilences, setActiveSilences] = useState<Silence[]>([]);

  useEffect(() => {
    fetchActiveSilences();
  }, [alertRuleId]);

  const fetchActiveSilences = async () => {
    try {
      const response = await fetch(`/api/v1/alert-silences?alert_rule_id=${alertRuleId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('jwt_token')}`,
        },
      });
      if (response.ok) {
        const silences = await response.json();
        setActiveSilences(silences || []);
      }
    } catch (err) {
      console.error('Failed to fetch silences:', err);
    }
  };

  const handleCreateSilence = async () => {
    setIsLoading(true);
    setError('');

    try {
      const expiresAt = new Date(Date.now() + duration * 1000).toISOString();
      const response = await fetch('/api/v1/alert-silences', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('jwt_token')}`,
        },
        body: JSON.stringify({
          alert_rule_id: alertRuleId,
          duration_seconds: duration,
          reason: reason || 'Temporarily silenced',
          expires_at: expiresAt,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to create silence');
      }

      setDuration(3600);
      setReason('');
      await fetchActiveSilences();
      onSilenceCreated?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create silence');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeactivateSilence = async (silenceId: string) => {
    try {
      const response = await fetch(`/api/v1/alert-silences/${silenceId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('jwt_token')}`,
        },
      });

      if (response.ok) {
        await fetchActiveSilences();
      }
    } catch (err) {
      console.error('Failed to deactivate silence:', err);
    }
  };

  const getDurationLabel = (seconds: number) => {
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
    return `${Math.floor(seconds / 3600)}h`;
  };

  return (
    <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Silence Alert</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400">{alertRuleName}</p>

      <div className="space-y-3">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Duration
          </label>
          <select
            value={duration}
            onChange={(e) => setDuration(parseInt(e.target.value))}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          >
            <option value={300}>5 minutes</option>
            <option value={900}>15 minutes</option>
            <option value={1800}>30 minutes</option>
            <option value={3600}>1 hour</option>
            <option value={21600}>6 hours</option>
            <option value={86400}>1 day</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Reason (optional)
          </label>
          <input
            type="text"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="e.g., Maintenance window, False positive"
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          />
        </div>

        {error && (
          <div className="p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-sm text-red-700 dark:text-red-400">
            {error}
          </div>
        )}

        <button
          onClick={handleCreateSilence}
          disabled={isLoading}
          className="w-full px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 focus:outline-none focus:ring-2 focus:ring-yellow-500 disabled:bg-yellow-400 disabled:cursor-not-allowed"
        >
          {isLoading ? 'Creating...' : `Silence for ${getDurationLabel(duration)}`}
        </button>
      </div>

      {activeSilences.length > 0 && (
        <div className="border-t border-gray-200 dark:border-gray-600 pt-4">
          <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-2">Active Silences</h4>
          <div className="space-y-2">
            {activeSilences.map((silence) => (
              <div
                key={silence.id}
                className="flex justify-between items-center p-2 bg-white dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600"
              >
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white">{silence.reason}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Expires: {new Date(silence.expires_at).toLocaleString()}
                  </p>
                </div>
                <button
                  onClick={() => handleDeactivateSilence(silence.id)}
                  className="text-sm text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                >
                  Deactivate
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
