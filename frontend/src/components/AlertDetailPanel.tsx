import React, { useState, useEffect } from 'react';
import { X, AlertCircle, CheckCircle, AlertTriangle, Copy, ExternalLink } from 'lucide-react';
import type { AlertIncident, AlertEvent } from '../types/alertDashboard';
import { getAlertEvents, getRelatedAlerts } from '../api/alertDashboardApi';

interface AlertDetailPanelProps {
  alert: AlertIncident;
  onClose: () => void;
  onAcknowledge?: () => void;
  onResolve?: () => void;
}

export const AlertDetailPanel: React.FC<AlertDetailPanelProps> = ({
  alert,
  onClose,
  onAcknowledge,
  onResolve,
}) => {
  const [events, setEvents] = useState<AlertEvent[]>([]);
  const [relatedAlerts, setRelatedAlerts] = useState<AlertIncident[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [activeTab, setActiveTab] = useState<'details' | 'events' | 'related'>(
    'details'
  );
  const [notesInput, setNotesInput] = useState('');

  useEffect(() => {
    loadAlertData();
  }, [alert.id]);

  const loadAlertData = async () => {
    try {
      setIsLoading(true);
      const [eventsData, relatedData] = await Promise.all([
        getAlertEvents(alert.id, { limit: 20 }),
        getRelatedAlerts(alert.id),
      ]);
      setEvents(eventsData.events);
      setRelatedAlerts(relatedData);
    } catch (err) {
      console.error('Failed to load alert data', err);
    } finally {
      setIsLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800 border-red-300';
      case 'high':
        return 'bg-orange-100 text-orange-800 border-orange-300';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 border-yellow-300';
      default:
        return 'bg-blue-100 text-blue-800 border-blue-300';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'firing':
        return 'bg-red-100 text-red-800 border-red-300';
      case 'acknowledged':
        return 'bg-yellow-100 text-yellow-800 border-yellow-300';
      case 'resolved':
        return 'bg-green-100 text-green-800 border-green-300';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

  const getEventIcon = (eventType: string) => {
    switch (eventType) {
      case 'fired':
        return <AlertCircle className="text-red-600" size={20} />;
      case 'acknowledged':
        return <AlertTriangle className="text-yellow-600" size={20} />;
      case 'resolved':
        return <CheckCircle className="text-green-600" size={20} />;
      default:
        return <AlertCircle className="text-gray-600" size={20} />;
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-white border-b border-gray-200 p-6 flex justify-between items-start">
          <div className="flex-1">
            <h2 className="text-2xl font-bold text-gray-900">{alert.title}</h2>
            {alert.description && (
              <p className="text-sm text-gray-600 mt-1">{alert.description}</p>
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
        <div className="px-6 py-3 bg-gray-50 border-b border-gray-200 flex justify-between items-center flex-wrap gap-4">
          <div className="flex items-center gap-4">
            <div>
              <span className="text-xs text-gray-600 uppercase">Status</span>
              <p
                className={`px-2 py-1 rounded text-sm font-medium ${getStatusColor(
                  alert.status
                )} capitalize mt-1`}
              >
                {alert.status}
              </p>
            </div>
            <div>
              <span className="text-xs text-gray-600 uppercase">Severity</span>
              <p
                className={`px-2 py-1 rounded text-sm font-medium ${getSeverityColor(
                  alert.severity
                )} capitalize mt-1`}
              >
                {alert.severity}
              </p>
            </div>
            <div>
              <span className="text-xs text-gray-600 uppercase">Rule</span>
              <p className="text-sm font-medium text-gray-900 mt-1">
                {alert.alert_rule_name}
              </p>
            </div>
          </div>

          <div className="flex gap-2">
            {alert.status === 'firing' && onAcknowledge && (
              <button
                onClick={onAcknowledge}
                className="px-4 py-2 border border-yellow-600 text-yellow-600 hover:bg-yellow-50 font-medium rounded-lg"
              >
                Acknowledge
              </button>
            )}
            {alert.status !== 'resolved' && onResolve && (
              <button
                onClick={onResolve}
                className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg"
              >
                Resolve
              </button>
            )}
          </div>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-200">
          {(['details', 'events', 'related'] as const).map((tab) => (
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
        <div className="p-6 space-y-6">
          {activeTab === 'details' && (
            <div className="space-y-6">
              {/* Metric Information */}
              {alert.metric_name && (
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-3">
                    Metric Information
                  </h3>
                  <div className="bg-gray-50 p-4 rounded-lg space-y-2">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Metric:</span>
                      <span className="font-mono text-gray-900">
                        {alert.metric_name}
                      </span>
                    </div>
                    {alert.metric_value !== undefined && (
                      <div className="flex justify-between">
                        <span className="text-gray-600">Current Value:</span>
                        <span className="font-mono font-semibold text-gray-900">
                          {alert.metric_value.toFixed(2)}
                        </span>
                      </div>
                    )}
                    {alert.threshold_value !== undefined && (
                      <div className="flex justify-between">
                        <span className="text-gray-600">Threshold:</span>
                        <span className="font-mono font-semibold text-gray-900">
                          {alert.threshold_value.toFixed(2)}
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Timing Information */}
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Timeline</h3>
                <div className="bg-gray-50 p-4 rounded-lg space-y-3">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Fired:</span>
                    <span className="text-gray-900">
                      {new Date(alert.fired_at).toLocaleString()}
                    </span>
                  </div>
                  {alert.acknowledged_at && (
                    <div className="flex justify-between">
                      <span className="text-gray-600">Acknowledged:</span>
                      <span className="text-gray-900">
                        {new Date(alert.acknowledged_at).toLocaleString()}
                        {alert.acknowledged_by && (
                          <span className="text-xs text-gray-600 ml-2">
                            by {alert.acknowledged_by}
                          </span>
                        )}
                      </span>
                    </div>
                  )}
                  {alert.resolved_at && (
                    <div className="flex justify-between">
                      <span className="text-gray-600">Resolved:</span>
                      <span className="text-gray-900">
                        {new Date(alert.resolved_at).toLocaleString()}
                        {alert.resolved_by && (
                          <span className="text-xs text-gray-600 ml-2">
                            by {alert.resolved_by}
                          </span>
                        )}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              {/* Metadata */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Database</h4>
                  <p className="text-gray-900">{alert.database_id}</p>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Source Type</h4>
                  <p className="text-gray-900 capitalize">{alert.source_type}</p>
                </div>
                {alert.environment && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-2">Environment</h4>
                    <p className="text-gray-900">{alert.environment}</p>
                  </div>
                )}
                {alert.team && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-2">Team</h4>
                    <p className="text-gray-900">{alert.team}</p>
                  </div>
                )}
              </div>

              {/* Runbook */}
              {alert.runbook_url && (
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-3">Runbook</h3>
                  <a
                    href={alert.runbook_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 text-blue-600 hover:text-blue-700 font-medium"
                  >
                    View Runbook
                    <ExternalLink size={16} />
                  </a>
                </div>
              )}

              {/* Tags */}
              {alert.tags && alert.tags.length > 0 && (
                <div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-3">Tags</h3>
                  <div className="flex flex-wrap gap-2">
                    {alert.tags.map((tag) => (
                      <span
                        key={tag}
                        className="bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'events' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Event Timeline</h3>
              {isLoading ? (
                <p className="text-gray-600">Loading events...</p>
              ) : events.length === 0 ? (
                <p className="text-gray-600">No events recorded</p>
              ) : (
                <div className="space-y-3">
                  {events.map((event) => (
                    <div key={event.id} className="flex gap-3 p-3 bg-gray-50 rounded-lg">
                      <div className="flex-shrink-0 pt-0.5">
                        {getEventIcon(event.event_type)}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-gray-900 capitalize">
                          {event.event_type}
                        </p>
                        <p className="text-sm text-gray-600">{event.message}</p>
                        {event.actor && (
                          <p className="text-xs text-gray-500 mt-1">by {event.actor}</p>
                        )}
                      </div>
                      <div className="flex-shrink-0 text-right">
                        <p className="text-xs text-gray-500">
                          {new Date(event.timestamp).toLocaleTimeString()}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'related' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Related Alerts</h3>
              {isLoading ? (
                <p className="text-gray-600">Loading related alerts...</p>
              ) : relatedAlerts.length === 0 ? (
                <p className="text-gray-600">No related alerts</p>
              ) : (
                <div className="space-y-2">
                  {relatedAlerts.map((related) => (
                    <div key={related.id} className="p-3 bg-gray-50 rounded-lg">
                      <p className="font-medium text-gray-900">{related.title}</p>
                      <p className="text-sm text-gray-600">{related.alert_rule_name}</p>
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

export default AlertDetailPanel;
