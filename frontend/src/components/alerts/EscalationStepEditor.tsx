import React, { useState } from 'react';

export interface EscalationStep {
  step_number: number;
  wait_minutes: number;
  notification_channel: 'slack' | 'pagerduty' | 'email' | 'sms' | 'webhook';
  channel_config: Record<string, string>;
  requires_acknowledgment?: boolean;
}

interface EscalationStepEditorProps {
  step: EscalationStep;
  stepNumber: number;
  onUpdate: (step: EscalationStep) => void;
  onRemove: () => void;
}

const CHANNEL_TYPES = ['slack', 'pagerduty', 'email', 'sms', 'webhook'] as const;

export const EscalationStepEditor: React.FC<EscalationStepEditorProps> = ({
  step,
  stepNumber,
  onUpdate,
  onRemove,
}) => {
  const [delayMinutes, setDelayMinutes] = useState(step.wait_minutes);
  const [channelType, setChannelType] = useState(step.notification_channel);
  const [requiresAck, setRequiresAck] = useState(step.requires_acknowledgment || false);
  const [channelConfig, setChannelConfig] = useState(step.channel_config || {});

  const handleUpdate = () => {
    onUpdate({
      ...step,
      wait_minutes: delayMinutes,
      notification_channel: channelType as EscalationStep['notification_channel'],
      requires_acknowledgment: requiresAck,
      channel_config: channelConfig,
    });
  };

  const handleDelayChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(e.target.value, 10);
    setDelayMinutes(value);
  };

  const handleChannelChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newChannel = e.target.value;
    setChannelType(newChannel as EscalationStep['notification_channel']);
    setChannelConfig({}); // Reset config when channel changes
  };

  const handleConfigChange = (key: string, value: string) => {
    const newConfig = { ...channelConfig, [key]: value };
    setChannelConfig(newConfig);
  };

  React.useEffect(() => {
    handleUpdate();
  }, [delayMinutes, channelType, requiresAck, channelConfig]);

  const getChannelFields = () => {
    switch (channelType) {
      case 'slack':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Channel ID
            </label>
            <input
              type="text"
              value={channelConfig.channel_id || ''}
              onChange={(e) => handleConfigChange('channel_id', e.target.value)}
              placeholder="e.g., C1234567890"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
          </div>
        );
      case 'email':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Recipients
            </label>
            <input
              type="text"
              value={channelConfig.recipients || ''}
              onChange={(e) => handleConfigChange('recipients', e.target.value)}
              placeholder="e.g., team@example.com"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
          </div>
        );
      case 'pagerduty':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Integration Key
            </label>
            <input
              type="text"
              value={channelConfig.integration_key || ''}
              onChange={(e) => handleConfigChange('integration_key', e.target.value)}
              placeholder="PagerDuty integration key"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
          </div>
        );
      case 'sms':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Phone Numbers
            </label>
            <input
              type="text"
              value={channelConfig.phone_numbers || ''}
              onChange={(e) => handleConfigChange('phone_numbers', e.target.value)}
              placeholder="e.g., +1234567890,+0987654321"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
          </div>
        );
      case 'webhook':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Webhook URL
            </label>
            <input
              type="text"
              value={channelConfig.url || ''}
              onChange={(e) => handleConfigChange('url', e.target.value)}
              placeholder="e.g., https://example.com/webhook"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="border border-gray-300 dark:border-gray-600 rounded-lg p-4 bg-white dark:bg-gray-800">
      <div className="flex justify-between items-center mb-4">
        <h4 className="text-sm font-semibold text-gray-900 dark:text-white">Step {stepNumber}</h4>
        <button
          onClick={onRemove}
          className="text-sm text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
        >
          Remove
        </button>
      </div>

      <div className="space-y-4">
        {/* Delay Input */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Delay (minutes)
          </label>
          <input
            type="number"
            min="0"
            value={delayMinutes}
            onChange={handleDelayChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          />
        </div>

        {/* Channel Type Dropdown */}
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Notification Channel
          </label>
          <select
            value={channelType}
            onChange={handleChannelChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          >
            {CHANNEL_TYPES.map((type) => (
              <option key={type} value={type}>
                {type.charAt(0).toUpperCase() + type.slice(1)}
              </option>
            ))}
          </select>
        </div>

        {/* Channel-specific Configuration */}
        {getChannelFields()}

        {/* Acknowledgment Checkbox */}
        <div className="flex items-center">
          <input
            type="checkbox"
            id={`ack-${stepNumber}`}
            checked={requiresAck}
            onChange={(e) => setRequiresAck(e.target.checked)}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600"
          />
          <label
            htmlFor={`ack-${stepNumber}`}
            className="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300"
          >
            Requires Acknowledgment
          </label>
        </div>
      </div>
    </div>
  );
};
