import React from 'react';

export interface TimelineStep {
  step_number: number;
  wait_minutes: number;
  notification_channel: string;
  requires_acknowledgment?: boolean;
}

interface EscalationTimelineProps {
  steps: TimelineStep[];
}

const getChannelIcon = (channelType: string): string => {
  switch (channelType.toLowerCase()) {
    case 'slack':
      return '💬';
    case 'pagerduty':
      return '📱';
    case 'email':
      return '📧';
    case 'sms':
      return '📞';
    case 'webhook':
      return '🔗';
    default:
      return '📬';
  }
};

const formatDelay = (minutes: number): string => {
  if (minutes === 0) return 'Now';
  if (minutes < 60) return `+${minutes}m`;
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  if (mins === 0) return `+${hours}h`;
  return `+${hours}h ${mins}m`;
};

export const EscalationTimeline: React.FC<EscalationTimelineProps> = ({ steps }) => {
  if (steps.length === 0) {
    return (
      <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        <p className="text-sm text-gray-500 dark:text-gray-400 text-center">
          No escalation steps configured
        </p>
      </div>
    );
  }

  return (
    <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <div className="flex items-center justify-between overflow-x-auto pb-2">
        {steps.map((step, index) => (
          <div key={step.step_number} className="flex items-center min-w-max">
            {/* Step Card */}
            <div className="flex flex-col items-center">
              <div className="w-16 h-16 rounded-lg bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 flex items-center justify-center shadow-sm">
                <div className="text-center">
                  <div className="text-xl">{getChannelIcon(step.notification_channel)}</div>
                  <div className="text-xs font-medium text-gray-700 dark:text-gray-300 mt-1">
                    {formatDelay(step.wait_minutes)}
                  </div>
                </div>
              </div>
              <div className="text-xs text-gray-600 dark:text-gray-400 mt-2 font-medium">
                {step.notification_channel.charAt(0).toUpperCase() +
                  step.notification_channel.slice(1)}
              </div>
              {step.requires_acknowledgment && (
                <div className="text-xs text-blue-600 dark:text-blue-400 mt-1 flex items-center gap-1">
                  <span>✓</span>
                  <span>Requires Ack</span>
                </div>
              )}
            </div>

            {/* Arrow to next step */}
            {index < steps.length - 1 && (
              <div className="mx-2 text-gray-400 dark:text-gray-600">
                <svg
                  className="w-6 h-6"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
