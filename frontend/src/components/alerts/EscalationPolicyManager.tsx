import React, { useState, useEffect } from 'react';

interface EscalationStep {
  step_number: number;
  wait_minutes: number;
  notification_channel: string;
  channel_config: Record<string, string>;
}

interface EscalationPolicy {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  steps: EscalationStep[];
}

interface EscalationPolicyManagerProps {
  alertRuleId: string;
  onPolicyLinked?: () => void;
}

export const EscalationPolicyManager: React.FC<EscalationPolicyManagerProps> = ({
  alertRuleId,
  onPolicyLinked,
}) => {
  const [policies, setPolicies] = useState<EscalationPolicy[]>([]);
  const [selectedPolicyId, setSelectedPolicyId] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchPolicies();
  }, []);

  const fetchPolicies = async () => {
    try {
      const response = await fetch('/api/v1/escalation-policies', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('jwt_token')}`,
        },
      });
      if (response.ok) {
        const data = await response.json();
        setPolicies(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch policies:', err);
    }
  };

  const handleLinkPolicy = async () => {
    if (!selectedPolicyId) return;

    setIsLoading(true);
    setError('');

    try {
      const response = await fetch(`/api/v1/alert-rules/${alertRuleId}/escalation-policies`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('jwt_token')}`,
        },
        body: JSON.stringify({
          policy_id: selectedPolicyId,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to link policy');
      }

      setSelectedPolicyId('');
      onPolicyLinked?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to link policy');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Escalation Policy</h3>

      <div className="space-y-3">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Select Policy
          </label>
          <select
            value={selectedPolicyId}
            onChange={(e) => setSelectedPolicyId(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          >
            <option value="">Choose a policy...</option>
            {policies
              .filter((p) => p.is_active)
              .map((policy) => (
                <option key={policy.id} value={policy.id}>
                  {policy.name}
                </option>
              ))}
          </select>
        </div>

        {selectedPolicyId && (
          <div className="p-3 bg-blue-50 dark:bg-blue-900/20 rounded border border-blue-200 dark:border-blue-800">
            {policies
              .filter((p) => p.id === selectedPolicyId)
              .map((policy) => (
                <div key={policy.id} className="space-y-2">
                  <p className="text-sm font-medium text-gray-900 dark:text-white">{policy.name}</p>
                  {policy.description && (
                    <p className="text-xs text-gray-600 dark:text-gray-400">{policy.description}</p>
                  )}
                  <div className="text-xs text-gray-600 dark:text-gray-400">
                    <p className="font-medium">Steps:</p>
                    <ul className="list-disc list-inside">
                      {policy.steps.map((step) => (
                        <li key={step.step_number}>
                          Step {step.step_number}: {step.notification_channel} after {step.wait_minutes}m
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              ))}
          </div>
        )}

        {error && (
          <div className="p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-sm text-red-700 dark:text-red-400">
            {error}
          </div>
        )}

        <button
          onClick={handleLinkPolicy}
          disabled={isLoading || !selectedPolicyId}
          className="w-full px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 disabled:bg-purple-400 disabled:cursor-not-allowed"
        >
          {isLoading ? 'Linking...' : 'Link Policy'}
        </button>
      </div>
    </div>
  );
};
