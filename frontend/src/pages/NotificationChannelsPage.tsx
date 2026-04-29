import React, { useState, useEffect } from 'react';
import {
  Plus,
  Search,
  Filter,
  Download,
  Upload,
  AlertCircle,
  MessageCircle,
  Mail,
  Webhook as WebhookIcon,
  Phone,
  FileText,
  CheckCircle,
  AlertTriangle,
} from 'lucide-react';
import type { NotificationChannel } from '../types/notifications';
import {
  listNotificationChannels,
  exportChannels,
  importChannels,
  bulkChannelAction,
  testNotificationChannel,
} from '../api/notificationsApi';
import NotificationChannelForm from '../components/NotificationChannelForm';
import ChannelDetailsModal from '../components/ChannelDetailsModal';

export const NotificationChannelsPage: React.FC = () => {
  const [channels, setChannels] = useState<NotificationChannel[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // UI State
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [selectedChannelId, setSelectedChannelId] = useState<string | null>(null);
  const [selectedChannels, setSelectedChannels] = useState<Set<string>>(new Set());

  // Filters
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [showFilters, setShowFilters] = useState(false);

  /**
   * Load channels
   */
  useEffect(() => {
    loadChannels();
  }, [searchTerm, typeFilter, statusFilter]);

  const loadChannels = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await listNotificationChannels({
        search: searchTerm || undefined,
        type: typeFilter || undefined,
        status: statusFilter || undefined,
      });
      setChannels(response.channels);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load channels');
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Handle channel creation
   */
  const handleChannelCreated = (channel: NotificationChannel) => {
    setChannels([channel, ...channels]);
    setShowCreateForm(false);
  };

  /**
   * Handle channel updated
   */
  const handleChannelUpdated = (channel: NotificationChannel) => {
    setChannels(channels.map((c) => (c.id === channel.id ? channel : c)));
    setSelectedChannelId(null);
  };

  /**
   * Handle channel deleted
   */
  const handleChannelDeleted = (channelId: string) => {
    setChannels(channels.filter((c) => c.id !== channelId));
    setSelectedChannelId(null);
  };

  /**
   * Test channel
   */
  const handleTestChannel = async (channelId: string) => {
    try {
      await testNotificationChannel(channelId);
      setError('Test message sent successfully');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Test failed');
    }
  };

  /**
   * Handle export
   */
  const handleExport = async () => {
    try {
      const blob = await exportChannels();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `notification-channels-${new Date().toISOString().split('T')[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Export failed');
    }
  };

  /**
   * Handle import
   */
  const handleImport = async (file: File) => {
    try {
      const result = await importChannels(file);
      await loadChannels();
      setError(
        `Imported ${result.imported} channels${
          result.skipped ? `, skipped ${result.skipped}` : ''
        }`
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed');
    }
  };

  /**
   * Handle bulk delete
   */
  const handleBulkDelete = async () => {
    if (selectedChannels.size === 0) return;
    if (!window.confirm('Delete selected channels? This cannot be undone.')) return;

    try {
      await bulkChannelAction({
        action: 'delete',
        channel_ids: Array.from(selectedChannels),
      });
      await loadChannels();
      setSelectedChannels(new Set());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Bulk delete failed');
    }
  };

  /**
   * Toggle channel selection
   */
  const toggleChannelSelection = (channelId: string) => {
    const newSelected = new Set(selectedChannels);
    if (newSelected.has(channelId)) {
      newSelected.delete(channelId);
    } else {
      newSelected.add(channelId);
    }
    setSelectedChannels(newSelected);
  };

  /**
   * Get channel icon
   */
  const getChannelIcon = (type: string) => {
    switch (type) {
      case 'slack':
        return <MessageCircle size={20} className="text-blue-500" />;
      case 'email':
        return <Mail size={20} className="text-red-500" />;
      case 'webhook':
        return <WebhookIcon size={20} className="text-purple-500" />;
      case 'pagerduty':
        return <Phone size={20} className="text-orange-500" />;
      case 'jira':
        return <FileText size={20} className="text-blue-600" />;
      default:
        return <AlertCircle size={20} />;
    }
  };

  if (showCreateForm) {
    return (
      <NotificationChannelForm
        onCreated={handleChannelCreated}
        onCancel={() => setShowCreateForm(false)}
      />
    );
  }

  if (selectedChannelId) {
    const channel = channels.find((c) => c.id === selectedChannelId);
    if (channel) {
      return (
        <ChannelDetailsModal
          channel={channel}
          onClose={() => setSelectedChannelId(null)}
          onUpdated={handleChannelUpdated}
          onDeleted={handleChannelDeleted}
          onTest={() => handleTestChannel(channel.id)}
        />
      );
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">
            Notification Channels
          </h2>
          <p className="text-sm text-gray-600 mt-1">
            Configure and manage notification delivery channels
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleExport}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700 font-medium"
          >
            <Download size={18} />
            Export
          </button>
          <label className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700 font-medium cursor-pointer">
            <Upload size={18} />
            Import
            <input
              type="file"
              accept=".json"
              onChange={(e) => {
                if (e.target.files?.[0]) {
                  handleImport(e.target.files[0]);
                }
              }}
              className="hidden"
            />
          </label>
          <button
            onClick={() => setShowCreateForm(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg"
          >
            <Plus size={18} />
            New Channel
          </button>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div
          className={`p-4 border rounded-lg flex gap-2 ${
            error.includes('successfully')
              ? 'bg-green-50 border-green-200 text-green-700'
              : 'bg-red-50 border-red-200 text-red-700'
          }`}
        >
          <AlertCircle size={20} className="flex-shrink-0 mt-0.5" />
          <div>{error}</div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg border border-gray-200 p-4 space-y-4">
        <div className="flex gap-2">
          <div className="flex-1 relative">
            <Search size={18} className="absolute left-3 top-2.5 text-gray-400" />
            <input
              type="text"
              placeholder="Search channels..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <button
            onClick={() => setShowFilters(!showFilters)}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700"
          >
            <Filter size={18} />
            Filters
          </button>
        </div>

        {showFilters && (
          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-gray-200">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Type
              </label>
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Types</option>
                <option value="slack">Slack</option>
                <option value="email">Email</option>
                <option value="webhook">Webhook</option>
                <option value="pagerduty">PagerDuty</option>
                <option value="jira">Jira</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Statuses</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="error">Error</option>
              </select>
            </div>
          </div>
        )}
      </div>

      {/* Bulk Actions */}
      {selectedChannels.size > 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 flex justify-between items-center">
          <p className="font-medium text-blue-900">
            {selectedChannels.size} channel{selectedChannels.size !== 1 ? 's' : ''}{' '}
            selected
          </p>
          <button
            onClick={handleBulkDelete}
            className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg"
          >
            Delete Selected
          </button>
        </div>
      )}

      {/* Channels Grid */}
      {isLoading ? (
        <div className="p-8 text-center text-gray-600">Loading channels...</div>
      ) : channels.length === 0 ? (
        <div className="p-8 text-center text-gray-600">
          <AlertCircle size={32} className="mx-auto mb-2 opacity-50" />
          <p>No notification channels configured yet</p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="mt-4 text-blue-600 hover:text-blue-700 font-medium"
          >
            Create your first channel
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {channels.map((channel) => (
            <div
              key={channel.id}
              className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition"
            >
              <div className="flex justify-between items-start mb-3">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-gray-100 rounded-lg">
                    {getChannelIcon(channel.type)}
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900">{channel.name}</h3>
                    <p className="text-xs text-gray-600 capitalize">
                      {channel.type}
                    </p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={selectedChannels.has(channel.id)}
                  onChange={() => toggleChannelSelection(channel.id)}
                  className="rounded"
                />
              </div>

              {channel.description && (
                <p className="text-sm text-gray-600 mb-3">{channel.description}</p>
              )}

              {/* Status Badge */}
              <div className="flex items-center gap-2 mb-3">
                {channel.status === 'active' ? (
                  <>
                    <CheckCircle size={16} className="text-green-600" />
                    <span className="text-sm text-green-600 font-medium">Active</span>
                  </>
                ) : channel.status === 'error' ? (
                  <>
                    <AlertTriangle size={16} className="text-red-600" />
                    <span className="text-sm text-red-600 font-medium">Error</span>
                    {channel.last_error && (
                      <span className="text-xs text-red-500 truncate">
                        {channel.last_error}
                      </span>
                    )}
                  </>
                ) : (
                  <>
                    <AlertCircle size={16} className="text-gray-400" />
                    <span className="text-sm text-gray-600 font-medium">
                      Inactive
                    </span>
                  </>
                )}
              </div>

              {/* Stats */}
              {channel.total_sent && (
                <div className="text-xs text-gray-600 mb-3">
                  {channel.success_count}/{channel.total_sent} successful deliveries
                </div>
              )}

              {/* Actions */}
              <div className="flex gap-2">
                <button
                  onClick={() => handleTestChannel(channel.id)}
                  className="flex-1 px-3 py-2 border border-gray-300 hover:bg-gray-50 text-gray-700 text-sm font-medium rounded"
                >
                  Test
                </button>
                <button
                  onClick={() => setSelectedChannelId(channel.id)}
                  className="flex-1 px-3 py-2 border border-gray-300 hover:bg-gray-50 text-gray-700 text-sm font-medium rounded"
                >
                  View
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default NotificationChannelsPage;
