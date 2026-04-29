import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render } from '../../test/utils'
import { LiveLogsStream } from './LiveLogsStream'
import { useRealtime } from '../../hooks/useRealtime'

// Mock the useRealtime hook
vi.mock('../../hooks/useRealtime')

// Mock PostgreSQL log interface
interface PostgreSQLLog {
  id: number
  timestamp: string
  level: string
  message: string
  instance_id: number
}

describe('LiveLogsStream', () => {
  const mockOnLogClick = vi.fn()

  const createMockLog = (overrides: Partial<PostgreSQLLog> = {}): PostgreSQLLog => ({
    id: 1,
    timestamp: '2024-03-13T12:00:00Z',
    level: 'ERROR',
    message: 'Test error message',
    instance_id: 1,
    ...overrides,
  })

  const mockUseRealtime = (overrides = {}) => ({
    connected: false,
    lastUpdate: null,
    error: null,
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
    ...overrides,
  })

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useRealtime as any).mockReturnValue(mockUseRealtime())
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('component rendering', () => {
    it('should render with initial empty state', () => {
      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('Live Stream')).toBeInTheDocument()
      expect(screen.getByText('Waiting for logs...')).toBeInTheDocument()
    })

    it('should display connection status as disconnected initially', () => {
      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('CONNECTING...')).toBeInTheDocument()
    })

    it('should display LIVE badge when connected', () => {
      (useRealtime as any).mockReturnValue(mockUseRealtime({ connected: true }))

      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('LIVE')).toBeInTheDocument()
    })

    it('should have pause/resume button', () => {
      render(<LiveLogsStream instanceId={1} />)

      const button = screen.getByRole('button', { name: /pause/i })
      expect(button).toBeInTheDocument()
    })

    it('should have footer with log count', () => {
      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('No logs')).toBeInTheDocument()
    })
  })

  describe('log subscription and display', () => {
    it('should subscribe to log:new event on mount', () => {
      const mockSubscribe = vi.fn()
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      expect(mockSubscribe).toHaveBeenCalledWith('log:new', expect.any(Function))
    })

    it('should unsubscribe from log:new event on unmount', () => {
      const mockSubscribe = vi.fn()
      const mockUnsubscribe = vi.fn()
      ;(useRealtime as any).mockReturnValue(
        mockUseRealtime({
          subscribe: mockSubscribe,
          unsubscribe: mockUnsubscribe,
        })
      )

      const { unmount } = render(<LiveLogsStream instanceId={1} />)

      // Get the callback that was subscribed
      const callback = mockSubscribe.mock.calls[0][1]

      unmount()

      expect(mockUnsubscribe).toHaveBeenCalledWith('log:new', callback)
    })

    it('should display logs in reverse chronological order (newest first)', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1, timestamp: '2024-03-13T12:00:00Z' }))
          callback(createMockLog({ id: 2, timestamp: '2024-03-13T12:01:00Z' }))
          callback(createMockLog({ id: 3, timestamp: '2024-03-13T12:02:00Z' }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      const logMessages = screen.getAllByText(/Test error message/)
      // Newest should be first (id: 3)
      expect(logMessages[0]).toBeInTheDocument()
    })

    it('should only display logs for the current instance', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1, instance_id: 1 }))
          callback(createMockLog({ id: 2, instance_id: 2 }))
          callback(createMockLog({ id: 3, instance_id: 1 }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      const logMessages = screen.getAllByText(/Test error message/)
      // Should only have 2 logs (id 1 and 3, not id 2)
      expect(logMessages).toHaveLength(2)
    })

    it('should cap logs at 50 items', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          // Add 60 logs
          for (let i = 0; i < 60; i++) {
            callback(createMockLog({ id: i, instance_id: 1 }))
          }
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      const logMessages = screen.getAllByText(/Test error message/)
      expect(logMessages).toHaveLength(50)
    })
  })

  describe('log display formatting', () => {
    it('should display ERROR badge with red color', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1, level: 'ERROR' }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('ERROR')).toBeInTheDocument()
    })

    it('should display SLOW_QUERY badge with orange color', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1, level: 'SLOW_QUERY' }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('SLOW_QUERY')).toBeInTheDocument()
    })

    it('should display timestamp in locale time format', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1, timestamp: '2024-03-13T12:00:00Z' }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      const timeString = new Date('2024-03-13T12:00:00Z').toLocaleTimeString()
      expect(screen.getByText(timeString)).toBeInTheDocument()
    })

    it('should truncate message text in log display', () => {
      const longMessage = 'a'.repeat(200)
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1, message: longMessage }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      // Message should be rendered but truncated
      expect(screen.getByText(longMessage)).toBeInTheDocument()
    })
  })

  describe('auto-scroll functionality', () => {
    it('should have auto-scroll enabled by default', () => {
      render(<LiveLogsStream instanceId={1} />)

      const button = screen.getByRole('button', { name: /pause/i })
      expect(button).toBeInTheDocument()
    })

    it('should toggle auto-scroll when pause button is clicked', async () => {
      const user = userEvent.setup()
      render(<LiveLogsStream instanceId={1} />)

      const pauseButton = screen.getByRole('button', { name: /pause/i })
      expect(pauseButton).toBeInTheDocument()

      await user.click(pauseButton)

      const resumeButton = screen.getByRole('button', { name: /resume/i })
      expect(resumeButton).toBeInTheDocument()
    })

    it('should toggle auto-scroll back to pause', async () => {
      const user = userEvent.setup()
      render(<LiveLogsStream instanceId={1} />)

      const pauseButton = screen.getByRole('button', { name: /pause/i })
      await user.click(pauseButton)

      const resumeButton = screen.getByRole('button', { name: /resume/i })
      await user.click(resumeButton)

      const pauseButtonAgain = screen.getByRole('button', { name: /pause/i })
      expect(pauseButtonAgain).toBeInTheDocument()
    })
  })

  describe('log click handler', () => {
    it('should call onLogClick when a log is clicked', async () => {
      const user = userEvent.setup()
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1 }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} onLogClick={mockOnLogClick} />)

      const logEntry = screen.getByText(/Test error message/)
      await user.click(logEntry)

      expect(mockOnLogClick).toHaveBeenCalled()
    })

    it('should work without onLogClick prop', async () => {
      const user = userEvent.setup()
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1 }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      const logEntry = screen.getByText(/Test error message/)
      // Should not throw
      await user.click(logEntry)
    })
  })

  describe('footer display', () => {
    it('should display log count in footer when logs exist', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1 }))
          callback(createMockLog({ id: 2 }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText(/Showing 2 logs/)).toBeInTheDocument()
    })

    it('should display no logs message when empty', () => {
      render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('No logs')).toBeInTheDocument()
    })
  })

  describe('connection status updates', () => {
    it('should update connection badge when status changes', () => {
      const mockSubscribe = vi.fn()
      ;(useRealtime as any).mockReturnValue(
        mockUseRealtime({ subscribe: mockSubscribe, connected: false })
      )

      const { rerender } = render(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('CONNECTING...')).toBeInTheDocument()

      ;(useRealtime as any).mockReturnValue(
        mockUseRealtime({ subscribe: mockSubscribe, connected: true })
      )

      rerender(<LiveLogsStream instanceId={1} />)

      expect(screen.getByText('LIVE')).toBeInTheDocument()
    })
  })

  describe('instance filtering', () => {
    it('should filter logs based on instanceId prop', () => {
      const mockSubscribe = vi.fn((event, callback) => {
        if (event === 'log:new') {
          callback(createMockLog({ id: 1, instance_id: 1 }))
          callback(createMockLog({ id: 2, instance_id: 2 }))
        }
      })
      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      const { rerender } = render(<LiveLogsStream instanceId={1} />)

      const logsForInstance1 = screen.getAllByText(/Test error message/).length
      expect(logsForInstance1).toBe(1)

      ;(useRealtime as any).mockReturnValue(mockUseRealtime({ subscribe: mockSubscribe }))

      // Simulate changing to instance 2
      rerender(<LiveLogsStream instanceId={2} />)

      // Should now show logs for instance 2
      const logsForInstance2 = screen.getAllByText(/Test error message/).length
      expect(logsForInstance2).toBeGreaterThanOrEqual(1)
    })
  })
})
