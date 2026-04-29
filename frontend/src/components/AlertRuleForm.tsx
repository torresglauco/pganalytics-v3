import React, { useState } from 'react';
import { AlertCircle, X } from 'lucide-react';
import type {
  CreateRuleRequest,
  RuleCondition,
  AlertRule,
} from '../types/alertRules';
import { createAlertRule, validateAlertRule } from '../api/alertRulesApi';
import RuleConditionBuilder from './RuleConditionBuilder';
import RuleTestModal from './RuleTestModal';

interface AlertRuleFormProps {
  databaseId: string;
  initialRule?: AlertRule;
  onCreated?: (rule: AlertRule) => void;
  onUpdated?: (rule: AlertRule) => void;
  onCancel?: () => void;
}

type FormStep = 'basic' | 'condition' | 'notifications' | 'review';

export const AlertRuleForm: React.FC<AlertRuleFormProps> = ({
  databaseId,
  initialRule,
  onCreated,
  onUpdated: _onUpdated,
  onCancel,
}) => {
  // Form state
  const [name, setName] = useState(initialRule?.name || '');
  const [description, setDescription] = useState(initialRule?.description || '');
  const [severity, setSeverity] = useState<any>(
    initialRule?.severity || 'medium'
  );
  const [condition, setCondition] = useState<RuleCondition | null>(
    initialRule?.condition || null
  );
  // Notification channels state removed - will be implemented when notification configuration is added
  const [currentStep, setCurrentStep] = useState<FormStep>('basic');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [showTestModal, setShowTestModal] = useState(false);

  // Tags and metadata
  const [tags, setTags] = useState<string[]>(initialRule?.tags || []);
  const [tagInput, setTagInput] = useState('');
  const [runbookUrl, setRunbookUrl] = useState(initialRule?.runbook_url || '');

  /**
   * Validate current step
   */
  const validateStep = async (): Promise<boolean> => {
    const errors: Record<string, string> = {};

    if (currentStep === 'basic') {
      if (!name.trim()) errors.name = 'Rule name is required';
      if (!severity) errors.severity = 'Severity is required';
    }

    if (currentStep === 'condition') {
      if (!condition) errors.condition = 'Condition is required';
    }

    if (Object.keys(errors).length > 0) {
      setValidationErrors(errors);
      return false;
    }

    return true;
  };

  /**
   * Move to next step
   */
  const handleNextStep = async () => {
    if (!(await validateStep())) return;

    const steps: FormStep[] = ['basic', 'condition', 'notifications', 'review'];
    const currentIndex = steps.indexOf(currentStep);
    if (currentIndex < steps.length - 1) {
      setCurrentStep(steps[currentIndex + 1]);
      setError(null);
    }
  };

  /**
   * Move to previous step
   */
  const handlePreviousStep = () => {
    const steps: FormStep[] = ['basic', 'condition', 'notifications', 'review'];
    const currentIndex = steps.indexOf(currentStep);
    if (currentIndex > 0) {
      setCurrentStep(steps[currentIndex - 1]);
      setError(null);
    }
  };

  /**
   * Add tag
   */
  const handleAddTag = () => {
    if (tagInput.trim() && !tags.includes(tagInput.trim())) {
      setTags([...tags, tagInput.trim()]);
      setTagInput('');
    }
  };

  /**
   * Remove tag
   */
  const handleRemoveTag = (tag: string) => {
    setTags(tags.filter((t) => t !== tag));
  };

  /**
   * Submit form
   */
  const handleSubmit = async () => {
    try {
      setIsLoading(true);
      setError(null);

      if (!name.trim()) {
        setError('Rule name is required');
        return;
      }

      if (!condition) {
        setError('Condition is required');
        return;
      }

      const request: CreateRuleRequest = {
        name: name.trim(),
        description: description.trim() || undefined,
        database_id: databaseId,
        severity,
        condition,
        notifications: [],
        tags: tags.length > 0 ? tags : undefined,
        runbook_url: runbookUrl.trim() || undefined,
      };

      // Validate before submitting
      const validation = await validateAlertRule(request);
      if (!validation.valid) {
        const errorMap: Record<string, string> = {};
        validation.errors.forEach((e) => {
          errorMap[e.field] = e.message;
        });
        setValidationErrors(errorMap);
        setError('Please fix the errors below');
        return;
      }

      const rule = await createAlertRule(request);
      onCreated?.(rule);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create rule');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      {/* Step Indicator */}
      <div className="flex justify-between items-center">
        <div className="flex gap-4">
          {['basic', 'condition', 'notifications', 'review'].map((step, i) => (
            <div key={step} className="flex items-center gap-2">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center font-medium ${
                  currentStep === step
                    ? 'bg-blue-600 text-white'
                    : ['basic', 'condition'].includes(step) && condition
                    ? 'bg-green-100 text-green-800'
                    : 'bg-gray-200 text-gray-600'
                }`}
              >
                {i + 1}
              </div>
              <span className="text-sm font-medium text-gray-700 capitalize">
                {step}
              </span>
              {i < 3 && <div className="w-8 h-1 bg-gray-200" />}
            </div>
          ))}
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 flex gap-2">
          <AlertCircle size={20} className="flex-shrink-0 mt-0.5" />
          <div>{error}</div>
        </div>
      )}

      {/* Step: Basic Info */}
      {currentStep === 'basic' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">
            Basic Information
          </h3>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Rule Name *
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                setValidationErrors({ ...validationErrors, name: '' });
              }}
              placeholder="e.g., High CPU Usage"
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                validationErrors.name ? 'border-red-500' : 'border-gray-300'
              }`}
            />
            {validationErrors.name && (
              <p className="text-sm text-red-600 mt-1">{validationErrors.name}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description of what this rule monitors"
              rows={3}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Severity *
            </label>
            <select
              value={severity}
              onChange={(e) => setSeverity(e.target.value)}
              className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                validationErrors.severity ? 'border-red-500' : 'border-gray-300'
              }`}
            >
              <option value="">Select severity</option>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
            {validationErrors.severity && (
              <p className="text-sm text-red-600 mt-1">
                {validationErrors.severity}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Runbook URL
            </label>
            <input
              type="url"
              value={runbookUrl}
              onChange={(e) => setRunbookUrl(e.target.value)}
              placeholder="https://wiki.example.com/runbooks/..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Tags
            </label>
            <div className="flex gap-2 mb-2">
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    handleAddTag();
                  }
                }}
                placeholder="Add tags..."
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
              <button
                onClick={handleAddTag}
                className="px-3 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg"
              >
                Add
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {tags.map((tag) => (
                <div
                  key={tag}
                  className="bg-blue-100 text-blue-800 px-3 py-1 rounded-full flex items-center gap-2"
                >
                  {tag}
                  <button
                    onClick={() => handleRemoveTag(tag)}
                    className="hover:text-blue-900"
                  >
                    <X size={16} />
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Step: Condition */}
      {currentStep === 'condition' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">
            Alert Condition
          </h3>
          <RuleConditionBuilder
            condition={condition}
            onChange={setCondition}
            databaseId={databaseId}
          />
        </div>
      )}

      {/* Step: Notifications */}
      {currentStep === 'notifications' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">
            Notifications
          </h3>
          <p className="text-sm text-gray-600">
            Configure which channels should receive notifications for this rule.
          </p>
          <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg text-sm text-blue-700">
            Notification channels can be configured from the Settings page.
          </div>
        </div>
      )}

      {/* Step: Review */}
      {currentStep === 'review' && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">
            Review & Test
          </h3>

          <div className="bg-gray-50 p-4 rounded-lg space-y-3">
            <div>
              <p className="text-sm font-medium text-gray-700">Name:</p>
              <p className="text-gray-900">{name}</p>
            </div>
            {description && (
              <div>
                <p className="text-sm font-medium text-gray-700">Description:</p>
                <p className="text-gray-900">{description}</p>
              </div>
            )}
            <div>
              <p className="text-sm font-medium text-gray-700">Severity:</p>
              <p className="text-gray-900 capitalize">{severity}</p>
            </div>
            {condition && (
              <div>
                <p className="text-sm font-medium text-gray-700">Condition:</p>
                <p className="text-gray-900 capitalize">{condition.type}</p>
              </div>
            )}
          </div>

          <button
            onClick={() => setShowTestModal(true)}
            className="w-full py-2 px-4 border border-blue-600 text-blue-600 hover:bg-blue-50 font-medium rounded-lg"
          >
            Test Rule Condition
          </button>

          {showTestModal && condition && (
            <RuleTestModal
              databaseId={databaseId}
              condition={condition}
              onClose={() => setShowTestModal(false)}
            />
          )}
        </div>
      )}

      {/* Navigation Buttons */}
      <div className="flex justify-between">
        <button
          onClick={onCancel || handlePreviousStep}
          className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700 font-medium"
        >
          {currentStep === 'basic' ? 'Cancel' : 'Previous'}
        </button>

        <div className="flex gap-2">
          {currentStep !== 'review' && (
            <button
              onClick={handleNextStep}
              disabled={isLoading}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg"
            >
              Next
            </button>
          )}
          {currentStep === 'review' && (
            <button
              onClick={handleSubmit}
              disabled={isLoading || !name || !condition}
              className="px-6 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white font-medium rounded-lg"
            >
              {isLoading ? 'Creating...' : 'Create Rule'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default AlertRuleForm;
