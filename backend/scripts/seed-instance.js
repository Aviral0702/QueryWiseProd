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
  const ownerEmail = process.env.OWNER_USER_EMAIL || process.env.QUERYWISE_OWNER_EMAIL;

  let ownerUserId = null;
  if (ownerEmail) {
    const u = await pool.query('SELECT id FROM users WHERE lower(email) = lower($1) LIMIT 1', [ownerEmail.trim()]);
    if (u.rows.length === 0) {
      console.error(`No user with email "${ownerEmail}". Sign in once with Google first, or omit OWNER_USER_EMAIL.`);
      await pool.end();
      process.exit(1);
    }
    ownerUserId = u.rows[0].id;
  }

  const r = await pool.query(
    `INSERT INTO db_instances (name, api_key, owner_user_id)
     VALUES ($1, encode(gen_random_bytes(24), 'hex'), $2)
     RETURNING id, name, api_key, owner_user_id`,
    [name, ownerUserId]
  );
  console.log('Created instance:', r.rows[0]);
  console.log('Use this API key in your agent config (backend.api_key):', r.rows[0].api_key);
  if (!ownerUserId) {
    console.log(
      'Note: With Google auth enabled, this instance has no owner until you set owner_user_id or create it via POST /api/v1/instances while logged in.'
    );
  }
  await pool.end();
}

seed().catch((e) => {
  console.error(e);
  process.exit(1);
});
