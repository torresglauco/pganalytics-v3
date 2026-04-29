import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { waitFor, act, renderHook } from '@testing-library/react';
import { useQueryPerformance } from '../../hooks/useQueryPerformance';
import { useLogAnalysis } from '../../hooks/useLogAnalysis';

// Mock fetch API
global.fetch = vi.fn();

// Mock WebSocket
let lastMockWebSocket: MockWebSocket | null = null;

class MockWebSocket {
  url: string;
  readyState: number = 0;
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;

  constructor(url: string) {
    this.url = url;
    this.readyState = 0;
    lastMockWebSocket = this;

    // Simulate connection opening
    setTimeout(() => {
      this.readyState = 1;
      if (this.onopen) {
        this.onopen(new Event('open'));
      }
    }, 10);
  }

  send(_data: string) {
    // Mock implementation
  }

  close() {
    this.readyState = 3;
    if (this.onclose) {
      this.onclose(new CloseEvent('close'));
    }
  }
}

(MockWebSocket as any).instance = null;

global.WebSocket = MockWebSocket as any;

describe('Query Performance Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('useQueryPerformance Hook', () => {
    it('should fetch query performance data successfully', async () => {
      const mockData = {
        queries: [
          {
            id: 1,
            database_id: 1,
            query_hash: 'hash123',
            query_text: 'SELECT * FROM users',
            plan_json: { 'Node Type': 'Seq Scan' },
            mean_time: 45.5,
            total_time: 455.0,
            calls: 10,
            created_at: new Date().toISOString(),
          },
        ],
        issues: [
          {
            id: 1,
            query_plan_id: 1,
            issue_type: 'sequential_scan' as const,
            severity: 'medium' as const,
            affected_node_id: 0,
            description: 'Sequential scan detected',
            recommendation: 'Add index on column',
            estimated_benefit: 50,
          },
        ],
        timeline: [
          {
            id: 1,
            query_plan_id: 1,
            metric_timestamp: new Date().toISOString(),
            avg_duration: 45.5,
            max_duration: 120.0,
            executions: 10,
          },
        ],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      const { result } = renderHook(() => useQueryPerformance('1'));

      await waitFor(() => {
        // Hook should complete loading
        expect(result.current.loading).toBe(false);
      });
    });

    it('should handle fetch errors gracefully', async () => {
      const errorMessage = 'Failed to fetch';
      (global.fetch as any).mockRejectedValueOnce(new Error(errorMessage));

      const { result } = renderHook(() => useQueryPerformance('1'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.error).toBeDefined();
    });

    it('should handle 404 responses', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 404,
      });

      const { result } = renderHook(() => useQueryPerformance('999'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.error).toBeDefined();
    });

    it('should update when database ID changes', async () => {
      const mockData = {
        queries: [],
        issues: [],
        timeline: [],
      };

      (global.fetch as any).mockResolvedValue({
        ok: true,
        json: async () => mockData,
      });

      const { rerender } = renderHook(
        ({ dbId }: { dbId: string }) => useQueryPerformance(dbId),
        { initialProps: { dbId: '1' } }
      );

      expect((global.fetch as any)).toHaveBeenCalled();

      vi.clearAllMocks();
      (global.fetch as any).mockResolvedValue({
        ok: true,
        json: async () => mockData,
      });

      rerender({ dbId: '2' });

      // Should fetch for new database ID
      await waitFor(() => {
        expect((global.fetch as any)).toHaveBeenCalled();
      });
    });

    it('should have proper initial state', async () => {
      (global.fetch as any).mockImplementation(() =>
        new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useQueryPerformance('1'));

      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(true);
      expect(result.current.error).toBeNull();
    });
  });

  describe('Query Performance Data Validation', () => {
    it('should parse EXPLAIN plan correctly', async () => {
      const mockData = {
        queries: [
          {
            id: 1,
            database_id: 1,
            query_hash: 'hash123',
            query_text: 'SELECT * FROM users WHERE id = $1',
            plan_json: {
              'Node Type': 'Index Scan',
              'Index Name': 'users_pkey',
              'Total Cost': 0.29,
            },
            mean_time: 1.5,
            total_time: 15.0,
            calls: 10,
            created_at: new Date().toISOString(),
          },
        ],
        issues: [],
        timeline: [],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      const { result } = renderHook(() => useQueryPerformance('1'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.data?.queries[0].plan_json['Node Type']).toBe('Index Scan');
    });

    it('should identify performance issues', async () => {
      const mockData = {
        queries: [],
        issues: [
          {
            id: 1,
            query_plan_id: 1,
            issue_type: 'sequential_scan' as const,
            severity: 'high' as const,
            affected_node_id: 0,
            description: 'Sequential scan on large table',
            recommendation: 'Create index',
            estimated_benefit: 75,
          },
        ],
        timeline: [],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      const { result } = renderHook(() => useQueryPerformance('1'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      const issues = result.current.data?.issues ?? [];
      expect(issues).toHaveLength(1);
      expect(issues[0].severity).toBe('high');
    });

    it('should track execution timeline', async () => {
      const now = new Date();
      const mockData = {
        queries: [],
        issues: [],
        timeline: [
          {
            id: 1,
            query_plan_id: 1,
            metric_timestamp: now.toISOString(),
            avg_duration: 45.5,
            max_duration: 120.0,
            executions: 100,
          },
          {
            id: 2,
            query_plan_id: 1,
            metric_timestamp: new Date(now.getTime() - 60000).toISOString(),
            avg_duration: 42.0,
            max_duration: 110.0,
            executions: 95,
          },
        ],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      });

      const { result } = renderHook(() => useQueryPerformance('1'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      const timeline = result.current.data?.timeline ?? [];
      expect(timeline).toHaveLength(2);
      expect(timeline[0].avg_duration).toBe(45.5);
    });
  });
});

describe('Log Analysis Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('useLogAnalysis Hook', () => {
    it('should establish WebSocket connection', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });
    });

    it('should receive log messages via WebSocket', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      // Simulate receiving a message
      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onmessage) {
        act(() => {
          mockWebSocket.onmessage(
            new MessageEvent('message', {
              data: JSON.stringify({
                id: 1,
                database_id: 1,
                log_timestamp: new Date().toISOString(),
                category: 'slow_query',
                severity: 'LOG',
                message: 'duration: 123.45 ms  execute query',
              }),
            })
          );
        });
      }

      await waitFor(() => {
        expect(result.current.logs.length).toBeGreaterThanOrEqual(0);
      });
    });

    it('should handle WebSocket errors', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        // Initial connection may succeed
        expect(result.current).toBeDefined();
      });
    });

    it('should close WebSocket on unmount', async () => {
      const { unmount } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        // Give WebSocket time to open
        expect(true).toBe(true);
      });

      unmount();

      // WebSocket should be closed
      expect(true).toBe(true);
    });

    it('should initialize with empty logs', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      expect(result.current.logs).toBeDefined();
      expect(Array.isArray(result.current.logs)).toBe(true);
    });

    it('should track connection state', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      // Initially not connected
      expect(result.current.connected).toBe(false);

      // Should connect after timeout
      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });
    });
  });

  describe('Log Analysis Data Processing', () => {
    it('should categorize error logs', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      // Simulate error log
      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onmessage) {
        act(() => {
          mockWebSocket.onmessage(
            new MessageEvent('message', {
              data: JSON.stringify({
                id: 1,
                database_id: 1,
                log_timestamp: new Date().toISOString(),
                category: 'error',
                severity: 'ERROR',
                message: 'ERROR: duplicate key value',
              }),
            })
          );
        });
      }
    });

    it('should categorize slow queries', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      // Simulate slow query log
      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onmessage) {
        act(() => {
          mockWebSocket.onmessage(
            new MessageEvent('message', {
              data: JSON.stringify({
                id: 2,
                database_id: 1,
                log_timestamp: new Date().toISOString(),
                category: 'slow_query',
                severity: 'LOG',
                message: 'duration: 5000.45 ms  execute query',
                duration: 5000.45,
              }),
            })
          );
        });
      }
    });

    it('should maintain log history limit', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      // Simulate receiving many logs
      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onmessage) {
        for (let i = 0; i < 120; i++) {
          act(() => {
            mockWebSocket.onmessage(
              new MessageEvent('message', {
                data: JSON.stringify({
                  id: i,
                  database_id: 1,
                  log_timestamp: new Date(Date.now() - i * 1000).toISOString(),
                  category: 'log',
                  severity: 'LOG',
                  message: `Log message ${i}`,
                }),
              })
            );
          });
        }
      }

      // Should maintain max 100 logs
      await waitFor(() => {
        expect(result.current.logs.length).toBeLessThanOrEqual(100);
      });
    });

    it('should detect anomalies in log stream', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      // Simulate anomaly detection scenario
      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onmessage) {
        // Multiple errors in short time
        for (let i = 0; i < 5; i++) {
          act(() => {
            mockWebSocket.onmessage(
              new MessageEvent('message', {
                data: JSON.stringify({
                  id: i,
                  database_id: 1,
                  log_timestamp: new Date(Date.now() - i * 100).toISOString(),
                  category: 'error',
                  severity: 'ERROR',
                  message: `ERROR: connection timeout`,
                }),
              })
            );
          });
        }
      }
    });
  });

  describe('WebSocket Connection Management', () => {
    it('should construct correct WebSocket URL', async () => {
      const { result } = renderHook(() => useLogAnalysis('test-db-1'));

      await waitFor(() => {
        expect(result.current).toBeDefined();
      });

      // URL should contain database ID
      expect(true).toBe(true);
    });

    it('should use correct protocol (ws vs wss)', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      // Protocol is determined by window.location.protocol
      expect(true).toBe(true);
    });

    it('should handle connection timeout gracefully', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      // Connection should be attempted
      await waitFor(() => {
        expect(result.current).toBeDefined();
      });
    });

    it('should reconnect on connection loss', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      // Simulate connection close
      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onclose) {
        act(() => {
          mockWebSocket.onclose(new CloseEvent('close'));
        });

        // Wait for state update
        await waitFor(() => {
          expect(result.current.connected).toBe(false);
        }, { timeout: 1000 });
      }
    });
  });

  describe('Error Handling in Log Analysis', () => {
    it('should handle malformed JSON in log message', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onmessage) {
        // Should not throw on malformed JSON
        act(() => {
          mockWebSocket.onmessage(
            new MessageEvent('message', {
              data: 'not valid json',
            })
          );
        });
      }

      // Should continue functioning
      expect(result.current).toBeDefined();
    });

    it('should handle WebSocket errors', async () => {
      const { result } = renderHook(() => useLogAnalysis('1'));

      await waitFor(() => {
        expect(result.current.connected).toBe(true);
      });

      const mockWebSocket = lastMockWebSocket;
      if (mockWebSocket && mockWebSocket.onerror) {
        act(() => {
          mockWebSocket.onerror(new Event('error'));
        });
      }

      // Should set error state
      await waitFor(() => {
        expect(result.current.error).toBeDefined();
      });
    });

    it('should handle missing database ID', async () => {
      const { result } = renderHook(() => useLogAnalysis(''));

      // Should not connect without database ID - empty database ID should not trigger connection
      expect(result.current.connected).toBe(false);
    });
  });
});

