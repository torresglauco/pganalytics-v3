import React, { useState, useEffect } from 'react';
import {
  X,
  Cpu,
  HardDrive,
  Network,
  MemoryStick,
  Database,
  RefreshCw,
  Clock,
} from 'lucide-react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import type { HostStatus, HostMetrics, HostInventory, TimeRange } from '../../types/host';
import { hostApi } from '../../api/hostApi';
import { HealthGauge } from '../charts/HealthGauge';
import { formatBytes, formatDateTime } from '../../utils/formatting';

interface HostDetailPanelProps {
  host: HostStatus;
  onClose: () => void;
}

export const HostDetailPanel: React.FC<HostDetailPanelProps> = ({
  host,
  onClose,
}) => {
  const [metrics, setMetrics] = useState<HostMetrics[]>([]);
  const [inventory, setInventory] = useState<HostInventory | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState<TimeRange>('24h');

  useEffect(() => {
    loadHostData();
  }, [host.collector_id, timeRange]);

  const loadHostData = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const [metricsData, inventoryData] = await Promise.all([
        hostApi.getHostMetrics(host.collector_id, timeRange),
        hostApi.getHostInventory(host.collector_id),
      ]);

      setMetrics(metricsData);
      setInventory(inventoryData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load host data');
    } finally {
      setIsLoading(false);
    }
  };

  const formatChartTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Prepare chart data
  const cpuData = metrics.map((m) => ({
    name: formatChartTime(m.time),
    user: m.cpu_user,
    system: m.cpu_system,
    iowait: m.cpu_iowait,
  }));

  const memoryData = metrics.map((m) => ({
    name: formatChartTime(m.time),
    used: m.memory_used_percent,
  }));

  const diskData = metrics.map((m) => ({
    name: formatChartTime(m.time),
    used: m.disk_used_percent,
  }));

  // Get latest metrics
  const latestMetrics = metrics[metrics.length - 1];

  // Calculate health score based on resource utilization
  const calculateHostHealth = (): number => {
    if (!latestMetrics) return 0;

    const cpuScore = Math.max(0, 100 - latestMetrics.cpu_user - latestMetrics.cpu_system);
    const memoryScore = Math.max(0, 100 - latestMetrics.memory_used_percent);
    const diskScore = Math.max(0, 100 - latestMetrics.disk_used_percent);

    return Math.round((cpuScore * 0.4 + memoryScore * 0.35 + diskScore * 0.25));
  };

  const timeRangeOptions: TimeRange[] = ['1h', '24h', '7d', '30d'];

  if (isLoading) {
    return (
      <div className="fixed inset-y-0 right-0 w-full max-w-3xl bg-white shadow-xl border-l border-gray-200 z-50 flex items-center justify-center">
        <div className="text-gray-600">Loading host details...</div>
      </div>
    );
  }

  return (
    <div className="fixed inset-y-0 right-0 w-full max-w-3xl bg-white shadow-xl border-l border-gray-200 z-50 overflow-y-auto">
      {/* Header */}
      <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between z-10">
        <div>
          <h2 className="text-xl font-bold text-gray-900">{host.hostname}</h2>
          <p className="text-sm text-gray-500">Host Details & Metrics</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={loadHostData}
            className="p-2 hover:bg-gray-100 rounded-lg"
            title="Refresh"
          >
            <RefreshCw size={18} className="text-gray-600" />
          </button>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg"
            title="Close"
          >
            <X size={18} className="text-gray-600" />
          </button>
        </div>
      </div>

      {error && (
        <div className="mx-6 mt-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
          {error}
        </div>
      )}

      <div className="p-6 space-y-6">
        {/* Time Range Selector */}
        <div className="flex items-center gap-2">
          <Clock size={18} className="text-gray-500" />
          <span className="text-sm text-gray-600">Time Range:</span>
          <div className="flex gap-1">
            {timeRangeOptions.map((option) => (
              <button
                key={option}
                onClick={() => setTimeRange(option)}
                className={`px-3 py-1 text-sm rounded-lg transition-colors ${
                  timeRange === option
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {option}
              </button>
            ))}
          </div>
        </div>

        {/* Health Score */}
        <div className="flex justify-center py-4">
          <HealthGauge score={calculateHostHealth()} title="Host Health" size="md" />
        </div>

        {/* Metrics Cards */}
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-blue-50 rounded-lg p-4">
            <div className="flex items-center gap-2 mb-2">
              <Cpu size={18} className="text-blue-600" />
              <span className="text-sm font-medium text-gray-600">CPU</span>
            </div>
            <div className="text-2xl font-bold text-gray-900">
              {latestMetrics ? (latestMetrics.cpu_user + latestMetrics.cpu_system).toFixed(1) : '0'}%
            </div>
            <div className="text-xs text-gray-500 mt-1">
              {latestMetrics?.cpu_cores || 0} cores
            </div>
          </div>

          <div className="bg-purple-50 rounded-lg p-4">
            <div className="flex items-center gap-2 mb-2">
              <MemoryStick size={18} className="text-purple-600" />
              <span className="text-sm font-medium text-gray-600">Memory</span>
            </div>
            <div className="text-2xl font-bold text-gray-900">
              {latestMetrics?.memory_used_percent.toFixed(1) || '0'}%
            </div>
            <div className="text-xs text-gray-500 mt-1">
              {latestMetrics ? formatBytes(latestMetrics.memory_used_mb * 1024 * 1024) : '0 GB'} / {latestMetrics ? formatBytes(latestMetrics.memory_total_mb * 1024 * 1024) : '0 GB'}
            </div>
          </div>

          <div className="bg-amber-50 rounded-lg p-4">
            <div className="flex items-center gap-2 mb-2">
              <HardDrive size={18} className="text-amber-600" />
              <span className="text-sm font-medium text-gray-600">Disk</span>
            </div>
            <div className="text-2xl font-bold text-gray-900">
              {latestMetrics?.disk_used_percent.toFixed(1) || '0'}%
            </div>
            <div className="text-xs text-gray-500 mt-1">
              {latestMetrics ? formatBytes(latestMetrics.disk_used_gb * 1024 * 1024 * 1024) : '0 GB'} / {latestMetrics ? formatBytes(latestMetrics.disk_total_gb * 1024 * 1024 * 1024) : '0 GB'}
            </div>
          </div>
        </div>

        {/* CPU Chart */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Cpu size={18} className="text-blue-600" />
            CPU Usage
          </h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={cpuData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="name" stroke="#64748b" style={{ fontSize: '12px' }} />
              <YAxis stroke="#64748b" style={{ fontSize: '12px' }} unit="%" />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#ffffff',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                }}
              />
              <Line type="monotone" dataKey="user" stroke="#3b82f6" name="User" dot={false} />
              <Line type="monotone" dataKey="system" stroke="#f59e0b" name="System" dot={false} />
              <Line type="monotone" dataKey="iowait" stroke="#ef4444" name="I/O Wait" dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Memory Chart */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <MemoryStick size={18} className="text-purple-600" />
            Memory Usage
          </h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={memoryData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="name" stroke="#64748b" style={{ fontSize: '12px' }} />
              <YAxis stroke="#64748b" style={{ fontSize: '12px' }} unit="%" />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#ffffff',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                }}
              />
              <Line type="monotone" dataKey="used" stroke="#8b5cf6" name="Used %" dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Disk Chart */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <HardDrive size={18} className="text-amber-600" />
            Disk Usage
          </h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={diskData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="name" stroke="#64748b" style={{ fontSize: '12px' }} />
              <YAxis stroke="#64748b" style={{ fontSize: '12px' }} unit="%" />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#ffffff',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                }}
              />
              <Line type="monotone" dataKey="used" stroke="#f59e0b" name="Used %" dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Load Averages */}
        {latestMetrics && (
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Load Averages</h3>
            <div className="grid grid-cols-3 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900">
                  {latestMetrics.cpu_load_1m.toFixed(2)}
                </div>
                <div className="text-sm text-gray-500">1 min</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900">
                  {latestMetrics.cpu_load_5m.toFixed(2)}
                </div>
                <div className="text-sm text-gray-500">5 min</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-gray-900">
                  {latestMetrics.cpu_load_15m.toFixed(2)}
                </div>
                <div className="text-sm text-gray-500">15 min</div>
              </div>
            </div>
          </div>
        )}

        {/* Inventory Section */}
        {inventory && (
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <Database size={18} className="text-blue-600" />
              System & PostgreSQL Configuration
            </h3>

            <div className="grid grid-cols-2 gap-6">
              {/* OS Info */}
              <div>
                <h4 className="text-sm font-medium text-gray-500 mb-2">Operating System</h4>
                <div className="space-y-1 text-sm">
                  <p><span className="text-gray-600">Name:</span> <span className="font-medium">{inventory.os_name}</span></p>
                  <p><span className="text-gray-600">Version:</span> <span className="font-medium">{inventory.os_version}</span></p>
                  <p><span className="text-gray-600">Kernel:</span> <span className="font-medium">{inventory.os_kernel}</span></p>
                </div>
              </div>

              {/* Hardware */}
              <div>
                <h4 className="text-sm font-medium text-gray-500 mb-2">Hardware</h4>
                <div className="space-y-1 text-sm">
                  <p><span className="text-gray-600">CPU Model:</span> <span className="font-medium">{inventory.cpu_model}</span></p>
                  <p><span className="text-gray-600">CPU Cores:</span> <span className="font-medium">{inventory.cpu_cores}</span></p>
                  <p><span className="text-gray-600">CPU Speed:</span> <span className="font-medium">{inventory.cpu_mhz.toFixed(0)} MHz</span></p>
                  <p><span className="text-gray-600">Total Memory:</span> <span className="font-medium">{formatBytes(inventory.memory_total_mb * 1024 * 1024)}</span></p>
                  <p><span className="text-gray-600">Total Disk:</span> <span className="font-medium">{formatBytes(inventory.disk_total_gb * 1024 * 1024 * 1024)}</span></p>
                </div>
              </div>

              {/* PostgreSQL Config */}
              <div className="col-span-2">
                <h4 className="text-sm font-medium text-gray-500 mb-2">PostgreSQL Configuration</h4>
                <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
                  <p><span className="text-gray-600">Version:</span> <span className="font-medium">{inventory.postgres_version}</span></p>
                  <p><span className="text-gray-600">Edition:</span> <span className="font-medium">{inventory.postgres_edition}</span></p>
                  <p><span className="text-gray-600">Port:</span> <span className="font-medium">{inventory.postgres_port}</span></p>
                  <p><span className="text-gray-600">Data Directory:</span> <span className="font-medium">{inventory.postgres_data_dir}</span></p>
                  <p><span className="text-gray-600">Max Connections:</span> <span className="font-medium">{inventory.postgres_max_connections}</span></p>
                  <p><span className="text-gray-600">Shared Buffers:</span> <span className="font-medium">{inventory.postgres_shared_buffers_mb} MB</span></p>
                  <p><span className="text-gray-600">Work Mem:</span> <span className="font-medium">{inventory.postgres_work_mem_mb} MB</span></p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Status History */}
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Status Information</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">Current Status:</span>
              <span className={`font-medium ${host.status === 'up' ? 'text-emerald-600' : 'text-red-600'}`}>
                {host.status.toUpperCase()}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">Last Seen:</span>
              <span className="font-medium">{host.last_seen ? formatDateTime(host.last_seen) : 'Never'}</span>
            </div>
            {host.status_changed_at && (
              <div className="flex justify-between">
                <span className="text-gray-600">Status Changed:</span>
                <span className="font-medium">{formatDateTime(host.status_changed_at)}</span>
              </div>
            )}
            <div className="flex justify-between">
              <span className="text-gray-600">Threshold:</span>
              <span className="font-medium">{host.configured_threshold_seconds} seconds</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HostDetailPanel;