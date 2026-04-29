import React, { useState } from 'react';
import { AlertCircle } from 'lucide-react';
import type {
  CreateChannelRequest,
  ChannelType,
  NotificationChannel,
} from '../types/notifications';
import {
  createNotificationChannel,
  validateChannelConfig,
} from '../api/notificationsApi';
import SlackChannelForm from './channels/SlackChannelForm';
import EmailChannelForm from './channels/EmailChannelForm';
import WebhookChannelForm from './channels/WebhookChannelForm';
import PagerDutyChannelForm from './channels/PagerDutyChannelForm';
import JiraChannelForm from './channels/JiraChannelForm';

interface NotificationChannelFormProps {
  onCreated?: (channel: NotificationChannel) => void;
  onCancel?: () => void;
}

export const NotificationChannelForm: React.FC<NotificationChannelFormProps> = ({
  onCreated,
  onCancel,
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [channelType, setChannelType] = useState<ChannelType | null>(null);
  const [config, setConfig] = useState<any>(null);

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  /**
   * Handle form submission
   */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      setError('Channel name is required');
      return;
    }

    if (!channelType) {
      setError('Channel type is required');
      return;
    }

    if (!config) {
      setError('Channel configuration is required');
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      const request: CreateChannelRequest = {
        name: name.trim(),
        description: description.trim() || undefined,
        type: channelType,
        config,
      };

      // Validate before submitting
      const validation = await validateChannelConfig(request);
      if (!validation.valid) {
        const errorMap: Record<string, string> = {};
        validation.errors?.forEach((e) => {
          errorMap[e.field] = e.message;
        });
        setValidationErrors(errorMap);
        setError('Please fix the errors below');
        return;
      }

      const channel = await createNotificationChannel(request);
      onCreated?.(channel);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create channel');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Create Notification Channel</h2>
        <p className="text-sm text-gray-600 mt-1">
          Set up a new notification delivery channel
        </p>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 flex gap-2">
          <AlertCircle size={20} className="flex-shrink-0 mt-0.5" />
          <div>{error}</div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6 bg-white p-6 rounded-lg border border-gray-200">
        {/* Basic Info */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">Basic Information</h3>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Channel Name *
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Critical Alerts Slack"
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                validationErrors.name ? 'border-red-500' : 'border-gray-300'
              }`}
            />
            {validationErrors.name && (
              <p className="text-sm text-red-600 mt-1">{validationErrors.name}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description for this channel"
              rows={2}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* Channel Type Selection */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">Channel Type</h3>

          <div className="grid grid-cols-2 gap-3">
            {[
              {
                value: 'slack' as const,
                label: 'Slack',
                desc: 'Send alerts to Slack channels',
              },
              {
                value: 'email' as const,
                label: 'Email',
                desc: 'Send alerts via SMTP email',
              },
              {
                value: 'webhook' as const,
                label: 'Webhook',
                desc: 'Send alerts to HTTP endpoint',
              },
              {
                value: 'pagerduty' as const,
                label: 'PagerDuty',
                desc: 'Create incidents in PagerDuty',
              },
              {
                value: 'jira' as const,
                label: 'Jira',
                desc: 'Create tickets in Jira',
              },
            ].map((type) => (
              <button
                key={type.value}
                type="button"
                onClick={() => {
                  setChannelType(type.value);
                  setConfig(null);
                }}
                className={`p-3 border-2 rounded-lg text-left transition ${
                  channelType === type.value
                    ? 'border-blue-600 bg-blue-50'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <div className="font-medium text-gray-900">{type.label}</div>
                <div className="text-xs text-gray-600">{type.desc}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Channel-Specific Configuration */}
        {channelType && (
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-gray-900">Configuration</h3>

            {channelType === 'slack' && (
              <SlackChannelForm config={config} onChange={setConfig} />
            )}
            {channelType === 'email' && (
              <EmailChannelForm config={config} onChange={setConfig} />
            )}
            {channelType === 'webhook' && (
              <WebhookChannelForm config={config} onChange={setConfig} />
            )}
            {channelType === 'pagerduty' && (
              <PagerDutyChannelForm config={config} onChange={setConfig} />
            )}
            {channelType === 'jira' && (
              <JiraChannelForm config={config} onChange={setConfig} />
            )}
          </div>
        )}

        {/* Buttons */}
        <div className="flex justify-between pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700 font-medium"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isLoading || !name || !channelType || !config}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg"
          >
            {isLoading ? 'Creating...' : 'Create Channel'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default NotificationChannelForm;
