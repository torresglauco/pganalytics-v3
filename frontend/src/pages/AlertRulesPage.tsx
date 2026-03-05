import React, { useState, useEffect } from 'react';
import { Plus, Filter, Download, Upload, AlertCircle, Search } from 'lucide-react';
import type { AlertRule } from '../types/alertRules';
import { listAlertRules, exportRules, importRules, bulkRuleAction } from '../api/alertRulesApi';
import AlertRuleForm from '../components/AlertRuleForm';
import RuleDetailsModal from '../components/RuleDetailsModal';
import BulkRuleActions from '../components/BulkRuleActions';

interface AlertRulesPageProps {
  databaseId: string;
  onClose?: () => void;
}

export const AlertRulesPage: React.FC<AlertRulesPageProps> = ({
  databaseId,
  onClose,
}) => {
  const [rules, setRules] = useState<AlertRule[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // UI State
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [selectedRuleId, setSelectedRuleId] = useState<string | null>(null);
  const [selectedRules, setSelectedRules] = useState<Set<string>>(new Set());

  // Filters
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [severityFilter, setSeverityFilter] = useState<string>('');
  const [showFilters, setShowFilters] = useState(false);

  // Pagination
  const [limit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);

  /**
   * Load alert rules
   */
  useEffect(() => {
    loadRules();
  }, [databaseId, searchTerm, statusFilter, severityFilter, offset]);

  const loadRules = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await listAlertRules(databaseId, {
        search: searchTerm || undefined,
        status: statusFilter || undefined,
        severity: severityFilter || undefined,
        limit,
        offset,
      });
      setRules(response.rules);
      setTotal(response.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load rules');
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Handle rule creation/update
   */
  const handleRuleCreated = (rule: AlertRule) => {
    setRules([rule, ...rules]);
    setShowCreateForm(false);
  };

  const handleRuleUpdated = (rule: AlertRule) => {
    setRules(rules.map((r) => (r.id === rule.id ? rule : r)));
    setSelectedRuleId(null);
  };

  const handleRuleDeleted = (ruleId: string) => {
    setRules(rules.filter((r) => r.id !== ruleId));
    setSelectedRuleId(null);
  };

  /**
   * Handle bulk actions
   */
  const handleBulkAction = async (action: string, value?: any) => {
    const ruleIds = Array.from(selectedRules);
    if (ruleIds.length === 0) return;

    try {
      await bulkRuleAction({
        action: action as any,
        rule_ids: ruleIds,
        new_value: value,
      });
      loadRules();
      setSelectedRules(new Set());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Bulk action failed');
    }
  };

  /**
   * Handle export
   */
  const handleExport = async (format: 'json' | 'csv') => {
    try {
      const blob = await exportRules(databaseId, format);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `alert-rules-${new Date().toISOString().split('T')[0]}.${format}`;
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
      const result = await importRules(databaseId, file);
      await loadRules();
      setError(
        `Imported ${result.imported} rules${result.skipped ? `, skipped ${result.skipped}` : ''}`
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed');
    }
  };

  /**
   * Toggle rule selection
   */
  const toggleRuleSelection = (ruleId: string) => {
    const newSelected = new Set(selectedRules);
    if (newSelected.has(ruleId)) {
      newSelected.delete(ruleId);
    } else {
      newSelected.add(ruleId);
    }
    setSelectedRules(newSelected);
  };

  /**
   * Select all rules
   */
  const toggleSelectAll = () => {
    if (selectedRules.size === rules.length) {
      setSelectedRules(new Set());
    } else {
      setSelectedRules(new Set(rules.map((r) => r.id)));
    }
  };

  if (showCreateForm) {
    return (
      <AlertRuleForm
        databaseId={databaseId}
        onCreated={handleRuleCreated}
        onCancel={() => setShowCreateForm(false)}
      />
    );
  }

  if (selectedRuleId) {
    const rule = rules.find((r) => r.id === selectedRuleId);
    if (rule) {
      return (
        <RuleDetailsModal
          rule={rule}
          onClose={() => setSelectedRuleId(null)}
          onUpdated={handleRuleUpdated}
          onDeleted={handleRuleDeleted}
        />
      );
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Alert Rules</h2>
          <p className="text-sm text-gray-600 mt-1">
            Create and manage alert rules for {databaseId}
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => handleExport('json')}
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
              accept=".json,.csv"
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
            New Rule
          </button>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 flex gap-2">
          <AlertCircle size={20} className="flex-shrink-0 mt-0.5" />
          <div className="flex-1">{error}</div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg border border-gray-200 p-4 space-y-4">
        {/* Search and Filter Toggle */}
        <div className="flex gap-2">
          <div className="flex-1 relative">
            <Search
              size={18}
              className="absolute left-3 top-2.5 text-gray-400"
            />
            <input
              type="text"
              placeholder="Search rules..."
              value={searchTerm}
              onChange={(e) => {
                setSearchTerm(e.target.value);
                setOffset(0);
              }}
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

        {/* Filter Options */}
        {showFilters && (
          <div className="grid grid-cols-3 gap-4 pt-4 border-t border-gray-200">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Status
              </label>
              <select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setOffset(0);
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All</option>
                <option value="enabled">Enabled</option>
                <option value="disabled">Disabled</option>
                <option value="testing">Testing</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Severity
              </label>
              <select
                value={severityFilter}
                onChange={(e) => {
                  setSeverityFilter(e.target.value);
                  setOffset(0);
                }}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
          </div>
        )}
      </div>

      {/* Bulk Actions */}
      {selectedRules.size > 0 && (
        <BulkRuleActions
          selectedCount={selectedRules.size}
          onAction={handleBulkAction}
        />
      )}

      {/* Rules Table */}
      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-gray-600">
            Loading rules...
          </div>
        ) : rules.length === 0 ? (
          <div className="p-8 text-center text-gray-600">
            <AlertCircle size={32} className="mx-auto mb-2 opacity-50" />
            <p>No alert rules created yet</p>
            <button
              onClick={() => setShowCreateForm(true)}
              className="mt-4 text-blue-600 hover:text-blue-700 font-medium"
            >
              Create your first rule
            </button>
          </div>
        ) : (
          <>
            {/* Table Header */}
            <div className="grid grid-cols-12 gap-4 p-4 bg-gray-50 border-b border-gray-200 text-sm font-medium text-gray-700">
              <div className="col-span-1">
                <input
                  type="checkbox"
                  checked={selectedRules.size === rules.length && rules.length > 0}
                  onChange={toggleSelectAll}
                  className="rounded"
                />
              </div>
              <div className="col-span-3">Name</div>
              <div className="col-span-2">Severity</div>
              <div className="col-span-2">Status</div>
              <div className="col-span-2">Type</div>
              <div className="col-span-2">Actions</div>
            </div>

            {/* Table Body */}
            {rules.map((rule) => (
              <div
                key={rule.id}
                className="grid grid-cols-12 gap-4 p-4 border-b border-gray-200 hover:bg-gray-50 items-center"
              >
                <div className="col-span-1">
                  <input
                    type="checkbox"
                    checked={selectedRules.has(rule.id)}
                    onChange={() => toggleRuleSelection(rule.id)}
                    className="rounded"
                  />
                </div>
                <div className="col-span-3">
                  <button
                    onClick={() => setSelectedRuleId(rule.id)}
                    className="text-blue-600 hover:text-blue-700 font-medium"
                  >
                    {rule.name}
                  </button>
                  {rule.description && (
                    <p className="text-sm text-gray-600 truncate">
                      {rule.description}
                    </p>
                  )}
                </div>
                <div className="col-span-2">
                  <span
                    className={`px-2 py-1 rounded text-sm font-medium ${
                      rule.severity === 'critical'
                        ? 'bg-red-100 text-red-800'
                        : rule.severity === 'high'
                        ? 'bg-orange-100 text-orange-800'
                        : rule.severity === 'medium'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-blue-100 text-blue-800'
                    }`}
                  >
                    {rule.severity}
                  </span>
                </div>
                <div className="col-span-2">
                  <span
                    className={`px-2 py-1 rounded text-sm font-medium ${
                      rule.status === 'enabled'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {rule.status}
                  </span>
                </div>
                <div className="col-span-2 text-sm text-gray-600">
                  {rule.condition.type}
                </div>
                <div className="col-span-2">
                  <button
                    onClick={() => setSelectedRuleId(rule.id)}
                    className="text-blue-600 hover:text-blue-700 font-medium text-sm"
                  >
                    View
                  </button>
                </div>
              </div>
            ))}
          </>
        )}
      </div>

      {/* Pagination */}
      {total > limit && (
        <div className="flex justify-between items-center">
          <p className="text-sm text-gray-600">
            Showing {offset + 1}-{Math.min(offset + limit, total)} of {total} rules
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setOffset(Math.max(0, offset - limit))}
              disabled={offset === 0}
              className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
            >
              Previous
            </button>
            <button
              onClick={() =>
                setOffset(
                  Math.min(offset + limit, Math.max(0, total - limit))
                )
              }
              disabled={offset + limit >= total}
              className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default AlertRulesPage;
