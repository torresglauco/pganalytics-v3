import React from 'react';

interface OpsGenieConfig {
  api_key: string;
  region?: 'us' | 'eu';
  team_id?: string;
  tags?: string[];
}

interface OpsGenieChannelFormProps {
  config: OpsGenieConfig | null;
  onChange: (config: OpsGenieConfig) => void;
}

export const OpsGenieChannelForm: React.FC<OpsGenieChannelFormProps> = ({
  config,
  onChange,
}) => {
  const handleChange = (field: keyof OpsGenieConfig, value: string | string[]) => {
    onChange({
      ...config,
      api_key: config?.api_key || '',
      [field]: value,
    });
  };

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          API Key *
        </label>
        <input
          type="password"
          value={config?.api_key || ''}
          onChange={(e) => handleChange('api_key', e.target.value)}
          placeholder="OpsGenie API Key"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        <p className="text-xs text-gray-500 mt-1">
          Find your API key in OpsGenie Settings {'>'} API Key Management
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Region
        </label>
        <select
          value={config?.region || 'us'}
          onChange={(e) => handleChange('region', e.target.value as 'us' | 'eu')}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        >
          <option value="us">US (api.opsgenie.com)</option>
          <option value="eu">EU (api.eu.opsgenie.com)</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Team ID
        </label>
        <input
          type="text"
          value={config?.team_id || ''}
          onChange={(e) => handleChange('team_id', e.target.value)}
          placeholder="Optional: Team ID for routing"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        <p className="text-xs text-gray-500 mt-1">
          Route alerts to a specific team
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Tags (comma-separated)
        </label>
        <input
          type="text"
          value={config?.tags?.join(', ') || ''}
          onChange={(e) =>
            handleChange(
              'tags',
              e.target.value.split(',').map((t) => t.trim()).filter(Boolean)
            )
          }
          placeholder="e.g., production, database, critical"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        <p className="text-xs text-gray-500 mt-1">
          Additional tags to add to alerts
        </p>
      </div>

      <div className="bg-blue-50 p-4 rounded-lg text-sm text-blue-700">
        <p className="font-medium mb-1">Priority Mapping</p>
        <p className="text-xs">
          Alert severities are automatically mapped to OpsGenie priorities:
          <br />
          Critical {'>'} P1, High {'>'} P2, Medium {'>'} P3, Low {'>'} P4
        </p>
      </div>
    </div>
  );
};

export default OpsGenieChannelForm;