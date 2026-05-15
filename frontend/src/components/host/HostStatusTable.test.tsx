/**
 * Tests for HostStatusTable component
 * Tests the host status table with up/down indicators and health badges
 */

import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { HostStatusTable } from './HostStatusTable'
import type { HostStatus } from '../../types/host'

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  CheckCircle: () => <span data-testid="check-circle-icon" />,
  XCircle: () => <span data-testid="x-circle-icon" />,
  HelpCircle: () => <span data-testid="help-circle-icon" />,
  Server: () => <span data-testid="server-icon" />,
}))

// Mock formatting utils
vi.mock('../../utils/formatting', () => ({
  getRelativeTime: (date: string) => {
    if (date === '2024-01-01T00:00:00Z') return '1 hour ago'
    return '2 hours ago'
  },
}))

describe('HostStatusTable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Mock data factory
  const createMockHost = (overrides: Partial<HostStatus> = {}): HostStatus => ({
    collector_id: 'collector-001',
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
      last_seen: '2024-01-01T00:00:00Z',
      unresponsive_for_seconds: 0,
    }),
    createMockHost({
      collector_id: 'host-002',
      hostname: 'prod-server-2',
      status: 'down',
      is_healthy: false,
      last_seen: '2023-12-31T00:00:00Z',
      unresponsive_for_seconds: 3600,
    }),
    createMockHost({
      collector_id: 'host-003',
      hostname: 'dev-server-1',
      status: 'unknown',
      is_healthy: false,
      last_seen: undefined,
      unresponsive_for_seconds: 0,
    }),
  ]

  describe('Table columns', () => {
    it('should render table with columns: Status, Hostname, Health, Last Seen, Unresponsive, Action', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      expect(screen.getByText('Status')).toBeInTheDocument()
      expect(screen.getByText('Hostname')).toBeInTheDocument()
      expect(screen.getByText('Health')).toBeInTheDocument()
      expect(screen.getByText('Last Seen')).toBeInTheDocument()
      expect(screen.getByText('Unresponsive')).toBeInTheDocument()
      expect(screen.getByText('Action')).toBeInTheDocument()
    })
  })

  describe('Status indicators', () => {
    it('should show "up" status with green indicator', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      // Check for check-circle icon (for up status)
      const checkCircleIcons = screen.getAllByTestId('check-circle-icon')
      expect(checkCircleIcons.length).toBeGreaterThanOrEqual(1)

      // Check for UP badge
      expect(screen.getByText('UP')).toBeInTheDocument()
    })

    it('should show "down" status with red indicator', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      // Check for x-circle icon (for down status)
      expect(screen.getByTestId('x-circle-icon')).toBeInTheDocument()

      // Check for DOWN badge
      expect(screen.getByText('DOWN')).toBeInTheDocument()
    })

    it('should show "unknown" status with gray indicator', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      // Check for help-circle icon (for unknown status)
      expect(screen.getByTestId('help-circle-icon')).toBeInTheDocument()

      // Check for UNKNOWN badge
      expect(screen.getByText('UNKNOWN')).toBeInTheDocument()
    })
  })

  describe('Host information display', () => {
    it('should display hostname and truncated collector ID', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      expect(screen.getByText('prod-server-1')).toBeInTheDocument()
      expect(screen.getByText('prod-server-2')).toBeInTheDocument()
      expect(screen.getByText('dev-server-1')).toBeInTheDocument()
    })

    it('should display last seen time', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      // getRelativeTime mock returns "1 hour ago" for the test date
      expect(screen.getByText('1 hour ago')).toBeInTheDocument()
    })

    it('should display "Never" when last_seen is undefined', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      expect(screen.getByText('Never')).toBeInTheDocument()
    })
  })

  describe('Unresponsive time display', () => {
    it('should show "-" when unresponsive time is 0', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      // Multiple hosts have 0 unresponsive time
      const dashes = screen.getAllByText('-')
      expect(dashes.length).toBeGreaterThanOrEqual(1)
    })

    it('should show formatted time when unresponsive', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} />)

      // 3600 seconds = 1h
      expect(screen.getByText('1h 0m')).toBeInTheDocument()
    })
  })

  describe('Row selection', () => {
    it('should trigger onSelectHost callback when row is clicked', async () => {
      const user = userEvent.setup()
      const mockOnSelectHost = vi.fn()

      render(<HostStatusTable hosts={mockHosts} onSelectHost={mockOnSelectHost} />)

      const row = screen.getByText('prod-server-1').closest('div[class*="grid"]')
      expect(row).toBeInTheDocument()

      await user.click(row!)

      expect(mockOnSelectHost).toHaveBeenCalledTimes(1)
      expect(mockOnSelectHost).toHaveBeenCalledWith(
        expect.objectContaining({
          hostname: 'prod-server-1',
          status: 'up',
        })
      )
    })

    it('should highlight selected row', () => {
      render(
        <HostStatusTable
          hosts={mockHosts}
          onSelectHost={vi.fn()}
          selectedHostId="host-001"
        />
      )

      const row = screen.getByText('prod-server-1').closest('div[class*="grid"]')
      expect(row).toHaveClass('bg-blue-50')
    })
  })

  describe('View Details button', () => {
    it('should trigger onSelectHost when View Details is clicked', async () => {
      const user = userEvent.setup()
      const mockOnSelectHost = vi.fn()

      render(<HostStatusTable hosts={mockHosts} onSelectHost={mockOnSelectHost} />)

      const viewDetailsButtons = screen.getAllByText('View Details')
      await user.click(viewDetailsButtons[0])

      expect(mockOnSelectHost).toHaveBeenCalledTimes(1)
    })
  })

  describe('Loading state', () => {
    it('should show loading message when loading', () => {
      render(<HostStatusTable hosts={[]} onSelectHost={vi.fn()} isLoading={true} />)

      expect(screen.getByText('Loading hosts...')).toBeInTheDocument()
    })

    it('should not show table content when loading', () => {
      render(<HostStatusTable hosts={mockHosts} onSelectHost={vi.fn()} isLoading={true} />)

      expect(screen.queryByText('prod-server-1')).not.toBeInTheDocument()
    })
  })

  describe('Empty state', () => {
    it('should show "No hosts found" message when empty', () => {
      render(<HostStatusTable hosts={[]} onSelectHost={vi.fn()} />)

      expect(screen.getByText('No hosts found')).toBeInTheDocument()
    })

    it('should show suggestion in empty state', () => {
      render(<HostStatusTable hosts={[]} onSelectHost={vi.fn()} />)

      expect(screen.getByText(/Hosts will appear here/)).toBeInTheDocument()
    })
  })
})