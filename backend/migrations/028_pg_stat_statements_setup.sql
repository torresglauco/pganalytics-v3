-- pgAnalytics v3.1.0 - pg_stat_statements Setup Helper
-- This migration creates a helper function to check pg_stat_statements availability
-- and provides setup instructions if not available

SET search_path TO pganalytics, public;

-- Create function to check if pg_stat_statements is available
CREATE OR REPLACE FUNCTION check_pg_stat_statements_available()
RETURNS TABLE(available boolean, message text)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Try to query pg_stat_statements
    BEGIN
        PERFORM 1 FROM pg_stat_statements LIMIT 1;
        RETURN QUERY SELECT true, 'pg_stat_statements is available'::text;
    EXCEPTION WHEN others THEN
        RETURN QUERY SELECT false,
            'pg_stat_statements not available. Add to postgresql.conf: shared_preload_libraries = ''pg_stat_statements'' and restart PostgreSQL.'::text;
    END;
END;
$$;

-- Create view for easy slow query access
CREATE OR REPLACE VIEW slow_queries AS
SELECT
    queryid,
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    min_exec_time,
    max_exec_time,
    rows,
    shared_blks_hit,
    shared_blks_read
FROM pg_stat_statements
ORDER BY mean_exec_time DESC;

-- Comment on the view
COMMENT ON VIEW slow_queries IS 'Pre-sorted view of pg_stat_statements for slow query analysis';