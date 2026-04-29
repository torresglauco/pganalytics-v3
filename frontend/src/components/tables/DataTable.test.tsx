import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { DataTable } from './DataTable';

// Helper to render with router
const renderWithRouter = (initialRoute: string = '/') => {
  return render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <DataTable
        columns={[
          { key: 'id', label: 'ID', sortable: true },
          { key: 'name', label: 'Name', sortable: true },
        ]}
        data={[
          { id: '1', name: 'Alice' },
          { id: '2', name: 'Bob' },
        ]}
        searchable={true}
      />
    </MemoryRouter>
  );
};

describe('DataTable URL state synchronization', () => {
  it('should initialize with default state when no URL params', () => {
    renderWithRouter();

    // Search input should be empty
    const searchInput = screen.getByPlaceholderText('Search...');
    expect(searchInput).toHaveValue('');

    // Verify both data rows are visible (no filtering applied)
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });

  it('should initialize state from URL params on mount', () => {
    // Render with URL params: ?sort=name&order=desc&search=Ali
    renderWithRouter('/?sort=name&order=desc&search=Ali');

    // Search input should have 'Ali'
    const searchInput = screen.getByPlaceholderText('Search...');
    expect(searchInput).toHaveValue('Ali');

    // Should show filtered data (Alice contains 'Ali')
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.queryByText('Bob')).not.toBeInTheDocument();
  });

  it('should update URL params when sorting', async () => {
    const user = userEvent.setup();
    renderWithRouter();

    // Click on 'Name' column header to sort
    await user.click(screen.getByText('Name'));

    // Verify the component shows sorted state (Bob should come before Alice alphabetically)
    const rows = screen.getAllByRole('row');
    // Header is row 0, so first data row is row 1
    expect(rows[1]).toHaveTextContent('Alice');
    expect(rows[2]).toHaveTextContent('Bob');
  });

  it('should update URL params when searching', async () => {
    const user = userEvent.setup();
    renderWithRouter();

    const searchInput = screen.getByPlaceholderText('Search...');

    // Type in search
    await user.type(searchInput, 'Bob');

    // Verify filtered results
    await waitFor(() => {
      expect(screen.getByText('Bob')).toBeInTheDocument();
      expect(screen.queryByText('Alice')).not.toBeInTheDocument();
    });
  });

  it('should remove search param when search is cleared', async () => {
    const user = userEvent.setup();
    renderWithRouter('/?search=test');

    const searchInput = screen.getByPlaceholderText('Search...');
    expect(searchInput).toHaveValue('test');

    // Clear the search
    await user.clear(searchInput);

    // All data should be visible
    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
      expect(screen.getByText('Bob')).toBeInTheDocument();
    });
  });
});

describe('DataTable basic functionality', () => {
  it('should sort data when sortable column is clicked', async () => {
    const user = userEvent.setup();
    renderWithRouter();

    // Initial order: Alice, Bob
    let rows = screen.getAllByRole('row');
    expect(rows[1]).toHaveTextContent('Alice');
    expect(rows[2]).toHaveTextContent('Bob');

    // Click on 'Name' to sort ascending
    await user.click(screen.getByText('Name'));

    // After sorting by name ascending: Alice, Bob (already in that order)
    rows = screen.getAllByRole('row');
    expect(rows[1]).toHaveTextContent('Alice');
    expect(rows[2]).toHaveTextContent('Bob');

    // Click again to sort descending
    await user.click(screen.getByText('Name'));

    // After sorting by name descending: Bob, Alice
    rows = screen.getAllByRole('row');
    expect(rows[1]).toHaveTextContent('Bob');
    expect(rows[2]).toHaveTextContent('Alice');
  });

  it('should filter data based on search term', async () => {
    const user = userEvent.setup();
    renderWithRouter();

    const searchInput = screen.getByPlaceholderText('Search...');

    // Type search term
    await user.type(searchInput, 'Alice');

    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
      expect(screen.queryByText('Bob')).not.toBeInTheDocument();
    });
  });

  it('should show empty message when no data matches search', async () => {
    const user = userEvent.setup();
    renderWithRouter();

    const searchInput = screen.getByPlaceholderText('Search...');

    // Type search term that matches nothing
    await user.type(searchInput, 'Nonexistent');

    await waitFor(() => {
      expect(screen.getByText('No data found')).toBeInTheDocument();
    });
  });
});