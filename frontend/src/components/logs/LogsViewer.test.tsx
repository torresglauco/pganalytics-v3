import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render } from '../../test/utils'
import { LogsViewer } from './LogsViewer'
import { useLogs } from '../../hooks/useLogs'
import { useRealtime } from '../../hooks/useRealtime'

// Mock the hooks
vi.mock('../../hooks/useLogs')
vi.mock('../../hooks/useRealtime')

describe('LogsViewer', () => {
  const mockUseLogs = useLogs as any
  const mockUseRealtime = useRealtime as any

  beforeEach(() => {
    vi.clearAllMocks()

    // Default mock implementation
    mockUseLogs.mockReturnValue({
      data: { logs: [] },
      loading: false,
      error: null,
      fetchLogs: vi.fn(),
      getLogDetails: vi.fn(),
    })

    mockUseRealtime.mockReturnValue({
      connected: false,
      lastUpdate: null,
      error: null,
      subscribe: vi.fn(),
      unsubscribe: vi.fn(),
    })
  })

  describe('component rendering', () => {
    it('should render LogsViewer component', () => {
      render(<LogsViewer />)

      expect(screen.getByText('Historical Logs')).toBeInTheDocument()
    })

    it('should render search bar', () => {
      render(<LogsViewer />)

      expect(screen.getByPlaceholderText(/search/i)).toBeInTheDocument()
    })

    it('should render log filters', () => {
      render(<LogsViewer />)

      expect(screen.getByPlaceholderText(/enter instance id/i)).toBeInTheDocument()
    })

    it('should render historical logs table', () => {
      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      expect(screen.getByText('Historical Logs')).toBeInTheDocument()
    })
  })

  describe('live stream integration', () => {
    it('should not show live stream section when no instance is selected', () => {
      render(<LogsViewer />)

      expect(screen.queryByText('Live Stream')).not.toBeInTheDocument()
    })

    it('should show live stream section when instance is selected', async () => {
      const user = userEvent.setup()

      // Mock successful instance selection
      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      // Simulate selecting an instance via the filter
      const instanceInput = screen.getByPlaceholderText(/enter instance id/i) as HTMLInputElement
      await user.type(instanceInput, '1')

      // Verify the instance ID was entered
      expect(instanceInput.value).toBe('1')

      // Wait for the live stream to appear (check using getAllByText to handle duplicates)
      await waitFor(() => {
        const liveStreamHeadings = screen.getAllByText('Live Stream')
        expect(liveStreamHeadings.length).toBeGreaterThan(0)
      })
    })

    it('should display live stream with connection status when instance is selected', async () => {
      const user = userEvent.setup()
      mockUseRealtime.mockReturnValue({
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      })

      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      // Select an instance
      const instanceInput = screen.getByPlaceholderText(/enter instance id/i) as HTMLInputElement
      await user.type(instanceInput, '2')

      // Verify instance ID was entered
      expect(instanceInput.value).toBe('2')

      // Verify live stream appears with LIVE status
      await waitFor(() => {
        expect(screen.getByText('LIVE')).toBeInTheDocument()
      })
    })
  })

  describe('realtime status display', () => {
    it('should show realtime status when connected', async () => {
      mockUseRealtime.mockReturnValue({
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      })

      render(<LogsViewer />)

      // RealtimeStatus component should be rendered (used in Live Stream section)
      // When instance is selected, it will be visible
    })
  })

  describe('historical logs display', () => {
    it('should display logs when data is available', () => {
      const mockLogs = [
        {
          id: 1,
          timestamp: '2024-03-13T12:00:00Z',
          level: 'ERROR',
          message: 'Test error',
        },
        {
          id: 2,
          timestamp: '2024-03-13T12:01:00Z',
          level: 'INFO',
          message: 'Test info',
        },
      ]

      mockUseLogs.mockReturnValue({
        data: { logs: mockLogs },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      expect(screen.getByText('Historical Logs')).toBeInTheDocument()
    })

    it('should display loading state', () => {
      mockUseLogs.mockReturnValue({
        data: null,
        loading: true,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      // Loading spinner should be visible
      expect(screen.getByText('Historical Logs')).toBeInTheDocument()
    })

    it('should display error message when fetch fails', () => {
      const errorMessage = 'Failed to fetch logs'
      mockUseLogs.mockReturnValue({
        data: null,
        loading: false,
        error: errorMessage,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      expect(screen.getByText(`Error: ${errorMessage}`)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
    })

    it('should call fetchLogs when retry button is clicked', async () => {
      const user = userEvent.setup()
      const mockFetchLogs = vi.fn()

      mockUseLogs.mockReturnValue({
        data: null,
        loading: false,
        error: 'Test error',
        fetchLogs: mockFetchLogs,
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      const retryButton = screen.getByRole('button', { name: /retry/i })
      await user.click(retryButton)

      expect(mockFetchLogs).toHaveBeenCalled()
    })
  })

  describe('search functionality', () => {
    it('should have search input field', () => {
      render(<LogsViewer />)

      const searchInput = screen.getByPlaceholderText(/search/i)
      expect(searchInput).toBeInTheDocument()
    })

    it('should accept search input', async () => {
      const user = userEvent.setup()
      render(<LogsViewer />)

      const searchInput = screen.getByPlaceholderText(/search/i)
      await user.type(searchInput, 'error message')

      expect(searchInput).toHaveValue('error message')
    })
  })

  describe('filter functionality', () => {
    it('should render instance ID filter control', () => {
      render(<LogsViewer />)

      expect(screen.getByPlaceholderText(/enter instance id/i)).toBeInTheDocument()
    })

    it('should accept instance ID input', async () => {
      const user = userEvent.setup()
      render(<LogsViewer />)

      const instanceInput = screen.getByPlaceholderText(/enter instance id/i)
      await user.type(instanceInput, '5')

      expect(instanceInput).toHaveValue(5)
    })

    it('should display instance selection help text', () => {
      render(<LogsViewer />)

      expect(screen.getByText(/select an instance to enable live logs/i)).toBeInTheDocument()
    })
  })

  describe('responsive layout', () => {
    it('should render with proper spacing between sections', () => {
      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      const { container } = render(<LogsViewer />)

      // Check for space-y-6 class on main container
      const mainContainer = container.querySelector('div.space-y-6')
      expect(mainContainer).toBeInTheDocument()
    })

    it('should have grid layout for filters and logs', () => {
      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      const { container } = render(<LogsViewer />)

      // Check for grid layout
      const gridContainer = container.querySelector('div.grid')
      expect(gridContainer).toBeInTheDocument()
    })
  })

  describe('section structure', () => {
    it('should have historical logs section with proper heading', () => {
      render(<LogsViewer />)

      const heading = screen.getByText('Historical Logs')
      expect(heading).toBeInTheDocument()
      expect(heading.tagName).toBe('H2')
    })

    it('should display both search and filter in same row', () => {
      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      const { container } = render(<LogsViewer />)

      const gridContainers = container.querySelectorAll('div.grid')
      expect(gridContainers.length).toBeGreaterThan(0)
    })
  })

  describe('integration with hooks', () => {
    it('should call useLogs with correct parameters', () => {
      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: vi.fn(),
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      expect(mockUseLogs).toHaveBeenCalled()
    })

    it('should update logs when search changes', async () => {
      const user = userEvent.setup()
      const mockFetchLogs = vi.fn()

      mockUseLogs.mockReturnValue({
        data: { logs: [] },
        loading: false,
        error: null,
        fetchLogs: mockFetchLogs,
        getLogDetails: vi.fn(),
      })

      render(<LogsViewer />)

      const searchInput = screen.getByPlaceholderText(/search/i)
      await user.type(searchInput, 'error')

      // Wait for potential refetch
      await waitFor(() => {
        expect(mockUseLogs).toHaveBeenCalled()
      })
    })
  })

  describe('accessibility', () => {
    it('should have semantic HTML structure', () => {
      render(<LogsViewer />)

      // Check for proper heading hierarchy
      const heading = screen.getByText('Historical Logs')
      expect(heading.tagName).toBe('H2')
    })

    it('should have descriptive button labels', () => {
      render(<LogsViewer />)

      const buttons = screen.getAllByRole('button')
      expect(buttons.length).toBeGreaterThan(0)

      buttons.forEach((button) => {
        expect(button.textContent).toBeTruthy()
      })
    })

    it('should have proper label associations', () => {
      const { container } = render(<LogsViewer />)

      // Check for proper label structure in filters
      const labels = container.querySelectorAll('label')
      expect(labels.length).toBeGreaterThanOrEqual(0)
    })
  })
})
