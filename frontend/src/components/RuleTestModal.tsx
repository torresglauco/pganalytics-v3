import React, { useState } from 'react';
import { AlertCircle, CheckCircle, X } from 'lucide-react';
import type { RuleCondition, RuleTestResult } from '../types/alertRules';
import { testAlertRule } from '../api/alertRulesApi';

interface RuleTestModalProps {
  databaseId: string;
  condition: RuleCondition;
  onClose: () => void;
}

export const RuleTestModal: React.FC<RuleTestModalProps> = ({
  databaseId,
  condition,
  onClose,
}) => {
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<RuleTestResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleTest = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const testResult = await testAlertRule(null, {
        name: 'Test Rule',
        database_id: databaseId,
        severity: 'medium',
        condition,
        notifications: [],
      });
      setResult(testResult);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Test failed');
    } finally {
      setIsLoading(false);
    }
  };

  React.useEffect(() => {
    handleTest();
  }, []);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 p-6 space-y-4">
        <div className="flex justify-between items-start">
          <h3 className="text-lg font-semibold text-gray-900">Test Rule Condition</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X size={24} />
          </button>
        </div>

        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex gap-2">
            <AlertCircle size={16} className="flex-shrink-0 mt-0.5" />
            <div>{error}</div>
          </div>
        )}

        {isLoading ? (
          <div className="text-center py-8">
            <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-blue-200 border-t-blue-600 mb-2"></div>
            <p className="text-gray-600">Testing condition...</p>
          </div>
        ) : result ? (
          <div className="space-y-4">
            {result.condition_met ? (
              <div className="p-4 bg-red-50 border border-red-200 rounded-lg flex gap-3">
                <AlertCircle className="text-red-600 flex-shrink-0" size={20} />
                <div>
                  <h4 className="font-semibold text-red-900">Condition Would Trigger</h4>
                  <p className="text-sm text-red-700 mt-1">
                    This rule condition is currently met and would generate an alert.
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-4 bg-green-50 border border-green-200 rounded-lg flex gap-3">
                <CheckCircle className="text-green-600 flex-shrink-0" size={20} />
                <div>
                  <h4 className="font-semibold text-green-900">Condition Not Met</h4>
                  <p className="text-sm text-green-700 mt-1">
                    This rule condition is not currently met.
                  </p>
                </div>
              </div>
            )}

            <div className="bg-gray-50 p-3 rounded-lg space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">Current Value:</span>
                <span className="font-mono font-medium text-gray-900">
                  {result.metric_value.toFixed(2)}
                </span>
              </div>
              {result.threshold_value !== undefined && (
                <div className="flex justify-between">
                  <span className="text-gray-600">Threshold:</span>
                  <span className="font-mono font-medium text-gray-900">
                    {result.threshold_value.toFixed(2)}
                  </span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-gray-600">Last Data Point:</span>
                <span className="font-mono text-gray-900">
                  {new Date(result.last_data_point.timestamp).toLocaleTimeString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">Evaluation Time:</span>
                <span className="font-mono text-gray-900">{result.evaluation_time_ms}ms</span>
              </div>
            </div>

            {result.sample_metrics && result.sample_metrics.length > 0 && (
              <div>
                <h4 className="font-medium text-gray-900 mb-2">Recent Data Points</h4>
                <div className="space-y-1 max-h-40 overflow-y-auto">
                  {result.sample_metrics.map((point, idx) => (
                    <div key={idx} className="flex justify-between text-sm text-gray-600">
                      <span>{new Date(point.timestamp).toLocaleTimeString()}</span>
                      <span className="font-mono">{point.value.toFixed(2)}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : null}

        <div className="flex gap-2 pt-4 border-t border-gray-200">
          <button
            onClick={handleTest}
            disabled={isLoading}
            className="flex-1 py-2 px-4 border border-blue-600 text-blue-600 hover:bg-blue-50 font-medium rounded-lg disabled:opacity-50"
          >
            Re-test
          </button>
          <button
            onClick={onClose}
            className="flex-1 py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default RuleTestModal;
