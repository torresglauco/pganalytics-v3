import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ConditionPreview } from './ConditionPreview'
import type { AlertCondition } from '../../types/alerts'

describe('ConditionPreview', () => {
  it('renders error count condition correctly', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'error_count',
      operator: '>',
      threshold: 5,
      timeWindow: 10,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.getByText(/Error Count/)).toBeInTheDocument()
    expect(screen.getByText(/>/)).toBeInTheDocument()
    expect(screen.getByText(/5/)).toBeInTheDocument()
    expect(screen.getByText(/in last 10 minutes/)).toBeInTheDocument()
  })

  it('renders slow query count condition correctly', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'slow_query_count',
      operator: '<',
      threshold: 2,
      timeWindow: 5,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.getByText(/Slow Query Count/)).toBeInTheDocument()
    expect(screen.getByText(/</)).toBeInTheDocument()
  })

  it('renders connection count condition correctly', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'connection_count',
      operator: '>=',
      threshold: 100,
      timeWindow: 15,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.getByText(/Connection Count/)).toBeInTheDocument()
  })

  it('renders cache hit ratio with percentage format', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'cache_hit_ratio',
      operator: '<',
      threshold: 0.5,
      timeWindow: 10,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.getByText(/Cache Hit Ratio/)).toBeInTheDocument()
    expect(screen.getByText(/50.0%/)).toBeInTheDocument()
  })

  it('displays duration when provided', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'error_count',
      operator: '>',
      threshold: 5,
      timeWindow: 10,
      duration: 5,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.getByText(/for 5 minutes/)).toBeInTheDocument()
  })

  it('does not display duration when not provided', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'error_count',
      operator: '>',
      threshold: 5,
      timeWindow: 10,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.queryByText(/for \d+ minute/)).not.toBeInTheDocument()
  })

  it('handles != operator correctly', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'error_count',
      operator: '!=',
      threshold: 0,
      timeWindow: 10,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.getByText(/!=/)).toBeInTheDocument()
  })

  it('displays minute singular form correctly', () => {
    const condition: AlertCondition = {
      id: '1',
      metricType: 'error_count',
      operator: '>',
      threshold: 5,
      timeWindow: 1,
    }

    render(<ConditionPreview condition={condition} />)

    expect(screen.getByText(/in last 1 minute/)).toBeInTheDocument()
  })
})
