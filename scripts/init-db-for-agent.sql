-- Enable pg_stat_statements (required for QueryWise agent).
-- Run once; for Docker Postgres 16+, the extension is available by default.
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Create a read-only user for the agent (optional; you can use the main user for local testing).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'querywise_reader') THEN
    CREATE ROLE querywise_reader WITH LOGIN PASSWORD 'querywise_reader';
  END IF;
END $$;

GRANT CONNECT ON DATABASE querywise TO querywise_reader;
GRANT USAGE ON SCHEMA pg_catalog TO querywise_reader;
GRANT SELECT ON pg_stat_statements TO querywise_reader;
GRANT SELECT ON pg_stat_user_tables TO querywise_reader;
GRANT SELECT ON pg_stat_user_indexes TO querywise_reader;
GRANT SELECT ON pg_index TO querywise_reader;

-- Ensure the main user can also read pg_stat_statements (for testing with user querywise).
GRANT SELECT ON pg_stat_statements TO querywise;
