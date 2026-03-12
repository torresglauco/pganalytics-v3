-- Triggers for automatic timestamp updates
-- These must run AFTER the schema is created (000_complete_schema.sql)

SET search_path TO pganalytics, public;

-- Create the update timestamp function FIRST
CREATE OR REPLACE FUNCTION pganalytics.update_updated_at_column() RETURNS TRIGGER AS 'BEGIN NEW.updated_at := CURRENT_TIMESTAMP; RETURN NEW; END;' LANGUAGE plpgsql;

-- Drop triggers if they exist (to allow re-running migrations)
DROP TRIGGER IF EXISTS trigger_users_updated_at ON pganalytics.users;
DROP TRIGGER IF EXISTS trigger_collectors_updated_at ON pganalytics.collectors;
DROP TRIGGER IF EXISTS trigger_servers_updated_at ON pganalytics.servers;
DROP TRIGGER IF EXISTS trigger_postgresql_instances_updated_at ON pganalytics.postgresql_instances;
DROP TRIGGER IF EXISTS trigger_databases_updated_at ON pganalytics.databases;
DROP TRIGGER IF EXISTS trigger_secrets_updated_at ON pganalytics.secrets;
DROP TRIGGER IF EXISTS trigger_alert_rules_updated_at ON pganalytics.alert_rules;
DROP TRIGGER IF EXISTS trigger_registration_secrets_updated_at ON pganalytics.registration_secrets;
DROP TRIGGER IF EXISTS trigger_managed_instances_updated_at ON pganalytics.managed_instances;
DROP TRIGGER IF EXISTS trigger_managed_instance_databases_updated_at ON pganalytics.managed_instance_databases;

-- Create triggers
CREATE TRIGGER trigger_users_updated_at BEFORE UPDATE ON pganalytics.users
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_collectors_updated_at BEFORE UPDATE ON pganalytics.collectors
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_servers_updated_at BEFORE UPDATE ON pganalytics.servers
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_postgresql_instances_updated_at BEFORE UPDATE ON pganalytics.postgresql_instances
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_databases_updated_at BEFORE UPDATE ON pganalytics.databases
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_secrets_updated_at BEFORE UPDATE ON pganalytics.secrets
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_alert_rules_updated_at BEFORE UPDATE ON pganalytics.alert_rules
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_registration_secrets_updated_at BEFORE UPDATE ON pganalytics.registration_secrets
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_managed_instances_updated_at BEFORE UPDATE ON pganalytics.managed_instances
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();

CREATE TRIGGER trigger_managed_instance_databases_updated_at BEFORE UPDATE ON pganalytics.managed_instance_databases
    FOR EACH ROW EXECUTE FUNCTION pganalytics.update_updated_at_column();
