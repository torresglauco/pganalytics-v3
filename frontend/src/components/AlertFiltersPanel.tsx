import React, { useState } from 'react';
import { X } from 'lucide-react';
import type { AlertFilters } from '../types/alertDashboard';

interface AlertFiltersPanelProps {
  filters: AlertFilters;
  onChange: (filters: AlertFilters) => void;
}

export const AlertFiltersPanel: React.FC<AlertFiltersPanelProps> = ({
  filters,
  onChange,
}) => {
  const [startDate, setStartDate] = useState(
    filters.date_range?.start || ''
  );
  const [endDate, setEndDate] = useState(filters.date_range?.end || '');

  const handleStatusChange = (status: any, checked: boolean) => {
    const current = filters.status || [];
    const updated = checked
      ? [...current, status]
      : current.filter((s) => s !== status);

    onChange({
      ...filters,
      status: updated.length > 0 ? updated : undefined,
    });
  };

  const handleSeverityChange = (severity: any, checked: boolean) => {
    const current = filters.severity || [];
    const updated = checked
      ? [...current, severity]
      : current.filter((s) => s !== severity);

    onChange({
      ...filters,
      severity: updated.length > 0 ? updated : undefined,
    });
  };

  const handleSourceChange = (source: any, checked: boolean) => {
    const current = filters.source_type || [];
    const updated = checked
      ? [...current, source]
      : current.filter((s) => s !== source);

    onChange({
      ...filters,
      source_type: updated.length > 0 ? updated : undefined,
    });
  };

  const handleDateRangeChange = () => {
    if (startDate && endDate) {
      onChange({
        ...filters,
        date_range: {
          start: startDate,
          end: endDate,
        },
      });
    }
  };

  const handleClearFilters = () => {
    onChange({});
    setStartDate('');
    setEndDate('');
  };

  return (
    <div className="space-y-4 pt-4 border-t border-gray-200">
      <div className="grid grid-cols-3 gap-6">
        {/* Status Filter */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Status
          </label>
          <div className="space-y-2">
            {['firing', 'acknowledged', 'resolved'].map((status) => (
              <div key={status} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id={`status_${status}`}
                  checked={filters.status?.includes(status as any) || false}
                  onChange={(e) => handleStatusChange(status, e.target.checked)}
                  className="rounded"
                />
                <label
                  htmlFor={`status_${status}`}
                  className="text-sm text-gray-700 capitalize"
                >
                  {status}
                </label>
              </div>
            ))}
          </div>
        </div>

        {/* Severity Filter */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Severity
          </label>
          <div className="space-y-2">
            {['critical', 'high', 'medium', 'low'].map((severity) => (
              <div key={severity} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id={`severity_${severity}`}
                  checked={filters.severity?.includes(severity as any) || false}
                  onChange={(e) => handleSeverityChange(severity, e.target.checked)}
                  className="rounded"
                />
                <label
                  htmlFor={`severity_${severity}`}
                  className="text-sm text-gray-700 capitalize"
                >
                  {severity}
                </label>
              </div>
            ))}
          </div>
        </div>

        {/* Source Type Filter */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Source Type
          </label>
          <div className="space-y-2">
            {['rule', 'anomaly', 'manual', 'integration'].map((source) => (
              <div key={source} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id={`source_${source}`}
                  checked={filters.source_type?.includes(source as any) || false}
                  onChange={(e) => handleSourceChange(source, e.target.checked)}
                  className="rounded"
                />
                <label
                  htmlFor={`source_${source}`}
                  className="text-sm text-gray-700 capitalize"
                >
                  {source}
                </label>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Date Range Filter */}
      <div className="grid grid-cols-2 gap-4 pt-4 border-t border-gray-200">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Start Date
          </label>
          <input
            type="datetime-local"
            value={startDate}
            onChange={(e) => setStartDate(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            End Date
          </label>
          <input
            type="datetime-local"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex justify-between pt-4 border-t border-gray-200">
        <button
          onClick={handleClearFilters}
          className="flex items-center gap-2 px-4 py-2 text-gray-700 hover:text-gray-900 font-medium"
        >
          <X size={18} />
          Clear Filters
        </button>
        <button
          onClick={handleDateRangeChange}
          disabled={!startDate || !endDate}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg"
        >
          Apply Filters
        </button>
      </div>
    </div>
  );
};

export default AlertFiltersPanel;
