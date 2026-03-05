import React, { useState, useEffect } from 'react';
import { X, Trash2, CheckCircle, AlertTriangle, Copy } from 'lucide-react';
import type { NotificationChannel, DeliveryStats } from '../types/notifications';
import {
  deleteNotificationChannel,
  getChannelDeliveryStats,
  getChannelDeliveryHistory,
  toggleNotificationChannel,
} from '../api/notificationsApi';

interface ChannelDetailsModalProps {
  channel: NotificationChannel;
  onClose: () => void;
  onUpdated?: (channel: NotificationChannel) => void;
  onDeleted?: (channelId: string) => void;
  onTest?: () => void;
}

export const ChannelDetailsModal: React.FC<ChannelDetailsModalProps> = ({
  channel,
  onClose,
  onUpdated,
  onDeleted,
  onTest,
}) => {
  const [stats, setStats] = useState<DeliveryStats | null>(null);
  const [deliveries, setDeliveries] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'details' | 'stats' | 'deliveries'>(
    'details'
  );

  /**
   * Load channel stats and deliveries
   */
  useEffect(() => {
    loadChannelData();
  }, [channel.id]);

  const loadChannelData = async () => {
    try {
      setIsLoading(true);
      const [statsData, deliveryData] = await Promise.all([
        getChannelDeliveryStats(channel.id),
        getChannelDeliveryHistory(channel.id, { limit: 10 }),
      ]);
      setStats(statsData);
      setDeliveries(deliveryData.deliveries || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Handle channel toggle
   */
  const handleToggle = async () => {
    try {
      setIsLoading(true);
      const updated = await toggleNotificationChannel(
        channel.id,
        channel.status !== 'active'
      );
      onUpdated?.(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update');
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Handle channel deletion
   */
  const handleDelete = async () => {
    if (!window.confirm('Delete this channel? This cannot be undone.')) return;

    try {
      setIsDeleting(true);
      await deleteNotificationChannel(channel.id);
      onDeleted?.(channel.id);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete');
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-white border-b border-gray-200 p-6 flex justify-between items-start">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">{channel.name}</h2>
            {channel.description && (
              <p className="text-sm text-gray-600 mt-1">{channel.description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <X size={24} />
          </button>
        </div>

        {/* Status Bar */}
        <div className="px-6 py-3 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
          <div className="flex gap-4">
            <div>
              <span className="text-xs text-gray-600 uppercase">Type</span>
              <p className="text-sm font-medium text-gray-900 capitalize">
                {channel.type}
              </p>
            </div>
            <div>
              <span className="text-xs text-gray-600 uppercase">Status</span>
              <p
                className={`text-sm font-medium capitalize ${
                  channel.status === 'active'
                    ? 'text-green-700'
                    : channel.status === 'error'
                    ? 'text-red-700'
                    : 'text-gray-700'
                }`}
              >
                {channel.status}
              </p>
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={handleToggle}
              disabled={isLoading}
              className={`px-4 py-2 font-medium rounded-lg transition ${
                channel.status === 'active'
                  ? 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200'
                  : 'bg-green-100 text-green-800 hover:bg-green-200'
              }`}
            >
              {channel.status === 'active' ? 'Disable' : 'Enable'}
            </button>
            <button
              onClick={onTest}
              disabled={isLoading}
              className="px-4 py-2 border border-blue-600 text-blue-600 hover:bg-blue-50 font-medium rounded-lg"
            >
              Test
            </button>
            <button
              onClick={handleDelete}
              disabled={isDeleting || isLoading}
              className="px-4 py-2 bg-red-100 hover:bg-red-200 text-red-800 font-medium rounded-lg"
            >
              <Trash2 size={18} />
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-200">
          {(['details', 'stats', 'deliveries'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-6 py-3 font-medium border-b-2 transition ${
                activeTab === tab
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-600 hover:text-gray-900'
              }`}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          {activeTab === 'details' && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-3">
                  Configuration
                </h3>
                <div className="bg-gray-50 p-4 rounded-lg space-y-2 text-sm">
                  <pre className="text-xs font-mono whitespace-pre-wrap break-words">
                    {JSON.stringify(channel.config, null, 2).replace(
                      /"(password|token|api_key|secret)"\s*:\s*"[^"]*"/g,
                      '"$1": "****"'
                    )}
                  </pre>
                </div>
              </div>

              {channel.last_error && (
                <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
                  <h4 className="font-semibold text-red-900 mb-1">Last Error</h4>
                  <p className="text-sm text-red-700">{channel.last_error}</p>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-1">Created</h4>
                  <p className="text-gray-900">
                    {new Date(channel.created_at).toLocaleString()}
                  </p>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-1">Updated</h4>
                  <p className="text-gray-900">
                    {new Date(channel.updated_at).toLocaleString()}
                  </p>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-1">Created By</h4>
                  <p className="text-gray-900">{channel.created_by}</p>
                </div>
                {channel.last_test_at && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-1">
                      Last Test
                    </h4>
                    <p className="text-gray-900 flex items-center gap-2">
                      {new Date(channel.last_test_at).toLocaleString()}
                      {channel.last_test_status ? (
                        <CheckCircle size={16} className="text-green-600" />
                      ) : (
                        <AlertTriangle size={16} className="text-red-600" />
                      )}
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === 'stats' && stats && (
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-blue-50 p-4 rounded-lg">
                <p className="text-xs text-gray-600 uppercase">Total Sent</p>
                <p className="text-3xl font-bold text-blue-600">{stats.total}</p>
              </div>
              <div className="bg-green-50 p-4 rounded-lg">
                <p className="text-xs text-gray-600 uppercase">Successful</p>
                <p className="text-3xl font-bold text-green-600">{stats.sent}</p>
              </div>
              <div className="bg-red-50 p-4 rounded-lg">
                <p className="text-xs text-gray-600 uppercase">Failed</p>
                <p className="text-3xl font-bold text-red-600">{stats.failed}</p>
              </div>
              <div className="bg-yellow-50 p-4 rounded-lg">
                <p className="text-xs text-gray-600 uppercase">Success Rate</p>
                <p className="text-3xl font-bold text-yellow-600">
                  {(stats.success_rate * 100).toFixed(1)}%
                </p>
              </div>
              <div className="bg-gray-50 p-4 rounded-lg col-span-2">
                <p className="text-xs text-gray-600 uppercase mb-1">
                  Avg Response Time
                </p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats.avg_response_time_ms.toFixed(0)}ms
                </p>
              </div>
            </div>
          )}

          {activeTab === 'deliveries' && (
            <div>
              {deliveries.length === 0 ? (
                <p className="text-gray-600 text-center py-8">No deliveries yet</p>
              ) : (
                <div className="space-y-3">
                  {deliveries.map((delivery, idx) => (
                    <div key={idx} className="flex gap-3 p-3 bg-gray-50 rounded-lg">
                      {delivery.status === 'sent' ? (
                        <CheckCircle className="text-green-600 flex-shrink-0" size={20} />
                      ) : (
                        <AlertTriangle className="text-red-600 flex-shrink-0" size={20} />
                      )}
                      <div className="flex-1">
                        <p className="font-medium text-gray-900 capitalize">
                          {delivery.status}
                        </p>
                        <p className="text-xs text-gray-600">
                          {new Date(delivery.sent_at).toLocaleString()}
                        </p>
                      </div>
                      <span className="text-right text-gray-900 text-sm">
                        {delivery.response_time_ms}ms
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ChannelDetailsModal;
