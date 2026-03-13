import { describe, it, expect, beforeEach, vi } from 'vitest';
import * as alertsService from './alerts';

// Mock fetch
global.fetch = vi.fn();

describe('Alerts Service', () => {
  const mockFetch = global.fetch as any;

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    localStorage.setItem('auth_token', 'test-token');
  });

  describe('createSilence', () => {
    it('sends correct API call to create silence', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'silence-1',
          alert_rule_id: 'rule-1',
          expires_at: '2025-12-31T10:00:00Z',
          reason: 'Test silence',
        }),
      });

      const result = await alertsService.createSilence('rule-1', {
        duration: 60,
        reason: 'Test silence',
        silenceType: 'rule',
      });

      expect(result.id).toBe('silence-1');
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/alerts/rule-1/silence',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });

    it('uses default reason if not provided', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'silence-1',
          alert_rule_id: 'rule-1',
          expires_at: '2025-12-31T10:00:00Z',
          reason: 'Temporarily silenced',
        }),
      });

      await alertsService.createSilence('rule-1', {
        duration: 60,
      });

      const callBody = JSON.parse(mockFetch.mock.calls[0][1].body);
      expect(callBody.reason).toBe('Temporarily silenced');
    });
  });

  describe('createEscalationPolicy', () => {
    it('sends correct API call to create escalation policy', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'policy-1',
          name: 'Test Policy',
          is_active: true,
          steps: [],
          created_at: '2025-03-13T10:00:00Z',
        }),
      });

      const policy = {
        name: 'Test Policy',
        description: 'A test policy',
        steps: [
          {
            step_number: 1,
            wait_minutes: 0,
            notification_channel: 'slack',
            channel_config: { channel_id: 'C123' },
          },
        ],
      };

      const result = await alertsService.createEscalationPolicy(policy);

      expect(result.id).toBe('policy-1');
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/escalation-policies',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });
  });

  describe('updateEscalationPolicy', () => {
    it('sends correct API call to update escalation policy', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'policy-1',
          name: 'Updated Policy',
          is_active: true,
          steps: [],
          updated_at: '2025-03-13T10:00:00Z',
        }),
      });

      const result = await alertsService.updateEscalationPolicy('policy-1', {
        name: 'Updated Policy',
      });

      expect(result.name).toBe('Updated Policy');
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/escalation-policies/policy-1',
        expect.objectContaining({
          method: 'PUT',
        })
      );
    });
  });

  describe('acknowledgeAlert', () => {
    it('sends correct API call to acknowledge alert', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'alert-1',
          acknowledged: true,
          acknowledged_at: '2025-03-13T10:00:00Z',
        }),
      });

      const result = await alertsService.acknowledgeAlert('alert-1', {
        note: 'Investigating',
      });

      expect(result.acknowledged).toBe(true);
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/alerts/alert-1/acknowledge',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });

    it('uses default note if not provided', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'alert-1',
          acknowledged: true,
          acknowledged_at: '2025-03-13T10:00:00Z',
        }),
      });

      await alertsService.acknowledgeAlert('alert-1');

      const callBody = JSON.parse(mockFetch.mock.calls[0][1].body);
      expect(callBody.note).toBe('Acknowledged');
    });
  });

  describe('getEscalationPolicies', () => {
    it('sends correct API call to get escalation policies', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          policies: [
            {
              id: 'policy-1',
              name: 'Policy 1',
              is_active: true,
              steps: [],
              created_at: '2025-03-13T10:00:00Z',
              updated_at: '2025-03-13T10:00:00Z',
            },
          ],
          total: 1,
        }),
      });

      const result = await alertsService.getEscalationPolicies();

      expect(result.policies.length).toBe(1);
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/escalation-policies',
        expect.any(Object)
      );
    });

    it('filters active policies when active_only is true', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          policies: [],
          total: 0,
        }),
      });

      await alertsService.getEscalationPolicies({ active_only: true });

      const callUrl = mockFetch.mock.calls[0][0];
      expect(callUrl).toContain('active_only=true');
    });
  });

  describe('getEscalationPolicy', () => {
    it('sends correct API call to get specific escalation policy', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 'policy-1',
          name: 'Policy 1',
          is_active: true,
          steps: [],
          created_at: '2025-03-13T10:00:00Z',
          updated_at: '2025-03-13T10:00:00Z',
        }),
      });

      const result = await alertsService.getEscalationPolicy('policy-1');

      expect(result.id).toBe('policy-1');
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/escalation-policies/policy-1',
        expect.any(Object)
      );
    });
  });

  describe('deleteEscalationPolicy', () => {
    it('sends correct API call to delete escalation policy', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      await alertsService.deleteEscalationPolicy('policy-1');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/escalation-policies/policy-1',
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });
  });

  describe('linkEscalationPolicy', () => {
    it('sends correct API call to link policy to rule', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          rule_id: 'rule-1',
          policy_id: 'policy-1',
          linked_at: '2025-03-13T10:00:00Z',
        }),
      });

      const result = await alertsService.linkEscalationPolicy('rule-1', 'policy-1');

      expect(result.rule_id).toBe('rule-1');
      expect(result.policy_id).toBe('policy-1');
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/alert-rules/rule-1/escalation-policies',
        expect.objectContaining({
          method: 'POST',
        })
      );
    });
  });

  describe('getSilences', () => {
    it('sends correct API call to get silences', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          silences: [],
          total: 0,
        }),
      });

      await alertsService.getSilences('rule-1');

      const callUrl = mockFetch.mock.calls[0][0];
      expect(callUrl).toContain('alert_rule_id=rule-1');
    });
  });

  describe('deleteSilence', () => {
    it('sends correct API call to delete silence', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      await alertsService.deleteSilence('silence-1');

      expect(mockFetch).toHaveBeenCalledWith(
        '/api/v1/alert-silences/silence-1',
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });
  });

  describe('getAlertAcknowledgments', () => {
    it('sends correct API call to get acknowledgments', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          acknowledgments: [],
          total: 0,
        }),
      });

      await alertsService.getAlertAcknowledgments('alert-1');

      const callUrl = mockFetch.mock.calls[0][0];
      expect(callUrl).toContain('/api/v1/alerts/alert-1/acknowledgments');
    });
  });

  describe('Error handling', () => {
    it('throws error when API call fails', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: async () => ({ message: 'Not found' }),
      });

      await expect(
        alertsService.createSilence('rule-1', { duration: 60 })
      ).rejects.toThrow('Not found');
    });

    it('includes authorization header in requests', async () => {
      localStorage.setItem('auth_token', 'my-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          policies: [],
          total: 0,
        }),
      });

      await alertsService.getEscalationPolicies();

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.Authorization).toBe('Bearer my-token');
    });
  });
});
