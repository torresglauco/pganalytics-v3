-- Rename RDS tables to Managed Instance
-- This migration renames all RDS-related tables to use "managed_instance" nomenclature

SET search_path TO pganalytics, public;

-- ============================================================================
-- RENAME TABLES
-- ============================================================================

-- Rename main RDS instances table
ALTER TABLE IF EXISTS rds_instances RENAME TO managed_instances;
ALTER TABLE IF EXISTS rds_databases RENAME TO managed_instance_databases;
ALTER TABLE IF EXISTS rds_metrics RENAME TO managed_instance_metrics;
ALTER TABLE IF EXISTS rds_performance_insights RENAME TO managed_instance_performance_insights;
ALTER TABLE IF EXISTS rds_backup_events RENAME TO managed_instance_backup_events;
ALTER TABLE IF EXISTS rds_maintenance_history RENAME TO managed_instance_maintenance_history;
ALTER TABLE IF EXISTS rds_monitoring_jobs RENAME TO managed_instance_monitoring_jobs;

-- ============================================================================
-- RENAME INDEXES
-- ============================================================================

-- Rename managed_instances indexes
ALTER INDEX IF EXISTS idx_rds_instances_region RENAME TO idx_managed_instances_region;
ALTER INDEX IF EXISTS idx_rds_instances_active RENAME TO idx_managed_instances_active;
ALTER INDEX IF EXISTS idx_rds_instances_environment RENAME TO idx_managed_instances_environment;
ALTER INDEX IF EXISTS idx_rds_instances_status RENAME TO idx_managed_instances_status;

-- Rename managed_instance_databases indexes
ALTER INDEX IF EXISTS idx_rds_databases_instance RENAME TO idx_managed_instance_databases_instance;
ALTER INDEX IF EXISTS idx_rds_databases_name RENAME TO idx_managed_instance_databases_name;
ALTER INDEX IF EXISTS idx_rds_databases_instance_name RENAME TO idx_managed_instance_databases_instance_name;

-- Rename managed_instance_metrics indexes
ALTER INDEX IF EXISTS idx_rds_metrics_instance_time RENAME TO idx_managed_instance_metrics_instance_time;
ALTER INDEX IF EXISTS idx_rds_metrics_type_time RENAME TO idx_managed_instance_metrics_type_time;

-- Rename managed_instance_performance_insights indexes
ALTER INDEX IF EXISTS idx_rds_perf_insights_instance RENAME TO idx_managed_instance_perf_insights_instance;
ALTER INDEX IF EXISTS idx_rds_perf_insights_period RENAME TO idx_managed_instance_perf_insights_period;

-- Rename managed_instance_backup_events indexes
ALTER INDEX IF EXISTS idx_rds_backups_instance RENAME TO idx_managed_instance_backups_instance;
ALTER INDEX IF EXISTS idx_rds_backups_type RENAME TO idx_managed_instance_backups_type;

-- Rename managed_instance_maintenance_history indexes
ALTER INDEX IF EXISTS idx_rds_maint_instance RENAME TO idx_managed_instance_maint_instance;
ALTER INDEX IF EXISTS idx_rds_maint_status RENAME TO idx_managed_instance_maint_status;

-- Rename managed_instance_monitoring_jobs indexes
ALTER INDEX IF EXISTS idx_rds_jobs_instance RENAME TO idx_managed_instance_jobs_instance;
ALTER INDEX IF EXISTS idx_rds_jobs_next_run RENAME TO idx_managed_instance_jobs_next_run;

-- ============================================================================
-- RENAME FOREIGN KEY CONSTRAINTS
-- ============================================================================

-- Rename constraints in managed_instance_databases
ALTER TABLE managed_instance_databases
  DROP CONSTRAINT IF EXISTS rds_databases_rds_instance_id_fkey,
  ADD CONSTRAINT managed_instance_databases_managed_instance_id_fkey
    FOREIGN KEY (rds_instance_id) REFERENCES managed_instances(id) ON DELETE CASCADE;

-- Rename constraints in managed_instance_metrics
ALTER TABLE managed_instance_metrics
  DROP CONSTRAINT IF EXISTS rds_metrics_rds_instance_id_fkey,
  ADD CONSTRAINT managed_instance_metrics_managed_instance_id_fkey
    FOREIGN KEY (rds_instance_id) REFERENCES managed_instances(id) ON DELETE CASCADE;

-- Rename constraints in managed_instance_performance_insights
ALTER TABLE managed_instance_performance_insights
  DROP CONSTRAINT IF EXISTS rds_performance_insights_rds_instance_id_fkey,
  ADD CONSTRAINT managed_instance_performance_insights_managed_instance_id_fkey
    FOREIGN KEY (rds_instance_id) REFERENCES managed_instances(id) ON DELETE CASCADE;

-- Rename constraints in managed_instance_backup_events
ALTER TABLE managed_instance_backup_events
  DROP CONSTRAINT IF EXISTS rds_backup_events_rds_instance_id_fkey,
  ADD CONSTRAINT managed_instance_backup_events_managed_instance_id_fkey
    FOREIGN KEY (rds_instance_id) REFERENCES managed_instances(id) ON DELETE CASCADE;

-- Rename constraints in managed_instance_maintenance_history
ALTER TABLE managed_instance_maintenance_history
  DROP CONSTRAINT IF EXISTS rds_maintenance_history_rds_instance_id_fkey,
  ADD CONSTRAINT managed_instance_maintenance_history_managed_instance_id_fkey
    FOREIGN KEY (rds_instance_id) REFERENCES managed_instances(id) ON DELETE CASCADE;

-- Rename constraints in managed_instance_monitoring_jobs
ALTER TABLE managed_instance_monitoring_jobs
  DROP CONSTRAINT IF EXISTS rds_monitoring_jobs_rds_instance_id_fkey,
  ADD CONSTRAINT managed_instance_monitoring_jobs_managed_instance_id_fkey
    FOREIGN KEY (rds_instance_id) REFERENCES managed_instances(id) ON DELETE CASCADE;

-- ============================================================================
-- RENAME COLUMNS
-- ============================================================================

-- Rename rds_instance_id columns to managed_instance_id
ALTER TABLE managed_instance_databases
  RENAME COLUMN rds_instance_id TO managed_instance_id;

ALTER TABLE managed_instance_metrics
  RENAME COLUMN rds_instance_id TO managed_instance_id;

ALTER TABLE managed_instance_performance_insights
  RENAME COLUMN rds_instance_id TO managed_instance_id;

ALTER TABLE managed_instance_backup_events
  RENAME COLUMN rds_instance_id TO managed_instance_id;

ALTER TABLE managed_instance_maintenance_history
  RENAME COLUMN rds_instance_id TO managed_instance_id;

ALTER TABLE managed_instance_monitoring_jobs
  RENAME COLUMN rds_instance_id TO managed_instance_id;

-- Update managed_instances table column naming
ALTER TABLE managed_instances
  RENAME COLUMN rds_endpoint TO endpoint;

-- ============================================================================
-- RENAME CLUSTER TABLES
-- ============================================================================

ALTER TABLE IF EXISTS rds_clusters RENAME TO managed_instance_clusters;
ALTER TABLE IF EXISTS idx_rds_clusters_status RENAME TO idx_managed_instance_clusters_status;

-- ============================================================================
-- SCHEMA MIGRATION TRACKING
-- ============================================================================

INSERT INTO schema_versions (version, description) VALUES
    ('009_rename_rds_to_managed_instance', 'Rename all RDS tables to Managed Instance nomenclature')
ON CONFLICT DO NOTHING;
