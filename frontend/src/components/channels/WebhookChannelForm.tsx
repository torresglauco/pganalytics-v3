import React from 'react';
import type { WebhookConfig } from '../../types/notifications';

interface WebhookChannelFormProps {
  config: WebhookConfig | null;
  onChange: (config: WebhookConfig) => void;
}

export const WebhookChannelForm: React.FC<WebhookChannelFormProps> = ({
  config,
  onChange,
}) => {
  const current = config || {
    url: '',
    method: 'POST',
    auth_type: 'none',
  };

  return (
    <div className="space-y-4 p-4 bg-purple-50 border border-purple-200 rounded-lg">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          URL *
        </label>
        <input
          type="url"
          value={current.url}
          onChange={(e) =>
            onChange({
              ...current,
              url: e.target.value,
            })
          }
          placeholder="https://example.com/webhooks/alerts"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Method
          </label>
          <select
            value={current.method}
            onChange={(e) =>
              onChange({
                ...current,
                method: e.target.value as 'POST' | 'PUT' | 'PATCH',
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          >
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="PATCH">PATCH</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Auth Type
          </label>
          <select
            value={current.auth_type || 'none'}
            onChange={(e) =>
              onChange({
                ...current,
                auth_type: (e.target.value as any) || 'none',
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          >
            <option value="none">None</option>
            <option value="basic">Basic Auth</option>
            <option value="bearer">Bearer Token</option>
            <option value="api_key">API Key</option>
          </select>
        </div>
      </div>

      {current.auth_type === 'basic' && (
        <div className="grid grid-cols-2 gap-4">
          <input
            type="text"
            placeholder="Username"
            value={current.auth_credentials?.username || ''}
            onChange={(e) =>
              onChange({
                ...current,
                auth_credentials: {
                  ...current.auth_credentials,
                  username: e.target.value,
                },
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          />
          <input
            type="password"
            placeholder="Password"
            value={current.auth_credentials?.password || ''}
            onChange={(e) =>
              onChange({
                ...current,
                auth_credentials: {
                  ...current.auth_credentials,
                  password: e.target.value,
                },
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          />
        </div>
      )}

      {current.auth_type === 'bearer' && (
        <input
          type="password"
          placeholder="Bearer Token"
          value={current.auth_credentials?.token || ''}
          onChange={(e) =>
            onChange({
              ...current,
              auth_credentials: {
                ...current.auth_credentials,
                token: e.target.value,
              },
            })
          }
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
        />
      )}

      {current.auth_type === 'api_key' && (
        <div className="grid grid-cols-2 gap-4">
          <input
            type="text"
            placeholder="Header Name"
            value={current.auth_credentials?.api_key_header || ''}
            onChange={(e) =>
              onChange({
                ...current,
                auth_credentials: {
                  ...current.auth_credentials,
                  api_key_header: e.target.value,
                },
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          />
          <input
            type="password"
            placeholder="API Key Value"
            value={current.auth_credentials?.api_key_value || ''}
            onChange={(e) =>
              onChange({
                ...current,
                auth_credentials: {
                  ...current.auth_credentials,
                  api_key_value: e.target.value,
                },
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          />
        </div>
      )}

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Custom Headers (JSON)
        </label>
        <textarea
          value={JSON.stringify(current.headers || {}, null, 2)}
          onChange={(e) => {
            try {
              const headers = JSON.parse(e.target.value);
              onChange({
                ...current,
                headers,
              });
            } catch {
              // Invalid JSON, ignore
            }
          }}
          placeholder='{"X-Custom-Header": "value"}'
          rows={3}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent font-mono text-sm"
        />
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="retry_enabled"
          checked={current.retry_enabled || false}
          onChange={(e) =>
            onChange({
              ...current,
              retry_enabled: e.target.checked || undefined,
            })
          }
          className="rounded"
        />
        <label htmlFor="retry_enabled" className="text-sm text-gray-700">
          Enable automatic retries
        </label>
      </div>

      {current.retry_enabled && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Max Retry Attempts
          </label>
          <input
            type="number"
            value={current.retry_max_attempts || 3}
            onChange={(e) =>
              onChange({
                ...current,
                retry_max_attempts: parseInt(e.target.value),
              })
            }
            min="1"
            max="10"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
          />
        </div>
      )}
    </div>
  );
};

export default WebhookChannelForm;
