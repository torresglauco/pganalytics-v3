import React from 'react';
import { Server, CheckCircle, XCircle, Cpu, MemoryStick, HardDrive } from 'lucide-react';
import type { HostStatus, HostSummaryStats } from '../../types/host';

interface HostInventorySummaryProps {
  hosts: HostStatus[];
  onFilterByStatus?: (status: 'up' | 'down' | 'unknown') => void;
}

export const HostInventorySummary: React.FC<HostInventorySummaryProps> = ({
  hosts,
  onFilterByStatus,
}) => {
  const stats: HostSummaryStats = {
    total_hosts: hosts.length,
    hosts_up: hosts.filter((h) => h.status === 'up').length,
    hosts_down: hosts.filter((h) => h.status === 'down').length,
    avg_cpu_percent: 0,
    avg_memory_percent: 0,
    avg_disk_percent: 0,
  };

  const cards = [
    {
      label: 'Total Hosts',
      value: stats.total_hosts,
      icon: <Server size={24} className="text-blue-600" />,
      color: 'bg-blue-50',
      clickable: false,
    },
    {
      label: 'Hosts Up',
      value: stats.hosts_up,
      icon: <CheckCircle size={24} className="text-emerald-600" />,
      color: 'bg-emerald-50',
      clickable: true,
      filter: 'up' as const,
    },
    {
      label: 'Hosts Down',
      value: stats.hosts_down,
      icon: <XCircle size={24} className="text-red-600" />,
      color: 'bg-red-50',
      clickable: true,
      filter: 'down' as const,
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {cards.map((card) => (
        <div
          key={card.label}
          className={`${card.color} rounded-lg border border-gray-200 p-4 ${
            card.clickable ? 'cursor-pointer hover:shadow-md transition-shadow' : ''
          }`}
          onClick={() => {
            if (card.clickable && card.filter && onFilterByStatus) {
              onFilterByStatus(card.filter);
            }
          }}
        >
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">{card.label}</p>
              <p className="text-3xl font-bold text-gray-900 mt-1">{card.value}</p>
            </div>
            <div className="p-3 bg-white rounded-lg shadow-sm">
              {card.icon}
            </div>
          </div>
          {card.clickable && (
            <p className="text-xs text-gray-500 mt-2">Click to filter</p>
          )}
        </div>
      ))}
    </div>
  );
};

export default HostInventorySummary;