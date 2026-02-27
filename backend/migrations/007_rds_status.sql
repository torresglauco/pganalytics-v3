-- Add status column to RDS instances
-- Status values: 'registering', 'registered', 'monitoring', 'paused'

SET search_path TO pganalytics, public;

-- Add status column to rds_instances table
ALTER TABLE rds_instances ADD COLUMN status VARCHAR(50) DEFAULT 'registered' CHECK (status IN ('registering', 'registered', 'monitoring', 'paused'));

-- Create index for efficient status queries
CREATE INDEX idx_rds_instances_monitoring_status ON rds_instances(status);

-- Update schema versions
INSERT INTO schema_versions (version, description) VALUES
    ('007_rds_status', 'Add status column to RDS instances with values: registering, registered, monitoring, paused')
ON CONFLICT DO NOTHING;
