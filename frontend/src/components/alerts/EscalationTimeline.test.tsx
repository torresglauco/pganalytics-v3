import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { EscalationTimeline, type TimelineStep } from './EscalationTimeline';

describe('EscalationTimeline', () => {
  it('renders empty state when no steps provided', () => {
    render(<EscalationTimeline steps={[]} />);

    expect(screen.getByText('No escalation steps configured')).toBeInTheDocument();
  });

  it('renders all steps', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
      },
      {
        step_number: 2,
        wait_minutes: 5,
        notification_channel: 'email',
      },
      {
        step_number: 3,
        wait_minutes: 15,
        notification_channel: 'pagerduty',
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.getByText('Slack')).toBeInTheDocument();
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('Pagerduty')).toBeInTheDocument();
  });

  it('formats delays correctly', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
      },
      {
        step_number: 2,
        wait_minutes: 5,
        notification_channel: 'email',
      },
      {
        step_number: 3,
        wait_minutes: 120,
        notification_channel: 'pagerduty',
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.getByText('Now')).toBeInTheDocument();
    expect(screen.getByText('+5m')).toBeInTheDocument();
    expect(screen.getByText('+2h')).toBeInTheDocument();
  });

  it('displays channel names capitalized', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
      },
      {
        step_number: 2,
        wait_minutes: 5,
        notification_channel: 'email',
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.getByText('Slack')).toBeInTheDocument();
    expect(screen.getByText('Email')).toBeInTheDocument();
  });

  it('shows acknowledgment badge when requires_acknowledgment is true', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
        requires_acknowledgment: true,
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.getByText('Requires Ack')).toBeInTheDocument();
  });

  it('does not show acknowledgment badge when requires_acknowledgment is false', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
        requires_acknowledgment: false,
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.queryByText('Requires Ack')).not.toBeInTheDocument();
  });

  it('renders arrows between steps', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
      },
      {
        step_number: 2,
        wait_minutes: 5,
        notification_channel: 'email',
      },
    ];

    const { container } = render(<EscalationTimeline steps={steps} />);

    // Check for SVG arrows between steps
    const svgs = container.querySelectorAll('svg');
    expect(svgs.length).toBeGreaterThan(0);
  });

  it('handles single step', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.getByText('Slack')).toBeInTheDocument();
    expect(screen.getByText('Now')).toBeInTheDocument();
  });

  it('handles multiple steps with mixed delays', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
      },
      {
        step_number: 2,
        wait_minutes: 30,
        notification_channel: 'email',
      },
      {
        step_number: 3,
        wait_minutes: 90,
        notification_channel: 'pagerduty',
      },
      {
        step_number: 4,
        wait_minutes: 180,
        notification_channel: 'sms',
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.getByText('+30m')).toBeInTheDocument();
    expect(screen.getByText('+1h 30m')).toBeInTheDocument();
    expect(screen.getByText('+3h')).toBeInTheDocument();
  });

  it('displays all channel types correctly', () => {
    const steps: TimelineStep[] = [
      {
        step_number: 1,
        wait_minutes: 0,
        notification_channel: 'slack',
      },
      {
        step_number: 2,
        wait_minutes: 5,
        notification_channel: 'email',
      },
      {
        step_number: 3,
        wait_minutes: 10,
        notification_channel: 'pagerduty',
      },
      {
        step_number: 4,
        wait_minutes: 15,
        notification_channel: 'sms',
      },
      {
        step_number: 5,
        wait_minutes: 20,
        notification_channel: 'webhook',
      },
    ];

    render(<EscalationTimeline steps={steps} />);

    expect(screen.getByText('Slack')).toBeInTheDocument();
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('Pagerduty')).toBeInTheDocument();
    expect(screen.getByText('Sms')).toBeInTheDocument();
    expect(screen.getByText('Webhook')).toBeInTheDocument();
  });
});
