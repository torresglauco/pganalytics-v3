import React, { useState, useEffect } from 'react';
import { X, Edit2, Trash2, CheckCircle, AlertCircle, Clock } from 'lucide-react';
import type { AlertRule } from '../types/alertRules';
import {
  getAlertRule,
  deleteAlertRule,
  getRuleStats,
  getRuleHistory,
  toggleAlertRule,
} from '../api/alertRulesApi';

interface RuleDetailsModalProps {
  rule: AlertRule;
  onClose: () => void;
  onUpdated?: (rule: AlertRule) => void;
  onDeleted?: (ruleId: string) => void;
}

export const RuleDetailsModal: React.FC<RuleDetailsModalProps> = ({
  rule,
  onClose,
  onUpdated,
  onDeleted,
}) => {
  const [currentRule, setCurrentRule] = useState(rule);
  const [stats, setStats] = useState<any>(null);
  const [history, setHistory] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'details' | 'history' | 'stats'>(
    'details'
  );

  /**
   * Load rule details and stats
   */
  useEffect(() => {
    loadRuleData();
  }, [rule.id]);

  const loadRuleData = async () => {
    try {
      setIsLoading(true);
      const [ruleData, statsData, historyData] = await Promise.all([
        getAlertRule(rule.id),
        getRuleStats(rule.id),
        getRuleHistory(rule.id, { limit: 10 }),
      ]);
      setCurrentRule(ruleData);
      setStats(statsData);
      setHistory(historyData.events || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load rule data');
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Toggle rule enabled/disabled
   */
  const handleToggle = async () => {
    try {
      setIsLoading(true);
      const updated = await toggleAlertRule(
        rule.id,
        currentRule.status === 'enabled' ? false : true
      );
      setCurrentRule(updated);
      onUpdated?.(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update rule');
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Delete rule
   */
  const handleDelete = async () => {
    if (!window.confirm('Delete this rule? This cannot be undone.')) {
      return;
    }

    try {
      setIsDeleting(true);
      await deleteAlertRule(rule.id);
      onDeleted?.(rule.id);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete rule');
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
            <h2 className="text-2xl font-bold text-gray-900">{currentRule.name}</h2>
            {currentRule.description && (
              <p className="text-sm text-gray-600 mt-1">{currentRule.description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <X size={24} />
          </button>
        </div>

        {/* Error */}
        {error && (
          <div className="mx-6 mt-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex gap-2">
            <AlertCircle size={16} className="flex-shrink-0 mt-0.5" />
            <div>{error}</div>
          </div>
        )}

        {/* Status Bar */}
        <div className="px-6 py-3 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
          <div className="flex gap-4">
            <div>
              <span className="text-xs text-gray-600 uppercase">Status</span>
              <p className="text-sm font-medium text-gray-900 capitalize">
                {currentRule.status}
              </p>
            </div>
            <div>
              <span className="text-xs text-gray-600 uppercase">Severity</span>
              <p className={`text-sm font-medium capitalize ${
                currentRule.severity === 'critical'
                  ? 'text-red-700'
                  : currentRule.severity === 'high'
                  ? 'text-orange-700'
                  : currentRule.severity === 'medium'
                  ? 'text-yellow-700'
                  : 'text-blue-700'
              }`}>
                {currentRule.severity}
              </p>
            </div>
            {currentRule.last_fired_at && (
              <div>
                <span className="text-xs text-gray-600 uppercase">Last Fired</span>
                <p className="text-sm font-medium text-gray-900">
                  {new Date(currentRule.last_fired_at).toLocaleDateString()}
                </p>
              </div>
            )}
          </div>

          <div className="flex gap-2">
            <button
              onClick={handleToggle}
              disabled={isLoading}
              className={`px-4 py-2 font-medium rounded-lg transition ${
                currentRule.status === 'enabled'
                  ? 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200'
                  : 'bg-green-100 text-green-800 hover:bg-green-200'
              }`}
            >
              {currentRule.status === 'enabled' ? 'Disable' : 'Enable'}
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
          {(['details', 'history', 'stats'] as const).map((tab) => (
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
              {/* Condition */}
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-3">
                  Condition
                </h3>
                <div className="bg-gray-50 p-4 rounded-lg space-y-2 font-mono text-sm">
                  <p>
                    <span className="text-gray-600">Type:</span>{' '}
                    <span className="font-medium text-gray-900">
                      {currentRule.condition.type}
                    </span>
                  </p>
                  {currentRule.condition.type === 'threshold' && (
                    <>
                      <p>
                        <span className="text-gray-600">Metric:</span>{' '}
                        <span className="font-medium">
                          {(currentRule.condition as any).metric_name}
                        </span>
                      </p>
                      <p>
                        <span className="text-gray-600">Operator:</span>{' '}
                        <span className="font-medium">
                          {(currentRule.condition as any).operator}
                        </span>
                      </p>
                      <p>
                        <span className="text-gray-600">Threshold:</span>{' '}
                        <span className="font-medium">
                          {(currentRule.condition as any).threshold_value}
                        </span>
                      </p>
                    </>
                  )}
                </div>
              </div>

              {/* Metadata */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-1">
                    Created
                  </h4>
                  <p className="text-gray-900">
                    {new Date(currentRule.created_at).toLocaleString()}
                  </p>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-1">
                    Updated
                  </h4>
                  <p className="text-gray-900">
                    {new Date(currentRule.updated_at).toLocaleString()}
                  </p>
                </div>
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-1">
                    Created By
                  </h4>
                  <p className="text-gray-900">{currentRule.created_by}</p>
                </div>
                {currentRule.runbook_url && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-1">
                      Runbook
                    </h4>
                    <a
                      href={currentRule.runbook_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:text-blue-700"
                    >
                      View Runbook
                    </a>
                  </div>
                )}
              </div>

              {/* Tags */}
              {currentRule.tags && currentRule.tags.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Tags</h4>
                  <div className="flex flex-wrap gap-2">
                    {currentRule.tags.map((tag) => (
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

          {activeTab === 'history' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-3">
                Recent Events
              </h3>
              {history.length === 0 ? (
                <p className="text-gray-600">No events recorded yet</p>
              ) : (
                <div className="space-y-3">
                  {history.map((event, idx) => (
                    <div
                      key={idx}
                      className="flex gap-3 p-3 bg-gray-50 rounded-lg"
                    >
                      {event.event_type === 'fired' ? (
                        <AlertCircle className="text-red-600 flex-shrink-0" size={20} />
                      ) : (
                        <CheckCircle className="text-green-600 flex-shrink-0" size={20} />
                      )}
                      <div className="flex-1">
                        <p className="font-medium text-gray-900 capitalize">
                          {event.event_type}
                        </p>
                        <p className="text-sm text-gray-600">
                          {new Date(event.timestamp).toLocaleString()}
                        </p>
                      </div>
                      <span className="text-right text-gray-900">
                        {event.metric_value.toFixed(2)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'stats' && stats && (
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-blue-50 p-4 rounded-lg">
                <p className="text-xs text-gray-600 uppercase">Evaluations</p>
                <p className="text-3xl font-bold text-blue-600">
                  {stats.evaluation_count}
                </p>
              </div>
              <div className="bg-red-50 p-4 rounded-lg">
                <p className="text-xs text-gray-600 uppercase">Times Fired</p>
                <p className="text-3xl font-bold text-red-600">{stats.fire_count}</p>
              </div>
              <div className="bg-gray-50 p-4 rounded-lg col-span-2">
                <p className="text-xs text-gray-600 uppercase mb-1">
                  Avg Evaluation Time
                </p>
                <p className="text-2xl font-bold text-gray-900">
                  {stats.avg_evaluation_time_ms.toFixed(2)}ms
                </p>
              </div>
              {stats.last_error && (
                <div className="bg-yellow-50 p-4 rounded-lg col-span-2">
                  <p className="text-xs text-gray-600 uppercase mb-1">Last Error</p>
                  <p className="text-sm text-yellow-800">{stats.last_error}</p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default RuleDetailsModal;
