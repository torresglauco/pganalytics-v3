import React, { useState, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ChevronUp, ChevronDown, Search } from 'lucide-react';

export interface Column<T> {
  key: keyof T;
  label: string;
  sortable?: boolean;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
  width?: string;
}

interface DataTableProps<T extends { id: string }> {
  columns: Column<T>[];
  data: T[];
  loading?: boolean;
  searchable?: boolean;
  selectable?: boolean;
  onRowClick?: (row: T) => void;
  title?: string;
  emptyMessage?: string;
}

export const DataTable = <T extends { id: string }>({
  columns,
  data,
  loading = false,
  searchable = true,
  selectable = false,
  onRowClick,
  title,
  emptyMessage = 'No data found',
}: DataTableProps<T>) => {
  // URL state synchronization for persistent filters/sort across navigation
  const [searchParams, setSearchParams] = useSearchParams();

  // Sort key from URL or null
  const sortKey = (searchParams.get('sort') as keyof T) || null;

  // Sort order from URL or default 'asc'
  const sortOrder = (searchParams.get('order') as 'asc' | 'desc') || 'asc';

  // Search term from URL or empty string
  const searchTerm = searchParams.get('search') || '';

  // selectedRows stays as useState (not URL-backed, transient selection)
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());

  // Filter data
  const filteredData = useMemo(() => {
    let result = data;

    if (searchTerm) {
      result = result.filter(row =>
        JSON.stringify(row).toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    return result;
  }, [data, searchTerm]);

  // Sort data
  const sortedData = useMemo(() => {
    if (!sortKey) return filteredData;

    return [...filteredData].sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];

      if (aVal === bVal) return 0;
      if (aVal === null || aVal === undefined) return 1;
      if (bVal === null || bVal === undefined) return -1;

      const comparison = aVal > bVal ? 1 : -1;
      return sortOrder === 'asc' ? comparison : -comparison;
    });
  }, [filteredData, sortKey, sortOrder]);

  const toggleSort = (key: keyof T) => {
    const newParams = new URLSearchParams(searchParams);

    if (sortKey === key) {
      // Toggle order if same column
      const newOrder = sortOrder === 'asc' ? 'desc' : 'asc';
      newParams.set('order', newOrder);
    } else {
      // New column, set sort and default to asc
      newParams.set('sort', String(key));
      newParams.set('order', 'asc');
    }

    setSearchParams(newParams, { replace: true });
  };

  const toggleAllRows = () => {
    if (selectedRows.size === sortedData.length) {
      setSelectedRows(new Set());
    } else {
      setSelectedRows(new Set(sortedData.map(row => row.id)));
    }
  };

  const toggleRow = (id: string) => {
    const newSelected = new Set(selectedRows);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedRows(newSelected);
  };

  if (loading) {
    return (
      <div className="space-y-2 p-6">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-12 bg-pg-slate/5 rounded animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Title and Search */}
      {title && <h3 className="text-lg font-semibold text-pg-dark">{title}</h3>}

      {searchable && (
        <div className="relative">
          <Search className="absolute left-3 top-3 w-4 h-4 text-pg-slate" />
          <input
            type="text"
            placeholder="Search..."
            value={searchTerm}
            onChange={(e) => {
              const value = e.target.value;
              const newParams = new URLSearchParams(searchParams);
              if (value) {
                newParams.set('search', value);
              } else {
                newParams.delete('search');
              }
              setSearchParams(newParams, { replace: true });
            }}
            className="w-full pl-10 pr-4 py-2 border border-pg-slate/20 rounded-lg focus:outline-none focus:ring-2 focus:ring-pg-cyan"
          />
        </div>
      )}

      {/* Table */}
      <div className="border border-pg-slate/10 rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-pg-slate/5 border-b border-pg-slate/10">
            <tr>
              {selectable && (
                <th className="w-12 px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    checked={selectedRows.size === sortedData.length && sortedData.length > 0}
                    onChange={toggleAllRows}
                    className="rounded"
                  />
                </th>
              )}
              {columns.map(col => (
                <th
                  key={String(col.key)}
                  className="px-4 py-3 text-left font-semibold text-pg-dark cursor-pointer hover:bg-pg-slate/10 transition"
                  style={{ width: col.width }}
                  onClick={() => col.sortable && toggleSort(col.key)}
                >
                  <div className="flex items-center gap-2">
                    {col.label}
                    {col.sortable && sortKey === col.key && (
                      <>
                        {sortOrder === 'asc' ? (
                          <ChevronUp className="w-4 h-4" />
                        ) : (
                          <ChevronDown className="w-4 h-4" />
                        )}
                      </>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sortedData.length === 0 ? (
              <tr>
                <td colSpan={columns.length + (selectable ? 1 : 0)} className="px-4 py-8 text-center text-pg-slate">
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              sortedData.map(row => (
                <tr
                  key={row.id}
                  className="border-b border-pg-slate/5 hover:bg-pg-slate/5 transition cursor-pointer"
                  onClick={() => onRowClick?.(row)}
                >
                  {selectable && (
                    <td className="w-12 px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedRows.has(row.id)}
                        onChange={() => toggleRow(row.id)}
                        onClick={(e) => e.stopPropagation()}
                        className="rounded"
                      />
                    </td>
                  )}
                  {columns.map(col => (
                    <td key={String(col.key)} className="px-4 py-3 text-pg-dark">
                      {col.render
                        ? col.render(row[col.key], row)
                        : String(row[col.key])}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