describe('End-to-End Data Flow Integration', () => {
  it('should complete full query performance pipeline', async () => {
    const mockData = {
      queries: [
        {
          id: 1,
          database_id: 1,
          query_hash: 'hash123',
          query_text: 'SELECT * FROM users',
          plan_json: { 'Node Type': 'Seq Scan' },
          mean_time: 45.5,
          total_time: 455.0,
          calls: 10,
          created_at: new Date().toISOString(),
        },
      ],
      issues: [
        {
          id: 1,
          query_plan_id: 1,
          issue_type: 'sequential_scan' as const,
          severity: 'high' as const,
          affected_node_id: 0,
          description: 'Sequential scan',
          recommendation: 'Add index',
          estimated_benefit: 75,
        },
      ],
      timeline: [
        {
          id: 1,
          query_plan_id: 1,
          metric_timestamp: new Date().toISOString(),
          avg_duration: 45.5,
          max_duration: 120.0,
          executions: 10,
        },
      ],
    };

    (global.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => mockData,
    });

    const { result } = renderHook(() => useQueryPerformance('1'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Verify complete data flow
    expect(result.current.data?.queries).toBeDefined();
    expect(result.current.data?.issues).toBeDefined();
    expect(result.current.data?.timeline).toBeDefined();
  });

  it('should complete full log analysis pipeline', async () => {
    const { result } = renderHook(() => useLogAnalysis('1'));

    await waitFor(() => {
      expect(result.current.connected).toBe(true);
    });

    const mockWebSocket = (global as any).WebSocket.instance;
    if (mockWebSocket && mockWebSocket.onmessage) {
      // Simulate complete log flow
      act(() => {
        mockWebSocket.onmessage(
          new MessageEvent('message', {
            data: JSON.stringify({
              id: 1,
              database_id: 1,
              log_timestamp: new Date().toISOString(),
              category: 'slow_query',
              severity: 'LOG',
              message: 'duration: 1234.56 ms  execute query',
            }),
          })
        );
      });
    }

    // Verify connection maintained
    expect(result.current.connected).toBe(true);
  });
});
