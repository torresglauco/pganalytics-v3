import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { render } from '../../test/utils'
import { RealtimeStatus } from './RealtimeStatus'
import { useRealtime } from '../../hooks/useRealtime'

// Mock the useRealtime hook
vi.mock('../../hooks/useRealtime')

describe('RealtimeStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('connection status display', () => {
    it('should render Live status when connected', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      render(<RealtimeStatus />)

      expect(screen.getByText('Live')).toBeInTheDocument()
    })

    it('should render Polling status when disconnected', () => {
      (useRealtime as any).mockReturnValue({ connected: false })

      render(<RealtimeStatus />)

      expect(screen.getByText('Polling')).toBeInTheDocument()
    })
  })

  describe('visual indicators', () => {
    it('should display green dot when connected', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      const { container } = render(<RealtimeStatus />)

      // Check for the green dot with animate-pulse class
      const greenDot = container.querySelector(
        'div.bg-green-500.animate-pulse'
      )
      expect(greenDot).toBeInTheDocument()
    })

    it('should display yellow dot when disconnected', () => {
      (useRealtime as any).mockReturnValue({ connected: false })

      const { container } = render(<RealtimeStatus />)

      // Check for the yellow dot
      const yellowDot = container.querySelector('div.bg-yellow-500')
      expect(yellowDot).toBeInTheDocument()
    })
  })

  describe('styling', () => {
    it('should have correct text color when connected', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      render(<RealtimeStatus />)

      const statusText = screen.getByText('Live')
      // Check for green text styling
      expect(statusText.className).toContain('text-green')
    })

    it('should have correct text color when disconnected', () => {
      (useRealtime as any).mockReturnValue({ connected: false })

      render(<RealtimeStatus />)

      const statusText = screen.getByText('Polling')
      // Check for yellow text styling
      expect(statusText.className).toContain('text-yellow')
    })

    it('should be small and compact in size', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      render(<RealtimeStatus />)

      // Check for small text size class
      const statusText = screen.getByText('Live')
      expect(statusText.className).toContain('text-xs')
    })

    it('should have small dot indicator', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      const { container } = render(<RealtimeStatus />)

      // Check for w-2 h-2 classes on the dot
      const dot = container.querySelector('div.w-2.h-2')
      expect(dot).toBeInTheDocument()
    })
  })

  describe('animation', () => {
    it('should animate pulse when connected', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      const { container } = render(<RealtimeStatus />)

      // Check for animate-pulse class when connected
      const pulseDot = container.querySelector('div.animate-pulse')
      expect(pulseDot).toBeInTheDocument()
    })

    it('should not animate when disconnected', () => {
      (useRealtime as any).mockReturnValue({ connected: false })

      const { container } = render(<RealtimeStatus />)

      // Check that there is no animate-pulse class when disconnected
      const pulseDot = container.querySelector('div.animate-pulse')
      expect(pulseDot).not.toBeInTheDocument()
    })
  })

  describe('dark mode support', () => {
    it('should have dark mode text color for connected state', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      render(<RealtimeStatus />)

      const statusText = screen.getByText('Live')
      expect(statusText.className).toContain('dark:text-green')
    })

    it('should have dark mode text color for disconnected state', () => {
      (useRealtime as any).mockReturnValue({ connected: false })

      render(<RealtimeStatus />)

      const statusText = screen.getByText('Polling')
      expect(statusText.className).toContain('dark:text-yellow')
    })
  })

  describe('component structure', () => {
    it('should have flex layout with gap', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      const { container } = render(<RealtimeStatus />)

      const wrapper = container.firstChild as HTMLElement
      expect(wrapper?.className).toContain('flex')
      expect(wrapper?.className).toContain('gap-2')
    })

    it('should display dot before text', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      const { container } = render(<RealtimeStatus />)

      const children = container.firstChild?.childNodes
      // First child should be the dot div
      expect(children?.[0]?.nodeType).toBe(1) // Element node
      // Second child should be the text span
      expect(children?.[1]?.nodeType).toBe(1) // Element node
    })
  })

  describe('status transitions', () => {
    it('should update when connection status changes', () => {
      (useRealtime as any).mockReturnValue({ connected: false })
      const { rerender } = render(<RealtimeStatus />)

      expect(screen.getByText('Polling')).toBeInTheDocument()

      ;(useRealtime as any).mockReturnValue({ connected: true })
      rerender(<RealtimeStatus />)

      expect(screen.getByText('Live')).toBeInTheDocument()
    })

    it('should change from Live to Polling', () => {
      (useRealtime as any).mockReturnValue({ connected: true })
      const { rerender } = render(<RealtimeStatus />)

      expect(screen.getByText('Live')).toBeInTheDocument()

      ;(useRealtime as any).mockReturnValue({ connected: false })
      rerender(<RealtimeStatus />)

      expect(screen.getByText('Polling')).toBeInTheDocument()
    })
  })

  describe('integration', () => {
    it('should work without additional props', () => {
      (useRealtime as any).mockReturnValue({ connected: true })

      render(<RealtimeStatus />)

      expect(screen.getByText('Live')).toBeInTheDocument()
    })

    it('should render successfully with useRealtime hook', () => {
      (useRealtime as any).mockReturnValue({
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
        error: null,
        subscribe: vi.fn(),
        unsubscribe: vi.fn(),
      })

      render(<RealtimeStatus />)

      expect(screen.getByText('Live')).toBeInTheDocument()
    })
  })

  describe('timestamp display', () => {
    it('should not display timestamp by default', () => {
      (useRealtime as any).mockReturnValue({
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
      })

      render(<RealtimeStatus />)

      const timeString = new Date('2024-03-13T12:00:00Z').toLocaleTimeString()
      expect(screen.queryByText(timeString)).not.toBeInTheDocument()
    })

    it('should display timestamp when showTimestamp prop is true', () => {
      (useRealtime as any).mockReturnValue({
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
      })

      render(<RealtimeStatus showTimestamp={true} />)

      const timeString = new Date('2024-03-13T12:00:00Z').toLocaleTimeString()
      expect(screen.getByText(timeString)).toBeInTheDocument()
    })

    it('should not display timestamp when lastUpdate is null', () => {
      (useRealtime as any).mockReturnValue({
        connected: true,
        lastUpdate: null,
      })

      render(<RealtimeStatus showTimestamp={true} />)

      // Should not throw, just not display timestamp
      expect(screen.getByText('Live')).toBeInTheDocument()
    })

    it('should display timestamp with correct styling', () => {
      (useRealtime as any).mockReturnValue({
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
      })

      render(<RealtimeStatus showTimestamp={true} />)

      const timeString = new Date('2024-03-13T12:00:00Z').toLocaleTimeString()
      const timestampElement = screen.getByText(timeString)

      expect(timestampElement.className).toContain('text-xs')
      expect(timestampElement.className).toContain('text-slate-500')
    })

    it('should display timestamp with dark mode support', () => {
      (useRealtime as any).mockReturnValue({
        connected: true,
        lastUpdate: '2024-03-13T12:00:00Z',
      })

      render(<RealtimeStatus showTimestamp={true} />)

      const timeString = new Date('2024-03-13T12:00:00Z').toLocaleTimeString()
      const timestampElement = screen.getByText(timeString)

      expect(timestampElement.className).toContain('dark:text-slate-400')
    })

    it('should work with disconnected state and timestamp', () => {
      (useRealtime as any).mockReturnValue({
        connected: false,
        lastUpdate: '2024-03-13T12:00:00Z',
      })

      render(<RealtimeStatus showTimestamp={true} />)

      expect(screen.getByText('Polling')).toBeInTheDocument()
      const timeString = new Date('2024-03-13T12:00:00Z').toLocaleTimeString()
      expect(screen.getByText(timeString)).toBeInTheDocument()
    })
  })
})
