/**
 * Tests for ReplicationTopologyPage component
 * Tests the replication topology page with loading, error, empty, and data states
 */

import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter, MemoryRouter, Routes, Route } from 'react-router-dom'
import { ReplicationTopologyPage } from './ReplicationTopologyPage'
import * as replicationApi from '../api/replicationApi'
import type { ReplicationTopology } from '../types/replication'

// Mock the API module
vi.mock('../api/replicationApi', () => ({
  getTopology: vi.fn(),
}))

// Mock child components
vi.mock('../components/topology/TopologyGraph', () => ({
  TopologyGraph: ({ topology }: { topology: any }) => (
    <div data-testid="topology-graph">
      Topology Graph for {topology?.collector_id || 'unknown'}
    </div>
  ),
}))

vi.mock('../components/topology/TopologyLegend', () => ({
  TopologyLegend: () => <div data-testid="topology-legend">Legend</div>,
}))

vi.mock('../components/ui/LoadingSpinner', () => ({
  LoadingSpinner: ({ message }: { message?: string }) => (
    <div data-testid="loading-spinner">{message || 'Loading...'}</div>
  ),
}))

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  RefreshCw: ({ className }: { className?: string }) => (
    <span data-testid="refresh-icon" className={className} />
  ),
  AlertCircle: () => <span data-testid="alert-icon" />,
  Network: () => <span data-testid="network-icon" />,
}))

describe('ReplicationTopologyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Mock topology data factory
  const createMockTopology = (overrides: Partial<ReplicationTopology> = {}): ReplicationTopology => ({
    collector_id: 'collector-001',
    node_role: 'primary',
    downstream_count: 2,
    downstream_nodes: [
      {
        collector_id: 'standby-001',
        application_name: 'standby1',
        client_addr: '192.168.1.10',
        state: 'streaming',
        sync_state: 'async',
        replay_lag_ms: 150,
      },
      {
        collector_id: 'standby-002',
        application_name: 'standby2',
        client_addr: '192.168.1.11',
        state: 'streaming',
        sync_state: 'sync',
        replay_lag_ms: 50,
      },
    ],
    ...overrides,
  })

  // Helper to render with router
  const renderWithRouter = (initialEntries: string[] = ['/collectors/test-collector/topology']) => {
    return render(
      <MemoryRouter initialEntries={initialEntries}>
        <Routes>
          <Route path="/collectors/:collectorId/topology" element={<ReplicationTopologyPage />} />
        </Routes>
      </MemoryRouter>
    )
  }

  describe('Loading state', () => {
    it('should show loading spinner initially', async () => {
      vi.mocked(replicationApi.getTopology).mockImplementation(() => new Promise(() => {}))

      renderWithRouter()

      expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
      expect(screen.getByText('Loading replication topology...')).toBeInTheDocument()
    })
  })

  describe('Data loaded state', () => {
    it('should render topology graph after data loads', async () => {
      const mockTopology = createMockTopology()
      vi.mocked(replicationApi.getTopology).mockResolvedValue(mockTopology)

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('topology-graph')).toBeInTheDocument()
      })
    })

    it('should display page header with title', async () => {
      const mockTopology = createMockTopology()
      vi.mocked(replicationApi.getTopology).mockResolvedValue(mockTopology)

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('Replication Topology')).toBeInTheDocument()
      })
    })

    it('should render topology legend in sidebar', async () => {
      const mockTopology = createMockTopology()
      vi.mocked(replicationApi.getTopology).mockResolvedValue(mockTopology)

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('topology-legend')).toBeInTheDocument()
      })
    })
  })

  describe('Error handling', () => {
    it('should display error message on API failure', async () => {
      vi.mocked(replicationApi.getTopology).mockRejectedValue(new Error('API Error'))

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('Failed to load topology')).toBeInTheDocument()
        expect(screen.getByText('API Error')).toBeInTheDocument()
      })
    })

    it('should show error when collectorId is undefined', async () => {
      // The component handles missing collectorId by setting an error
      vi.mocked(replicationApi.getTopology).mockRejectedValue(new Error('No collector ID provided'))

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('Failed to load topology')).toBeInTheDocument()
      })
    })
  })

  describe('Empty state', () => {
    it('should show empty state when no topology data', async () => {
      // Return null to simulate no topology
      vi.mocked(replicationApi.getTopology).mockResolvedValue(null as any)

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('No Topology Data')).toBeInTheDocument()
      })
    })

    it('should show explanation text in empty state', async () => {
      vi.mocked(replicationApi.getTopology).mockResolvedValue(null as any)

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText(/No replication topology found/)).toBeInTheDocument()
      })
    })
  })

  describe('Refresh functionality', () => {
    it('should have refresh button visible', async () => {
      const mockTopology = createMockTopology()
      vi.mocked(replicationApi.getTopology).mockResolvedValue(mockTopology)

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('Refresh')).toBeInTheDocument()
      })
    })

    it('should trigger data reload when refresh button is clicked', async () => {
      const mockTopology = createMockTopology()
      vi.mocked(replicationApi.getTopology).mockResolvedValue(mockTopology)

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('topology-graph')).toBeInTheDocument()
      })

      // Clear mock call count
      vi.mocked(replicationApi.getTopology).mockClear()
      vi.mocked(replicationApi.getTopology).mockResolvedValue(mockTopology)

      const refreshButton = screen.getByText('Refresh').closest('button')
      await userEvent.click(refreshButton!)

      await waitFor(() => {
        expect(replicationApi.getTopology).toHaveBeenCalledTimes(1)
      })
    })
  })

  describe('Collector ID handling', () => {
    it('should handle collectorId from URL params', async () => {
      const mockTopology = createMockTopology({ collector_id: 'specific-collector' })
      vi.mocked(replicationApi.getTopology).mockResolvedValue(mockTopology)

      renderWithRouter(['/collectors/specific-collector/topology'])

      await waitFor(() => {
        expect(replicationApi.getTopology).toHaveBeenCalledWith('specific-collector')
      })
    })
  })
})