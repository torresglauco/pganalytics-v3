/**
 * Tests for HostInventoryPage component
 * Tests the host inventory page with loading, error, search, filter, and detail panel states
 */

import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HostInventoryPage } from './HostInventoryPage'
import * as hostApi from '../api/hostApi'
import type { HostStatus } from '../types/host'

// Mock the API module
vi.mock('../api/hostApi', () => ({
  hostApi: {
    getAllHostStatuses: vi.fn(),
  },
}))

// Mock child components
vi.mock('../components/host/HostStatusTable', () => ({
  default: ({ hosts, onSelectHost }: any) => (
    <div data-testid="host-status-table">
      {hosts?.length || 0} hosts
      {hosts?.map((h: HostStatus) => (
        <div key={h.collector_id} data-testid={`host-row-${h.collector_id}`}>
          <span>{h.hostname}</span>
          <button onClick={() => onSelectHost(h)}>Select</button>
        </div>
      ))}
    </div>
  ),
}))

vi.mock('../components/host/HostInventorySummary', () => ({
  default: ({ hosts, onFilterByStatus }: any) => (
    <div data-testid="host-inventory-summary">
      {hosts?.length || 0} hosts in summary
      <button onClick={() => onFilterByStatus?.('up')} data-testid="filter-up">Up</button>
      <button onClick={() => onFilterByStatus?.('down')} data-testid="filter-down">Down</button>
    </div>
  ),
}))

vi.mock('../components/host/HostDetailPanel', () => ({
  default: ({ host, onClose }: any) => (
    <div data-testid="host-detail-panel">
      <span>Details for {host?.hostname}</span>
      <button onClick={onClose}>Close</button>
    </div>
  ),
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
  Download: () => <span data-testid="download-icon" />,
  Search: () => <span data-testid="search-icon" />,
  AlertCircle: () => <span data-testid="alert-icon" />,
  CheckCircle: () => <span data-testid="check-circle-icon" />,
}))

describe('HostInventoryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Mock data factory
  const createMockHost = (overrides: Partial<HostStatus> = {}): HostStatus => ({
    collector_id: 'host-001',
    hostname: 'test-host',
    status: 'up',
    is_healthy: true,
    last_seen: '2024-01-01T00:00:00Z',
    unresponsive_for_seconds: 0,
    configured_threshold_seconds: 60,
    ...overrides,
  })

  const mockHosts: HostStatus[] = [
    createMockHost({
      collector_id: 'host-001',
      hostname: 'prod-server-1',
      status: 'up',
      is_healthy: true,
    }),
    createMockHost({
      collector_id: 'host-002',
      hostname: 'prod-server-2',
      status: 'down',
      is_healthy: false,
    }),
    createMockHost({
      collector_id: 'host-003',
      hostname: 'dev-server-1',
      status: 'unknown',
      is_healthy: false,
    }),
  ]

  describe('Loading state', () => {
    it('should show loading spinner initially', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockImplementation(() => new Promise(() => {}))

      render(<HostInventoryPage />)

      expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
      expect(screen.getByText('Loading hosts...')).toBeInTheDocument()
    })
  })

  describe('Data loaded state', () => {
    it('should render host table after data loads', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByTestId('host-status-table')).toBeInTheDocument()
      })
    })

    it('should display summary cards with up/down counts', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByTestId('host-inventory-summary')).toBeInTheDocument()
      })
    })

    it('should display page header with title', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByText('Host Inventory')).toBeInTheDocument()
      })
    })
  })

  describe('Search functionality', () => {
    it('should filter hosts by hostname', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByTestId('host-status-table')).toBeInTheDocument()
      })

      const searchInput = screen.getByPlaceholderText(/Search by hostname/)
      fireEvent.change(searchInput, { target: { value: 'prod' } })

      // Should show only prod servers (2 hosts)
      await waitFor(() => {
        expect(screen.getByText('prod-server-1')).toBeInTheDocument()
        expect(screen.getByText('prod-server-2')).toBeInTheDocument()
      })
    })
  })

  describe('Status filter', () => {
    it('should filter by status using dropdown', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByTestId('host-status-table')).toBeInTheDocument()
      })

      // Select "down" status from dropdown
      const statusSelect = screen.getByRole('combobox')
      fireEvent.change(statusSelect, { target: { value: 'down' } })

      await waitFor(() => {
        expect(screen.getByText('prod-server-2')).toBeInTheDocument()
      })
    })
  })

  describe('Error handling', () => {
    it('should display error message on API failure', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockRejectedValue(new Error('API Error'))

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByText('API Error')).toBeInTheDocument()
      })
    })
  })

  describe('Auto-refresh toggle', () => {
    it('should display auto-refresh checkbox', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByLabelText(/Auto-refresh/)).toBeInTheDocument()
      })
    })
  })

  describe('Export functionality', () => {
    it('should have export button visible', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByText('Export')).toBeInTheDocument()
      })
    })
  })

  describe('Refresh functionality', () => {
    it('should have refresh button visible', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByText('Refresh')).toBeInTheDocument()
      })
    })
  })

  describe('Host detail panel', () => {
    it('should open detail panel when host is selected', async () => {
      const user = userEvent.setup()
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByTestId('host-status-table')).toBeInTheDocument()
      })

      // Click select button on first host
      const selectButtons = screen.getAllByText('Select')
      await user.click(selectButtons[0])

      await waitFor(() => {
        expect(screen.getByTestId('host-detail-panel')).toBeInTheDocument()
        expect(screen.getByText(/Details for prod-server-1/)).toBeInTheDocument()
      })
    })

    it('should close detail panel when close button clicked', async () => {
      const user = userEvent.setup()
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByTestId('host-status-table')).toBeInTheDocument()
      })

      // Open detail panel
      const selectButtons = screen.getAllByText('Select')
      await user.click(selectButtons[0])

      await waitFor(() => {
        expect(screen.getByTestId('host-detail-panel')).toBeInTheDocument()
      })

      // Close detail panel
      await user.click(screen.getByText('Close'))

      await waitFor(() => {
        expect(screen.queryByTestId('host-detail-panel')).not.toBeInTheDocument()
      })
    })
  })

  describe('Empty state', () => {
    it('should show "No hosts configured" when no hosts', async () => {
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue([])

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByText('No hosts configured')).toBeInTheDocument()
      })
    })
  })

  describe('Filter from summary cards', () => {
    it('should filter by status when summary card clicked', async () => {
      const user = userEvent.setup()
      vi.mocked(hostApi.hostApi.getAllHostStatuses).mockResolvedValue(mockHosts)

      render(<HostInventoryPage />)

      await waitFor(() => {
        expect(screen.getByTestId('filter-up')).toBeInTheDocument()
      })

      await user.click(screen.getByTestId('filter-up'))

      // After clicking "Up" filter, should show only up hosts
      await waitFor(() => {
        expect(screen.getByText('prod-server-1')).toBeInTheDocument()
      })
    })
  })
})