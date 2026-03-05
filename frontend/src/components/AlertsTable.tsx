import React from 'react';
import { AlertCircle, AlertTriangle, AlertOctagon, CheckCircle } from 'lucide-react';
import type { AlertIncident } from '../types/alertDashboard';

interface AlertsTableProps {
  alerts: AlertIncident[];
  selectedAlerts: Set<string>;
  onSelectAlert: (alertId: string) => void;
  onToggleSelection: (alertId: string) => void;
  onToggleSelectAll: () => void;
}

export const AlertsTable: React.FC<AlertsTableProps> = ({
  alerts,
  selectedAlerts,
  onSelectAlert,
  onToggleSelection,
  onToggleSelectAll,
}) => {
  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <AlertOctagon size={18} className="text-red-600" />;
      case 'high':
        return <AlertCircle size={18} className="text-orange-600" />;
      case 'medium':
        return <AlertTriangle size={18} className="text-yellow-600" />;
      default:
        return <AlertCircle size={18} className="text-blue-600" />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800';
      case 'high':
        return 'bg-orange-100 text-orange-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-blue-100 text-blue-800';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'firing':
        return <AlertCircle size={18} className="text-red-600" />;
      case 'acknowledged':
        return <AlertTriangle size={18} className="text-yellow-600" />;
      case 'resolved':
        return <CheckCircle size={18} className="text-green-600" />;
      default:
        return null;
    }
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;

    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;

    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  };

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      {/* Table Header */}
      <div className="grid grid-cols-12 gap-4 p-4 bg-gray-50 border-b border-gray-200 text-sm font-medium text-gray-700">
        <div className="col-span-1">
          <input
            type="checkbox"
            checked={selectedAlerts.size === alerts.length && alerts.length > 0}
            onChange={onToggleSelectAll}
            className="rounded"
          />
        </div>
        <div className="col-span-1">Status</div>
        <div className="col-span-3">Alert Title</div>
        <div className="col-span-2">Severity</div>
        <div className="col-span-2">Rule</div>
        <div className="col-span-2">Fired</div>
        <div className="col-span-1">Action</div>
      </div>

      {/* Table Body */}
      {alerts.map((alert) => (
        <div
          key={alert.id}
          className="grid grid-cols-12 gap-4 p-4 border-b border-gray-200 hover:bg-gray-50 items-center"
        >
          <div className="col-span-1">
            <input
              type="checkbox"
              checked={selectedAlerts.has(alert.id)}
              onChange={() => onToggleSelection(alert.id)}
              className="rounded"
            />
          </div>

          <div className="col-span-1 flex justify-center">
            {getStatusIcon(alert.status)}
          </div>

          <div className="col-span-3">
            <button
              onClick={() => onSelectAlert(alert.id)}
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              {alert.title}
            </button>
            {alert.description && (
              <p className="text-xs text-gray-600 mt-1 truncate">
                {alert.description}
              </p>
            )}
          </div>

          <div className="col-span-2">
            <div className="flex items-center gap-2">
              {getSeverityIcon(alert.severity)}
              <span
                className={`px-2 py-1 rounded text-xs font-medium ${getSeverityColor(
                  alert.severity
                )}`}
              >
                {alert.severity}
              </span>
            </div>
          </div>

          <div className="col-span-2">
            <p className="text-sm text-gray-600 truncate">{alert.alert_rule_name}</p>
            {alert.metric_name && (
              <p className="text-xs text-gray-500">{alert.metric_name}</p>
            )}
          </div>

          <div className="col-span-2">
            <p className="text-sm text-gray-600">{formatTime(alert.fired_at)}</p>
            <p className="text-xs text-gray-500">
              {new Date(alert.fired_at).toLocaleDateString()}
            </p>
          </div>

          <div className="col-span-1">
            <button
              onClick={() => onSelectAlert(alert.id)}
              className="text-blue-600 hover:text-blue-700 font-medium text-sm"
            >
              View
            </button>
          </div>
        </div>
      ))}
    </div>
  );
};

export default AlertsTable;
