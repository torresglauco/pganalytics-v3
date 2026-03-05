import React from 'react';
import type { JiraConfig } from '../../types/notifications';

interface JiraChannelFormProps {
  config: JiraConfig | null;
  onChange: (config: JiraConfig) => void;
}

export const JiraChannelForm: React.FC<JiraChannelFormProps> = ({
  config,
  onChange,
}) => {
  const current = config || {
    base_url: '',
    project_key: '',
    username: '',
    api_token: '',
  };

  return (
    <div className="space-y-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Jira URL *
        </label>
        <input
          type="url"
          value={current.base_url}
          onChange={(e) =>
            onChange({
              ...current,
              base_url: e.target.value,
            })
          }
          placeholder="https://your-domain.atlassian.net"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Project Key *
          </label>
          <input
            type="text"
            value={current.project_key}
            onChange={(e) =>
              onChange({
                ...current,
                project_key: e.target.value,
              })
            }
            placeholder="PROJ"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Issue Type
          </label>
          <input
            type="text"
            value={current.issue_type || ''}
            onChange={(e) =>
              onChange({
                ...current,
                issue_type: e.target.value || undefined,
              })
            }
            placeholder="Bug or Incident"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Username *
          </label>
          <input
            type="text"
            value={current.username}
            onChange={(e) =>
              onChange({
                ...current,
                username: e.target.value,
              })
            }
            placeholder="your-email@example.com"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            API Token *
          </label>
          <input
            type="password"
            value={current.api_token}
            onChange={(e) =>
              onChange({
                ...current,
                api_token: e.target.value,
              })
            }
            placeholder="Your Jira API token"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="auto_close"
          checked={current.auto_close || false}
          onChange={(e) =>
            onChange({
              ...current,
              auto_close: e.target.checked || undefined,
            })
          }
          className="rounded"
        />
        <label htmlFor="auto_close" className="text-sm text-gray-700">
          Auto-close issue when alert resolves
        </label>
      </div>

      <div className="p-3 bg-white border border-gray-200 rounded text-xs text-gray-600">
        <p>
          <strong>Setup Instructions:</strong>
        </p>
        <ol className="list-decimal list-inside mt-2 space-y-1">
          <li>In Jira, go to Account Settings → Security</li>
          <li>Create an API Token for your account</li>
          <li>Use your email and the API token for authentication</li>
          <li>Find your project key in project settings</li>
        </ol>
      </div>
    </div>
  );
};

export default JiraChannelForm;
