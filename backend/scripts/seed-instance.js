require('dotenv').config();
const { Pool } = require('pg');
const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: parseInt(process.env.DB_PORT || '5432', 10),
  user: process.env.DB_USER || 'querywise',
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME || 'querywise',
});

async function seed() {
  const name = process.argv[2] || 'default-instance';
  const r = await pool.query(
    'INSERT INTO db_instances (name, api_key) VALUES ($1, encode(gen_random_bytes(24), \'hex\')) RETURNING id, name, api_key',
    [name]
  );
  console.log('Created instance:', r.rows[0]);
  console.log('Use this API key in your agent config (backend.api_key):', r.rows[0].api_key);
  await pool.end();
}

seed().catch((e) => { console.error(e); process.exit(1); });
