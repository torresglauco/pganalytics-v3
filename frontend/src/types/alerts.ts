export type AlertSeverity = 'critical' | 'warning' | 'info';
export type AlertStatus = 'active' | 'resolved' | 'muted';
export type AlertType =
  | 'lock_contention'
  | 'table_bloat'
  | 'cache_miss'
  | 'connection_pool'
  | 'idle_transaction'
  | 'replication_lag'
  | 'metrics_collection_failure';

export interface Alert {
  id: string;
  collector_id: string;
  alert_type: AlertType;
  severity: AlertSeverity;
  status: AlertStatus;
  title: string;
  description: string;
  value?: number;
  threshold?: number;
  unit?: string;
  fired_at: Date;
  resolved_at?: Date;
  incident_id?: string;
  runbook_link?: string;
}

export interface Incident {
  id: string;
  group_name: string;
  state: 'active' | 'acknowledged' | 'resolved';
  severity: AlertSeverity;
  alerts: Alert[];
  root_cause?: string;
  confidence: number;
  suggested_actions: string[];
  created_at: Date;
  updated_at: Date;
  resolved_at?: Date;
}

export interface SuppressionRule {
  id: string;
  name: string;
  alert_type: AlertType;
  collector_id?: string;
  enabled: boolean;
  time_based?: {
    start_hour: number;
    end_hour: number;
    days: string[];
  };
}
