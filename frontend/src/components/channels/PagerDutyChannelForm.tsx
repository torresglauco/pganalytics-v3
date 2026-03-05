import React from 'react';
import type { PagerDutyConfig } from '../../types/notifications';

interface PagerDutyChannelFormProps {
  config: PagerDutyConfig | null;
  onChange: (config: PagerDutyConfig) => void;
}

export const PagerDutyChannelForm: React.FC<PagerDutyChannelFormProps> = ({
  config,
  onChange,
}) => {
  const current = config || {
    integration_key: '',
    urgency: 'high',
  };

  return (
    <div className="space-y-4 p-4 bg-orange-50 border border-orange-200 rounded-lg">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Integration Key *
        </label>
        <input
          type="password"
          value={current.integration_key}
          onChange={(e) =>
            onChange({
              ...current,
              integration_key: e.target.value,
            })
          }
          placeholder="Your PagerDuty integration key"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent"
        />
        <p className="text-xs text-gray-600 mt-1">
          Found in PagerDuty service settings under Integrations
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Service ID (optional)
        </label>
        <input
          type="text"
          value={current.service_id || ''}
          onChange={(e) =>
            onChange({
              ...current,
              service_id: e.target.value || undefined,
            })
          }
          placeholder="Your PagerDuty service ID"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Escalation Policy ID (optional)
        </label>
        <input
          type="text"
          value={current.escalation_policy_id || ''}
          onChange={(e) =>
            onChange({
              ...current,
              escalation_policy_id: e.target.value || undefined,
            })
          }
          placeholder="Your escalation policy ID"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Incident Urgency
        </label>
        <select
          value={current.urgency || 'high'}
          onChange={(e) =>
            onChange({
              ...current,
              urgency: (e.target.value as 'low' | 'high') || 'high',
            })
          }
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent"
        >
          <option value="low">Low</option>
          <option value="high">High</option>
        </select>
      </div>

      <div className="p-3 bg-white border border-gray-200 rounded text-xs text-gray-600">
        <p>
          <strong>Setup Instructions:</strong>
        </p>
        <ol className="list-decimal list-inside mt-2 space-y-1">
          <li>In PagerDuty, go to Service → Integrations</li>
          <li>Click "New Vendor Integration" and select pgAnalytics</li>
          <li>Copy the Integration Key and paste it above</li>
          <li>Configure escalation policy if needed</li>
        </ol>
      </div>
    </div>
  );
};

export default PagerDutyChannelForm;
