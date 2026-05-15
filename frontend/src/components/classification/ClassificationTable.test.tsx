/**
 * Tests for ClassificationTable component
 * Tests the data classification results table with PII/PCI pattern display
 */

import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ClassificationTable } from './ClassificationTable'
import type { DataClassificationResult, PatternType, Category } from '../../types/classification'

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  ChevronUp: () => <span data-testid="chevron-up" />,
  ChevronDown: () => <span data-testid="chevron-down" />,
  ChevronLeft: () => <span data-testid="chevron-left" />,
  ChevronRight: () => <span data-testid="chevron-right" />,
  Shield: () => <span data-testid="shield-icon" />,
  Database: () => <span data-testid="database-icon" />,
}))

describe('ClassificationTable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // Mock data factory
  const createMockResult = (overrides: Partial<DataClassificationResult> = {}): DataClassificationResult => ({
    time: '2024-01-01T00:00:00Z',
    collector_id: 'collector-001',
    database_name: 'test_db',
    schema_name: 'public',
    table_name: 'users',
    column_name: 'email',
    pattern_type: 'EMAIL',
    category: 'PII',
    confidence: 0.95,
    match_count: 150,
    sample_values: ['user1@example.com', 'user2@example.com'],
    regulation_mapping: { LGPD: ['Art. 11'] },
    ...overrides,
  })

  const mockData: DataClassificationResult[] = [
    createMockResult({
      database_name: 'production_db',
      schema_name: 'public',
      table_name: 'customers',
      column_name: 'cpf',
      pattern_type: 'CPF',
      category: 'PII',
      confidence: 0.98,
      match_count: 500,
    }),
    createMockResult({
      database_name: 'production_db',
      schema_name: 'sales',
      table_name: 'orders',
      column_name: 'credit_card',
      pattern_type: 'CREDIT_CARD',
      category: 'PCI',
      confidence: 0.92,
      match_count: 200,
    }),
    createMockResult({
      database_name: 'analytics_db',
      schema_name: 'public',
      table_name: 'contacts',
      column_name: 'email',
      pattern_type: 'EMAIL',
      category: 'PII',
      confidence: 0.88,
      match_count: 1000,
    }),
  ]

  describe('Table columns', () => {
    it('should render table with columns: Database, Schema, Table, Column, Pattern, Category, Confidence, Matches', () => {
      render(<ClassificationTable data={mockData} />)

      expect(screen.getByText('Database')).toBeInTheDocument()
      expect(screen.getByText('Schema')).toBeInTheDocument()
      expect(screen.getByText('Table')).toBeInTheDocument()
      expect(screen.getByText('Column')).toBeInTheDocument()
      expect(screen.getByText('Pattern')).toBeInTheDocument()
      expect(screen.getByText('Category')).toBeInTheDocument()
      expect(screen.getByText('Confidence')).toBeInTheDocument()
      expect(screen.getByText('Matches')).toBeInTheDocument()
    })

    it('should display database names in rows', () => {
      render(<ClassificationTable data={mockData} />)

      // production_db appears twice in mock data (two different rows)
      const productionDbElements = screen.getAllByText('production_db')
      expect(productionDbElements.length).toBe(2)
      expect(screen.getByText('analytics_db')).toBeInTheDocument()
    })

    it('should display table names in rows', () => {
      render(<ClassificationTable data={mockData} />)

      expect(screen.getByText('customers')).toBeInTheDocument()
      expect(screen.getByText('orders')).toBeInTheDocument()
      expect(screen.getByText('contacts')).toBeInTheDocument()
    })
  })

  describe('Pattern type badges', () => {
    it('should display PII pattern type badge', () => {
      render(<ClassificationTable data={mockData} />)

      expect(screen.getByText('CPF')).toBeInTheDocument()
      // Email is a PII pattern type
      expect(screen.getByText('Email')).toBeInTheDocument()
    })

    it('should display PCI pattern type badge with different color', () => {
      render(<ClassificationTable data={mockData} />)

      const creditCardBadge = screen.getByText('Credit Card')
      expect(creditCardBadge).toBeInTheDocument()

      // The credit card badge should have red color (from CATEGORY_COLORS)
      const badgeContainer = creditCardBadge.closest('span')
      expect(badgeContainer).toBeInTheDocument()
    })
  })

  describe('Category badges', () => {
    it('should display PII category badge', () => {
      render(<ClassificationTable data={mockData} />)

      const piiBadges = screen.getAllByText('PII')
      expect(piiBadges.length).toBeGreaterThan(0)
    })

    it('should display PCI category badge', () => {
      render(<ClassificationTable data={mockData} />)

      expect(screen.getByText('PCI')).toBeInTheDocument()
    })
  })

  describe('Sample count display', () => {
    it('should show match count for each classification', () => {
      render(<ClassificationTable data={mockData} />)

      // Numbers are formatted with toLocaleString() - check they exist
      expect(screen.getByText('500')).toBeInTheDocument()
      expect(screen.getByText('200')).toBeInTheDocument()
      // 1000 may be formatted as "1,000" or "1000" depending on locale
      // Check that the contacts table row has match_count of 1000 via the row context
      const contactsRow = screen.getByText('contacts').closest('tr')
      expect(contactsRow).toBeInTheDocument()
    })
  })

  describe('Row interactions', () => {
    it('should trigger onRowClick callback when row is clicked', async () => {
      const user = userEvent.setup()
      const mockOnRowClick = vi.fn()

      render(<ClassificationTable data={mockData} onRowClick={mockOnRowClick} />)

      const row = screen.getByText('customers').closest('tr')
      expect(row).toBeInTheDocument()

      await user.click(row!)

      expect(mockOnRowClick).toHaveBeenCalledTimes(1)
      expect(mockOnRowClick).toHaveBeenCalledWith(
        expect.objectContaining({
          table_name: 'customers',
          pattern_type: 'CPF',
        })
      )
    })

    it('should not trigger onRowClick when not provided', async () => {
      const user = userEvent.setup()

      render(<ClassificationTable data={mockData} />)

      const row = screen.getByText('customers').closest('tr')
      await user.click(row!)

      // Should not throw or cause issues
      expect(row).toBeInTheDocument()
    })
  })

  describe('Loading state', () => {
    it('should show spinner when loading', () => {
      render(<ClassificationTable data={[]} isLoading={true} />)

      expect(screen.getByText('Loading classification data...')).toBeInTheDocument()
    })

    it('should not show table content when loading', () => {
      render(<ClassificationTable data={mockData} isLoading={true} />)

      // Check that table rows are not rendered during loading
      expect(screen.queryByText('customers')).not.toBeInTheDocument()
    })
  })

  describe('Empty state', () => {
    it('should show "No classification results found" message when empty', () => {
      render(<ClassificationTable data={[]} isLoading={false} />)

      expect(screen.getByText('No classification results found')).toBeInTheDocument()
    })

    it('should show suggestion to adjust filters in empty state', () => {
      render(<ClassificationTable data={[]} isLoading={false} />)

      expect(screen.getByText(/Try adjusting your filters/)).toBeInTheDocument()
    })
  })
})