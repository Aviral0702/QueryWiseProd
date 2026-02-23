const { Pool } = require('pg');

const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: parseInt(process.env.DB_PORT || '5432', 10),
  user: process.env.DB_USER || 'querywise',
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME || 'querywise',
  max: 20,
  idleTimeoutMillis: 30000,
});

module.exports = { pool };
