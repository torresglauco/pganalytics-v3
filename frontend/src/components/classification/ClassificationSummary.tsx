import React from 'react';
import { Database, Table, Columns, Shield, CreditCard, AlertTriangle } from 'lucide-react';
import type { ClassificationReportResponse } from '../../types/classification';

interface ClassificationSummaryProps {
  report: ClassificationReportResponse | null;
  isLoading?: boolean;
  onFilterClick?: (filterType: string, value?: string) => void;
}

interface SummaryCardProps {
  title: string;
  value: number;
  icon: React.ReactNode;
  color: string;
  onClick?: () => void;
}

const SummaryCard: React.FC<SummaryCardProps> = ({
  title,
  value,
  icon,
  color,
  onClick,
}) => (
  <div
    className={`bg-white rounded-lg border border-gray-200 p-4 ${
      onClick ? 'cursor-pointer hover:border-gray-300 hover:shadow-sm transition-all' : ''
    }`}
    onClick={onClick}
  >
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm text-gray-600">{title}</p>
        <p className={`text-2xl font-bold mt-1 ${color}`}>
          {value.toLocaleString()}
        </p>
      </div>
      <div className={`p-3 rounded-lg ${color.replace('text-', 'bg-').replace('-600', '-100')}`}>
        {icon}
      </div>
    </div>
  </div>
);

export const ClassificationSummary: React.FC<ClassificationSummaryProps> = ({
  report,
  isLoading = false,
  onFilterClick,
}) => {
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="bg-white rounded-lg border border-gray-200 p-4 animate-pulse">
            <div className="h-4 bg-gray-200 rounded w-20 mb-2"></div>
            <div className="h-8 bg-gray-200 rounded w-16"></div>
          </div>
        ))}
      </div>
    );
  }

  if (!report) {
    return null;
  }

  const cards = [
    {
      title: 'Total Databases',
      value: report.total_databases,
      icon: <Database size={20} className="text-blue-600" />,
      color: 'text-blue-600',
      filterType: 'database',
    },
    {
      title: 'Total Tables',
      value: report.total_tables,
      icon: <Table size={20} className="text-green-600" />,
      color: 'text-green-600',
      filterType: 'table',
    },
    {
      title: 'Total Columns',
      value: report.total_columns,
      icon: <Columns size={20} className="text-gray-600" />,
      color: 'text-gray-600',
      filterType: undefined,
    },
    {
      title: 'PII Columns',
      value: report.pii_columns,
      icon: <Shield size={20} className="text-blue-600" />,
      color: 'text-blue-600',
      filterType: 'category',
      filterValue: 'PII',
    },
    {
      title: 'PCI Columns',
      value: report.pci_columns,
      icon: <CreditCard size={20} className="text-red-600" />,
      color: 'text-red-600',
      filterType: 'category',
      filterValue: 'PCI',
    },
    {
      title: 'Sensitive Columns',
      value: report.sensitive_columns,
      icon: <AlertTriangle size={20} className="text-amber-600" />,
      color: 'text-amber-600',
      filterType: 'category',
      filterValue: 'SENSITIVE',
    },
  ];

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
      {cards.map((card, index) => (
        <SummaryCard
          key={index}
          title={card.title}
          value={card.value}
          icon={card.icon}
          color={card.color}
          onClick={
            card.filterType && onFilterClick
              ? () => onFilterClick(card.filterType!, card.filterValue)
              : undefined
          }
        />
      ))}
    </div>
  );
};

export default ClassificationSummary;