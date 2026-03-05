import React from 'react';
import type { EmailConfig } from '../../types/notifications';

interface EmailChannelFormProps {
  config: EmailConfig | null;
  onChange: (config: EmailConfig) => void;
}

export const EmailChannelForm: React.FC<EmailChannelFormProps> = ({
  config,
  onChange,
}) => {
  const current = config || {
    smtp_server: '',
    smtp_port: 587,
    from_address: '',
    recipients: [],
    smtp_use_tls: true,
  };

  const handleAddRecipient = (email: string) => {
    if (email && !current.recipients.includes(email)) {
      onChange({
        ...current,
        recipients: [...current.recipients, email],
      });
    }
  };

  const handleRemoveRecipient = (email: string) => {
    onChange({
      ...current,
      recipients: current.recipients.filter((r) => r !== email),
    });
  };

  return (
    <div className="space-y-4 p-4 bg-red-50 border border-red-200 rounded-lg">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            SMTP Server *
          </label>
          <input
            type="text"
            value={current.smtp_server}
            onChange={(e) =>
              onChange({
                ...current,
                smtp_server: e.target.value,
              })
            }
            placeholder="smtp.gmail.com"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            SMTP Port *
          </label>
          <input
            type="number"
            value={current.smtp_port}
            onChange={(e) =>
              onChange({
                ...current,
                smtp_port: parseInt(e.target.value),
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent"
          />
        </div>
      </div>

      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="smtp_use_tls"
          checked={current.smtp_use_tls}
          onChange={(e) =>
            onChange({
              ...current,
              smtp_use_tls: e.target.checked,
            })
          }
          className="rounded"
        />
        <label htmlFor="smtp_use_tls" className="text-sm text-gray-700">
          Use TLS/SSL
        </label>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Username
          </label>
          <input
            type="text"
            value={current.smtp_username || ''}
            onChange={(e) =>
              onChange({
                ...current,
                smtp_username: e.target.value || undefined,
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Password
          </label>
          <input
            type="password"
            value={current.smtp_password || ''}
            onChange={(e) =>
              onChange({
                ...current,
                smtp_password: e.target.value || undefined,
              })
            }
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            From Address *
          </label>
          <input
            type="email"
            value={current.from_address}
            onChange={(e) =>
              onChange({
                ...current,
                from_address: e.target.value,
              })
            }
            placeholder="alerts@example.com"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            From Name
          </label>
          <input
            type="text"
            value={current.from_name || ''}
            onChange={(e) =>
              onChange({
                ...current,
                from_name: e.target.value || undefined,
              })
            }
            placeholder="pgAnalytics Alerts"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Recipients *
        </label>
        <div className="flex gap-2 mb-2">
          <input
            type="email"
            id="new_recipient"
            placeholder="user@example.com"
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent"
          />
          <button
            type="button"
            onClick={() => {
              const input = document.getElementById('new_recipient') as HTMLInputElement;
              if (input && input.value) {
                handleAddRecipient(input.value);
                input.value = '';
              }
            }}
            className="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg font-medium"
          >
            Add
          </button>
        </div>
        <div className="flex flex-wrap gap-2">
          {current.recipients.map((email) => (
            <div
              key={email}
              className="bg-red-100 text-red-800 px-3 py-1 rounded-full flex items-center gap-2 text-sm"
            >
              {email}
              <button
                type="button"
                onClick={() => handleRemoveRecipient(email)}
                className="hover:text-red-900 font-bold"
              >
                ×
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default EmailChannelForm;
