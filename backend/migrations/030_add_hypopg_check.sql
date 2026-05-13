-- pgAnalytics v3.3.0 - Hypopg Extension Check
-- Adds tracking for hypopg availability per database
-- Part of Wave 2: Index Intelligence Features

-- Add column to track hypopg availability per database
ALTER TABLE databases ADD COLUMN IF NOT EXISTS hypopg_available BOOLEAN DEFAULT FALSE;

-- Add index to recommendations for faster status queries
CREATE INDEX IF NOT EXISTS idx_recommendations_benefit
    ON index_recommendations(estimated_benefit DESC);