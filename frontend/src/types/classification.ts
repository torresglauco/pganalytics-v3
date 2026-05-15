/**
 * Data Classification Type Definitions
 * Types for PII/PCI detection results and classification reports
 * Mirrors backend/pkg/models/classification_models.go
 */

/**
 * Pattern types for data detection
 */
export type PatternType = 'CPF' | 'CNPJ' | 'EMAIL' | 'PHONE' | 'CREDIT_CARD' | 'CUSTOM';

/**
 * Category types for classification
 */
export type Category = 'PII' | 'PCI' | 'SENSITIVE' | 'CUSTOM';

/**
 * Data classification result from PII/PCI detection
 * Represents a detected sensitive data pattern in a column
 */
export interface DataClassificationResult {
  time: string;
  collector_id: string;
  database_name: string;
  schema_name: string;
  table_name: string;
  column_name: string;
  pattern_type: PatternType;
  category: Category;
  confidence: number; // 0.0 to 1.0
  match_count: number; // Number of matching rows
  sample_values: string[]; // Masked sample values (up to 5)
  regulation_mapping: Record<string, string[]>; // LGPD/GDPR article references
}

/**
 * Custom detection pattern for user-defined sensitive data
 */
export interface CustomPattern {
  id: number;
  tenant_id: string | null; // NULL for global patterns
  pattern_name: string;
  pattern_regex: string;
  category: Category;
  validation_algorithm: 'Luhn' | 'Mod11' | 'None';
  description: string;
  enabled: boolean;
  created_at?: string;
  updated_at?: string;
}

/**
 * Classification summary for a single database
 */
export interface DatabaseClassificationSummary {
  database_name: string;
  schema_count: number;
  table_count: number;
  pii_column_count: number;
  pci_column_count: number;
  sensitive_count: number;
  highest_risk_table?: string;
}

/**
 * Aggregated classification report response
 */
export interface ClassificationReportResponse {
  total_databases: number;
  total_tables: number;
  total_columns: number;
  pii_columns: number;
  pci_columns: number;
  sensitive_columns: number;
  custom_columns: number;
  pattern_breakdown: Record<PatternType, number>;
  category_breakdown: Record<Category, number>;
  database_summary?: DatabaseClassificationSummary[];
}

/**
 * Filter parameters for classification results query
 */
export interface ClassificationFilter {
  database?: string;
  schema?: string;
  table?: string;
  pattern_type?: PatternType;
  category?: Category;
  time_range?: '1h' | '24h' | '7d' | '30d';
  limit?: number;
  offset?: number;
}

/**
 * Response wrapper for classification metrics
 */
export interface ClassificationMetricsResponse {
  metric_type: string;
  count: number;
  time_range: string;
  data: DataClassificationResult[];
}

/**
 * Response wrapper for custom patterns
 */
export interface CustomPatternResponse {
  count: number;
  patterns: CustomPattern[];
}

/**
 * Pattern color mapping for UI display
 */
export const PATTERN_COLORS: Record<PatternType, string> = {
  CPF: '#f59e0b',      // amber
  CNPJ: '#10b981',     // emerald
  EMAIL: '#3b82f6',    // blue
  PHONE: '#8b5cf6',    // violet
  CREDIT_CARD: '#ef4444', // red
  CUSTOM: '#6b7280',   // gray
};

/**
 * Category color mapping for UI display
 */
export const CATEGORY_COLORS: Record<Category, string> = {
  PII: '#3b82f6',      // blue
  PCI: '#ef4444',      // red
  SENSITIVE: '#f59e0b', // amber
  CUSTOM: '#6b7280',   // gray
};

/**
 * Pattern type display labels
 */
export const PATTERN_LABELS: Record<PatternType, string> = {
  CPF: 'CPF',
  CNPJ: 'CNPJ',
  EMAIL: 'Email',
  PHONE: 'Phone',
  CREDIT_CARD: 'Credit Card',
  CUSTOM: 'Custom',
};

/**
 * Category display labels
 */
export const CATEGORY_LABELS: Record<Category, string> = {
  PII: 'PII',
  PCI: 'PCI',
  SENSITIVE: 'Sensitive',
  CUSTOM: 'Custom',
};