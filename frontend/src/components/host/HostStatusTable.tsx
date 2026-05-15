import React from 'react';
import { CheckCircle, XCircle, HelpCircle, Server } from 'lucide-react';
import type { HostStatus } from '../../types/host';
import { getRelativeTime } from '../../utils/formatting';

interface HostStatusTableProps {
  hosts: HostStatus[];
  selectedHostId?: string;
  onSelectHost: (host: HostStatus) => void;
  isLoading?: boolean;
}

export const HostStatusTable: React.FC<HostStatusTableProps> = ({
  hosts,
  selectedHostId,
  onSelectHost,
  isLoading = false,
}) => {
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'up':
        return <CheckCircle size={18} className="text-emerald-600" />;
      case 'down':
        return <XCircle size={18} className="text-red-600" />;
      default:
        return <HelpCircle size={18} className="text-gray-400" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'up':
        return 'bg-emerald-100 text-emerald-800';
      case 'down':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-600';
    }
  };

  const formatUnresponsiveTime = (seconds: number) => {
    if (seconds === 0) return '-';
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ${hours % 24}h`;
    if (hours > 0) return `${hours}h ${minutes % 60}m`;
    return `${minutes}m`;
  };

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-8 text-center">
        <p className="text-gray-600">Loading hosts...</p>
      </div>
    );
  }

  if (hosts.length === 0) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-8 text-center">
        <Server size={32} className="mx-auto mb-2 opacity-50 text-gray-400" />
        <p className="text-gray-600">No hosts found</p>
        <p className="text-sm text-gray-500 mt-2">Hosts will appear here when collectors are configured</p>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      {/* Table Header */}
      <div className="grid grid-cols-12 gap-4 p-4 bg-gray-50 border-b border-gray-200 text-sm font-medium text-gray-700">
        <div className="col-span-1">Status</div>
        <div className="col-span-3">Hostname</div>
        <div className="col-span-2">Health</div>
        <div className="col-span-2">Last Seen</div>
        <div className="col-span-2">Unresponsive</div>
        <div className="col-span-2">Action</div>
      </div>

      {/* Table Body */}
      {hosts.map((host) => (
        <div
          key={host.collector_id}
          className={`grid grid-cols-12 gap-4 p-4 border-b border-gray-200 hover:bg-gray-50 items-center cursor-pointer transition-colors ${
            selectedHostId === host.collector_id ? 'bg-blue-50' : ''
          }`}
          onClick={() => onSelectHost(host)}
        >
          <div className="col-span-1 flex justify-center">
            {getStatusIcon(host.status)}
          </div>

          <div className="col-span-3">
            <p className="font-medium text-gray-900">{host.hostname}</p>
            <p className="text-xs text-gray-500 truncate">{host.collector_id.substring(0, 8)}...</p>
          </div>

          <div className="col-span-2">
            <span
              className={`px-2 py-1 rounded text-xs font-medium ${getStatusBadge(host.status)}`}
            >
              {host.status.toUpperCase()}
            </span>
          </div>

          <div className="col-span-2">
            <p className="text-sm text-gray-600">
              {host.last_seen ? getRelativeTime(host.last_seen) : 'Never'}
            </p>
          </div>

          <div className="col-span-2">
            <p className="text-sm text-gray-600">
              {formatUnresponsiveTime(host.unresponsive_for_seconds)}
            </p>
            {host.unresponsive_for_seconds > 0 && (
              <p className="text-xs text-gray-500">
                threshold: {host.configured_threshold_seconds}s
              </p>
            )}
          </div>

          <div className="col-span-2">
            <button
              onClick={(e) => {
                e.stopPropagation();
                onSelectHost(host);
              }}
              className="text-blue-600 hover:text-blue-700 font-medium text-sm"
            >
              View Details
            </button>
          </div>
        </div>
      ))}
    </div>
  );
};

export default HostStatusTable;