import React, { useMemo } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  flexRender,
  createColumnDef,
  type SortingState,
} from '@tanstack/react-table';
import { ChevronUp, ChevronDown, ChevronLeft, ChevronRight, Shield, Database } from 'lucide-react';
import type { DataClassificationResult, PatternType, Category } from '../../types/classification';
import { PATTERN_COLORS, CATEGORY_COLORS, PATTERN_LABELS, CATEGORY_LABELS } from '../../types/classification';

interface ClassificationTableProps {
  data: DataClassificationResult[];
  isLoading?: boolean;
  onRowClick?: (row: DataClassificationResult) => void;
}

export const ClassificationTable: React.FC<ClassificationTableProps> = ({
  data,
  isLoading = false,
  onRowClick,
}) => {
  const [sorting, setSorting] = React.useState<SortingState>([]);

  const getPatternBadgeStyle = (patternType: PatternType): string => {
    const color = PATTERN_COLORS[patternType] || '#6b7280';
    return `background-color: ${color}20; color: ${color}; border: 1px solid ${color}40;`;
  };

  const getCategoryBadgeStyle = (category: Category): string => {
    const color = CATEGORY_COLORS[category] || '#6b7280';
    return `background-color: ${color}20; color: ${color}; border: 1px solid ${color}40;`;
  };

  const getConfidenceColor = (confidence: number): string => {
    if (confidence >= 0.9) return 'text-green-600';
    if (confidence >= 0.7) return 'text-yellow-600';
    return 'text-red-600';
  };

  const columns = useMemo(() => [
    {
      accessorKey: 'database_name',
      header: 'Database',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <Database size={14} className="text-gray-400" />
          <span className="font-medium">{row.original.database_name}</span>
        </div>
      ),
    },
    {
      accessorKey: 'schema_name',
      header: 'Schema',
    },
    {
      accessorKey: 'table_name',
      header: 'Table',
    },
    {
      accessorKey: 'column_name',
      header: 'Column',
      cell: ({ row }) => (
        <span className="font-mono text-sm">{row.original.column_name}</span>
      ),
    },
    {
      accessorKey: 'pattern_type',
      header: 'Pattern',
      cell: ({ row }) => {
        const patternType = row.original.pattern_type as PatternType;
        return (
          <span
            className="px-2 py-1 rounded text-xs font-medium"
            style={getPatternBadgeStyle(patternType)}
          >
            {PATTERN_LABELS[patternType] || patternType}
          </span>
        );
      },
    },
    {
      accessorKey: 'category',
      header: 'Category',
      cell: ({ row }) => {
        const category = row.original.category as Category;
        return (
          <span
            className="px-2 py-1 rounded text-xs font-medium"
            style={getCategoryBadgeStyle(category)}
          >
            {CATEGORY_LABELS[category] || category}
          </span>
        );
      },
    },
    {
      accessorKey: 'confidence',
      header: 'Confidence',
      cell: ({ row }) => {
        const confidence = row.original.confidence;
        return (
          <span className={`font-medium ${getConfidenceColor(confidence)}`}>
            {(confidence * 100).toFixed(1)}%
          </span>
        );
      },
    },
    {
      accessorKey: 'match_count',
      header: 'Matches',
      cell: ({ row }) => (
        <span className="font-mono text-sm">
          {row.original.match_count.toLocaleString()}
        </span>
      ),
    },
  ], []);

  const table = useReactTable({
    data,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: {
      pagination: { pageSize: 10 },
    },
  });

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-8">
        <div className="flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <span className="ml-3 text-gray-600">Loading classification data...</span>
        </div>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-8">
        <div className="flex flex-col items-center justify-center text-gray-500">
          <Shield size={48} className="text-gray-300 mb-4" />
          <p className="text-lg font-medium">No classification results found</p>
          <p className="text-sm">Try adjusting your filters or run a new classification scan.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50 border-b border-gray-200">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    className="px-4 py-3 text-left text-sm font-medium text-gray-700"
                  >
                    {header.isPlaceholder ? null : (
                      <div
                        className={`flex items-center gap-2 ${
                          header.column.getCanSort() ? 'cursor-pointer select-none' : ''
                        }`}
                        onClick={header.column.getToggleSortingHandler()}
                      >
                        {flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                        {header.column.getIsSorted() === 'asc' && (
                          <ChevronUp size={14} />
                        )}
                        {header.column.getIsSorted() === 'desc' && (
                          <ChevronDown size={14} />
                        )}
                      </div>
                    )}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody className="divide-y divide-gray-200">
            {table.getRowModel().rows.map((row) => (
              <tr
                key={row.id}
                className={`hover:bg-gray-50 ${onRowClick ? 'cursor-pointer' : ''}`}
                onClick={() => onRowClick?.(row.original)}
              >
                {row.getVisibleCells().map((cell) => (
                  <td key={cell.id} className="px-4 py-3 text-sm text-gray-900">
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200">
        <div className="flex items-center gap-2 text-sm text-gray-600">
          <span>
            Showing {table.getState().pagination.pageIndex * table.getState().pagination.pageSize + 1} to{' '}
            {Math.min(
              (table.getState().pagination.pageIndex + 1) * table.getState().pagination.pageSize,
              data.length
            )}{' '}
            of {data.length} results
          </span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
            className="p-2 rounded border border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            <ChevronLeft size={16} />
          </button>
          <span className="text-sm text-gray-600">
            Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
          </span>
          <button
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
            className="p-2 rounded border border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            <ChevronRight size={16} />
          </button>
        </div>
      </div>
    </div>
  );
};

export default ClassificationTable;