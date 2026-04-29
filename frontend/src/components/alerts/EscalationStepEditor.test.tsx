import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { EscalationStepEditor, type EscalationStep } from './EscalationStepEditor';

describe('EscalationStepEditor', () => {
  const mockStep: EscalationStep = {
    step_number: 1,
    wait_minutes: 0,
    notification_channel: 'slack',
    channel_config: { channel_id: 'C123456' },
    requires_acknowledgment: false,
  };

  const mockOnUpdate = vi.fn();
  const mockOnRemove = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the component', () => {
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    expect(screen.getByText('Step 1')).toBeInTheDocument();
    expect(screen.getByText('Remove')).toBeInTheDocument();
  });

  it('renders delay input with default value', () => {
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const delayInput = screen.getByDisplayValue('0') as HTMLInputElement;
    expect(delayInput).toBeInTheDocument();
  });

  it('renders channel type dropdown', () => {
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const selectElement = screen.getByDisplayValue('Slack') as HTMLSelectElement;
    expect(selectElement).toBeInTheDocument();
    expect(screen.getByText('Email')).toBeInTheDocument();
    expect(screen.getByText('Pagerduty')).toBeInTheDocument();
  });

  it('allows delay change', async () => {
    const user = userEvent.setup();
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const delayInput = screen.getByDisplayValue('0') as HTMLInputElement;
    await user.clear(delayInput);
    await user.type(delayInput, '5');

    await waitFor(() => {
      expect(mockOnUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          wait_minutes: 5,
        })
      );
    });
  });

  it('allows channel type change', async () => {
    const user = userEvent.setup();
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const selectElement = screen.getByDisplayValue('Slack') as HTMLSelectElement;
    await user.selectOptions(selectElement, 'email');

    await waitFor(() => {
      expect(mockOnUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          notification_channel: 'email',
        })
      );
    });
  });

  it('renders acknowledgment checkbox', () => {
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const checkbox = screen.getByRole('checkbox') as HTMLInputElement;
    expect(checkbox).toBeInTheDocument();
    expect(checkbox.checked).toBe(false);
  });

  it('allows acknowledgment toggle', async () => {
    const user = userEvent.setup();
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const checkbox = screen.getByRole('checkbox') as HTMLInputElement;
    await user.click(checkbox);

    await waitFor(() => {
      expect(mockOnUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          requires_acknowledgment: true,
        })
      );
    });
  });

  it('calls onRemove when remove button is clicked', async () => {
    const user = userEvent.setup();
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const removeButton = screen.getByRole('button', { name: 'Remove' });
    await user.click(removeButton);

    expect(mockOnRemove).toHaveBeenCalled();
  });

  it('renders slack channel config fields', () => {
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    expect(screen.getByPlaceholderText('e.g., C1234567890')).toBeInTheDocument();
  });

  it('renders email channel config fields when selected', async () => {
    const user = userEvent.setup();
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const selectElement = screen.getByDisplayValue('Slack') as HTMLSelectElement;
    await user.selectOptions(selectElement, 'email');

    await waitFor(() => {
      expect(screen.getByPlaceholderText('e.g., team@example.com')).toBeInTheDocument();
    });
  });

  it('renders pagerduty channel config fields when selected', async () => {
    const user = userEvent.setup();
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const selectElement = screen.getByDisplayValue('Slack') as HTMLSelectElement;
    await user.selectOptions(selectElement, 'pagerduty');

    await waitFor(() => {
      expect(screen.getByPlaceholderText('PagerDuty integration key')).toBeInTheDocument();
    });
  });

  it('renders webhook channel config fields when selected', async () => {
    const user = userEvent.setup();
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={1}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const selectElement = screen.getByDisplayValue('Slack') as HTMLSelectElement;
    await user.selectOptions(selectElement, 'webhook');

    await waitFor(() => {
      expect(screen.getByPlaceholderText('e.g., https://example.com/webhook')).toBeInTheDocument();
    });
  });

  it('updates onUpdate callback when delay changes', async () => {
    const user = userEvent.setup();
    const step = { ...mockStep, wait_minutes: 10 };

    render(
      <EscalationStepEditor
        step={step}
        stepNumber={2}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    const delayInput = screen.getByDisplayValue('10') as HTMLInputElement;
    await user.clear(delayInput);
    await user.type(delayInput, '15');

    await waitFor(() => {
      expect(mockOnUpdate).toHaveBeenCalledWith(
        expect.objectContaining({
          wait_minutes: 15,
        })
      );
    });
  });

  it('displays step number in title', () => {
    render(
      <EscalationStepEditor
        step={mockStep}
        stepNumber={3}
        onUpdate={mockOnUpdate}
        onRemove={mockOnRemove}
      />
    );

    expect(screen.getByText('Step 3')).toBeInTheDocument();
  });
});
