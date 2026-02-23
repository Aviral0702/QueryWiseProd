require('dotenv').config();
const { Pool } = require('pg');
const fs = require('fs');
const path = require('path');

const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: parseInt(process.env.DB_PORT || '5432', 10),
  user: process.env.DB_USER || 'querywise',
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME || 'querywise',
});

const SQL = `
-- Users (Google OAuth SSO)
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  google_id TEXT UNIQUE NOT NULL,
  email TEXT NOT NULL,
  name TEXT,
  avatar_url TEXT,
  created_at TIMESTAMP DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);

-- DB instances (one per agent / monitored database)
CREATE TABLE IF NOT EXISTS db_instances (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  api_key TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT now()
);

-- Raw query metrics from ingest
CREATE TABLE IF NOT EXISTS query_metrics (
  id SERIAL PRIMARY KEY,
  db_instance_id INTEGER NOT NULL REFERENCES db_instances(id) ON DELETE CASCADE,
  query_hash TEXT NOT NULL,
  avg_time_ms FLOAT NOT NULL,
  calls BIGINT NOT NULL,
  shared_blks_read BIGINT NOT NULL DEFAULT 0,
  shared_blks_hit BIGINT NOT NULL DEFAULT 0,
  rows_count BIGINT NOT NULL DEFAULT 0,
  recorded_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_query_metrics_db_instance_id ON query_metrics(db_instance_id);
CREATE INDEX IF NOT EXISTS idx_query_metrics_recorded_at ON query_metrics(recorded_at);
CREATE INDEX IF NOT EXISTS idx_query_metrics_db_recorded ON query_metrics(db_instance_id, recorded_at DESC);
`;

async function migrate() {
  const client = await pool.connect();
  try {
    await client.query(SQL);
    console.log('Migrations completed.');
  } catch (err) {
    console.error('Migration failed:', err);
    process.exit(1);
  } finally {
    client.release();
    await pool.end();
  }
}

migrate();
