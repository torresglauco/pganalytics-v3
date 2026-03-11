import React from 'react';
import { AlertCircle } from 'lucide-react';
import type { SlackConfig } from '../../types/notifications';

interface SlackChannelFormProps {
  config: SlackConfig | null;
  onChange: (config: SlackConfig) => void;
}

export const SlackChannelForm: React.FC<SlackChannelFormProps> = ({
  config,
  onChange,
}) => {
  const current = config || {
    webhook_url: '',
    channel: '',
    username: 'pgAnalytics Alerts',
    icon_emoji: '🚨',
  };

  return (
    <div className="space-y-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Webhook URL *
        </label>
        <input
          type="url"
          value={current.webhook_url}
          onChange={(e) =>
            onChange({
              ...current,
              webhook_url: e.target.value,
            })
          }
          placeholder="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
        />
        <p className="text-xs text-gray-600 mt-1">
          Get your webhook URL from Slack app configuration
        </p>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Channel (optional)
        </label>
        <input
          type="text"
          value={current.channel || ''}
          onChange={(e) =>
            onChange({
              ...current,
              channel: e.target.value || undefined,
            })
          }
          placeholder="#alerts or @username (uses webhook default if empty)"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Bot Username
        </label>
        <input
          type="text"
          value={current.username || ''}
          onChange={(e) =>
            onChange({
              ...current,
              username: e.target.value,
            })
          }
          placeholder="pgAnalytics Alerts"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Bot Emoji Icon
        </label>
        <input
          type="text"
          value={current.icon_emoji || ''}
          onChange={(e) =>
            onChange({
              ...current,
              icon_emoji: e.target.value,
            })
          }
          placeholder="🚨"
          maxLength={2}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          User/Group Mentions
        </label>
        <input
          type="text"
          value={current.mentions || ''}
          onChange={(e) =>
            onChange({
              ...current,
              mentions: e.target.value || undefined,
            })
          }
          placeholder="@channel or @team (for critical alerts)"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
        />
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="thread_replies"
          checked={current.thread_replies || false}
          onChange={(e) =>
            onChange({
              ...current,
              thread_replies: e.target.checked || undefined,
            })
          }
          className="rounded"
        />
        <label htmlFor="thread_replies" className="text-sm text-gray-700">
          Thread replies for alert updates
        </label>
      </div>

      <div className="p-3 bg-white border border-gray-200 rounded text-xs text-gray-600">
        <p>
          <strong>How to get a webhook URL:</strong>
        </p>
        <ol className="list-decimal list-inside mt-2 space-y-1">
          <li>Go to slack.com/apps and create a new app</li>
          <li>Enable "Incoming Webhooks" in App Features</li>
          <li>Click "Add New Webhook to Workspace" and select a channel</li>
          <li>Copy the webhook URL and paste it above</li>
        </ol>
      </div>
    </div>
  );
};

export default SlackChannelForm;
