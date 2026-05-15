import React, { useState, useEffect } from 'react';
import { X, Filter, ChevronDown } from 'lucide-react';
import type { ClassificationFilter, PatternType, Category } from '../../types/classification';
import { PATTERN_LABELS, CATEGORY_LABELS } from '../../types/classification';

interface ClassificationFiltersProps {
  filters: ClassificationFilter;
  databases: string[];
  schemas: string[];
  tables: string[];
  onChange: (filters: ClassificationFilter) => void;
  onReset: () => void;
}

export const ClassificationFilters: React.FC<ClassificationFiltersProps> = ({
  filters,
  databases,
  schemas,
  tables,
  onChange,
  onReset,
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [localFilters, setLocalFilters] = useState<ClassificationFilter>(filters);

  useEffect(() => {
    setLocalFilters(filters);
  }, [filters]);

  const handleFilterChange = (key: keyof ClassificationFilter, value: string | undefined) => {
    setLocalFilters((prev) => ({
      ...prev,
      [key]: value || undefined,
    }));
  };

  const handleApply = () => {
    onChange(localFilters);
  };

  const handleReset = () => {
    setLocalFilters({});
    onReset();
  };

  const timeRanges = [
    { value: '1h', label: 'Last Hour' },
    { value: '24h', label: 'Last 24 Hours' },
    { value: '7d', label: 'Last 7 Days' },
    { value: '30d', label: 'Last 30 Days' },
  ];

  const patternTypes: PatternType[] = ['CPF', 'CNPJ', 'EMAIL', 'PHONE', 'CREDIT_CARD', 'CUSTOM'];
  const categories: Category[] = ['PII', 'PCI', 'SENSITIVE', 'CUSTOM'];

  const hasActiveFilters = Object.values(filters).some((v) => v !== undefined && v !== '');

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      {/* Header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-4 py-3 flex items-center justify-between bg-gray-50 hover:bg-gray-100 transition-colors"
      >
        <div className="flex items-center gap-2">
          <Filter size={18} className="text-gray-500" />
          <span className="font-medium text-gray-700">Filters</span>
          {hasActiveFilters && (
            <span className="px-2 py-0.5 bg-blue-100 text-blue-700 text-xs rounded-full">
              Active
            </span>
          )}
        </div>
        <ChevronDown
          size={18}
          className={`text-gray-500 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
        />
      </button>

      {/* Filter Panel */}
      {isExpanded && (
        <div className="p-4 space-y-4 border-t border-gray-200">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Database Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Database
              </label>
              <select
                value={localFilters.database || ''}
                onChange={(e) => handleFilterChange('database', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
              >
                <option value="">All Databases</option>
                {databases.map((db) => (
                  <option key={db} value={db}>
                    {db}
                  </option>
                ))}
              </select>
            </div>

            {/* Schema Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Schema
              </label>
              <select
                value={localFilters.schema || ''}
                onChange={(e) => handleFilterChange('schema', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                disabled={!localFilters.database && schemas.length === 0}
              >
                <option value="">All Schemas</option>
                {schemas.map((schema) => (
                  <option key={schema} value={schema}>
                    {schema}
                  </option>
                ))}
              </select>
            </div>

            {/* Table Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Table
              </label>
              <select
                value={localFilters.table || ''}
                onChange={(e) => handleFilterChange('table', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                disabled={!localFilters.schema && tables.length === 0}
              >
                <option value="">All Tables</option>
                {tables.map((table) => (
                  <option key={table} value={table}>
                    {table}
                  </option>
                ))}
              </select>
            </div>

            {/* Time Range Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Time Range
              </label>
              <select
                value={localFilters.time_range || '24h'}
                onChange={(e) => handleFilterChange('time_range', e.target.value as '1h' | '24h' | '7d' | '30d')}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
              >
                {timeRanges.map((range) => (
                  <option key={range.value} value={range.value}>
                    {range.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Pattern Type Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Pattern Type
              </label>
              <select
                value={localFilters.pattern_type || ''}
                onChange={(e) => handleFilterChange('pattern_type', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
              >
                <option value="">All Patterns</option>
                {patternTypes.map((type) => (
                  <option key={type} value={type}>
                    {PATTERN_LABELS[type]}
                  </option>
                ))}
              </select>
            </div>

            {/* Category Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Category
              </label>
              <select
                value={localFilters.category || ''}
                onChange={(e) => handleFilterChange('category', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
              >
                <option value="">All Categories</option>
                {categories.map((cat) => (
                  <option key={cat} value={cat}>
                    {CATEGORY_LABELS[cat]}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
            <button
              onClick={handleReset}
              className="flex items-center gap-2 px-4 py-2 text-gray-700 hover:text-gray-900 font-medium text-sm"
            >
              <X size={16} />
              Reset
            </button>
            <button
              onClick={handleApply}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg text-sm"
            >
              Apply Filters
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default ClassificationFilters;