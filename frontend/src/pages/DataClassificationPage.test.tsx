/**
 * Tests for DataClassificationPage component
 * Tests the data classification page with loading, error, filters, and summary cards
 */

import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { DataClassificationPage } from './DataClassificationPage'
import * as classificationApi from '../api/classificationApi'
import type { ClassificationReportResponse, DataClassificationResult } from '../types/classification'

// Mock the API module
vi.mock('../api/classificationApi', () => ({
  getClassificationResults: vi.fn(),
  getClassificationReport: vi.fn(),
}))

// Mock child components
vi.mock('../components/classification/ClassificationTable', () => ({
  default: ({ data, isLoading, onRowClick }: any) => (
    <div data-testid="classification-table">
      {isLoading ? 'Loading...' : `${data?.length || 0} results`}
    </div>
  ),
}))

vi.mock('../components/classification/ClassificationFilters', () => ({
  default: ({ onReset }: any) => (
    <div data-testid="classification-filters">
      <button onClick={onReset} data-testid="reset-filters">Reset</button>
    </div>
  ),
}))

vi.mock('../components/classification/ClassificationSummary', () => ({
  default: ({ report, isLoading }: any) => (
    <div data-testid="classification-summary">
      {isLoading ? 'Loading summary...' : `PII: ${report?.pii_columns || 0}, PCI: ${report?.pci_columns || 0}`}
    </div>
  ),
}))

vi.mock('../components/classification/PatternBreakdownChart', () => ({
  default: ({ data, isLoading }: any) => (
    <div data-testid="pattern-breakdown-chart">
      {isLoading ? 'Loading chart...' : `Patterns: ${Object.keys(data || {}).length}`}
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
  Shield: () => <span data-testid="shield-icon" />,
  Database: () => <span data-testid="database-icon" />,
  RefreshCw: ({ className }: { className?: string }) => (
    <span data-testid="refresh-icon" className={className} />
  ),
  Download: () => <span data-testid="download-icon" />,
  ChevronRight: () => <span data-testid="chevron-right" />,
  AlertCircle: () => <span data-testid="alert-icon" />,
}))

describe('DataClassificationPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Mock data factory
  const createMockResults = (): DataClassificationResult[] => [
    {
      time: '2024-01-01T00:00:00Z',
      collector_id: 'collector-001',
      database_name: 'production_db',
      schema_name: 'public',
      table_name: 'users',
      column_name: 'email',
      pattern_type: 'EMAIL',
      category: 'PII',
      confidence: 0.95,
      match_count: 150,
      sample_values: ['user@example.com'],
      regulation_mapping: {},
    },
    {
      time: '2024-01-01T00:00:00Z',
      collector_id: 'collector-001',
      database_name: 'production_db',
      schema_name: 'public',
      table_name: 'customers',
      column_name: 'cpf',
      pattern_type: 'CPF',
      category: 'PII',
      confidence: 0.98,
      match_count: 500,
      sample_values: ['***.456.789-**'],
      regulation_mapping: {},
    },
  ]

  const createMockReport = (): ClassificationReportResponse => ({
    total_databases: 2,
    total_tables: 10,
    total_columns: 50,
    pii_columns: 15,
    pci_columns: 5,
    sensitive_columns: 8,
    custom_columns: 0,
    pattern_breakdown: {
      CPF: 10,
      EMAIL: 15,
      PHONE: 5,
      CREDIT_CARD: 5,
      CNPJ: 0,
      CUSTOM: 0,
    },
    category_breakdown: {
      PII: 20,
      PCI: 5,
      SENSITIVE: 3,
      CUSTOM: 0,
    },
  })

  // Helper to render with router
  const renderWithRouter = (initialEntries: string[] = ['/collectors/test-collector/classification']) => {
    return render(
      <MemoryRouter initialEntries={initialEntries}>
        <Routes>
          <Route path="/collectors/:collectorId/classification" element={<DataClassificationPage />} />
        </Routes>
      </MemoryRouter>
    )
  }

  describe('Loading state', () => {
    it('should show loading spinner initially', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockImplementation(() => new Promise(() => {}))
      vi.mocked(classificationApi.getClassificationReport).mockImplementation(() => new Promise(() => {}))

      renderWithRouter()

      expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    })
  })

  describe('Data loaded state', () => {
    it('should render classification table after data loads', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('classification-table')).toBeInTheDocument()
      })
    })

    it('should display summary cards with pattern counts', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('classification-summary')).toBeInTheDocument()
        expect(screen.getByText(/PII: 15/)).toBeInTheDocument()
        expect(screen.getByText(/PCI: 5/)).toBeInTheDocument()
      })
    })

    it('should display pattern breakdown chart with data', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('pattern-breakdown-chart')).toBeInTheDocument()
      })
    })

    it('should display page header with title', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('Data Classification')).toBeInTheDocument()
      })
    })
  })

  describe('Filters', () => {
    it('should display classification filters component', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('classification-filters')).toBeInTheDocument()
      })
    })

    it('should reset filters when reset button clicked', async () => {
      const user = userEvent.setup()
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByTestId('reset-filters')).toBeInTheDocument()
      })

      await user.click(screen.getByTestId('reset-filters'))

      // Should have called the API again after filter reset
      await waitFor(() => {
        expect(classificationApi.getClassificationResults).toHaveBeenCalled()
      })
    })
  })

  describe('Breadcrumb navigation', () => {
    it('should display breadcrumbs for drill-down', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('All Databases')).toBeInTheDocument()
      })
    })
  })

  describe('Error handling', () => {
    it('should display error message on API failure', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockRejectedValue(new Error('API Error'))
      vi.mocked(classificationApi.getClassificationReport).mockRejectedValue(new Error('API Error'))

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('API Error')).toBeInTheDocument()
      })
    })
  })

  describe('Export button', () => {
    it('should have export button visible', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('Export')).toBeInTheDocument()
      })
    })
  })

  describe('Refresh functionality', () => {
    it('should have refresh button visible', async () => {
      vi.mocked(classificationApi.getClassificationResults).mockResolvedValue({
        metric_type: 'classification',
        count: 2,
        time_range: '24h',
        data: createMockResults(),
      })
      vi.mocked(classificationApi.getClassificationReport).mockResolvedValue(createMockReport())

      renderWithRouter()

      await waitFor(() => {
        expect(screen.getByText('Refresh')).toBeInTheDocument()
      })
    })
  })
})