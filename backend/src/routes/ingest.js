const express = require('express');
const { pool } = require('../db');

const router = express.Router();

// POST /api/v1/ingest — receive metrics from agent (uses requireApiKey + ingestLimiter in index)
router.post('/', async (req, res) => {
  const { db_id, timestamp, queries } = req.body;
  const instance = req.dbInstance;

  if (db_id != null && db_id !== '') {
    const sid = String(db_id).trim();
    const matchesId = String(instance.id) === sid;
    const matchesName = instance.name === sid;
    if (!matchesId && !matchesName) {
      console.warn(
        'Ingest db_id does not match instance for this API key (metrics still stored). ' +
          `payload.db_id=${JSON.stringify(sid)} instance={ id: ${instance.id}, name: ${JSON.stringify(instance.name)} }`
      );
    }
  }

  if (!queries || !Array.isArray(queries)) {
    return res.status(400).json({ error: 'Missing or invalid "queries" array' });
  }

  const client = await pool.connect();
  try {
    const recordedAt = timestamp ? new Date(timestamp) : new Date();
    for (const q of queries) {
      await client.query(
        `INSERT INTO query_metrics (db_instance_id, query_hash, avg_time_ms, calls, shared_blks_read, shared_blks_hit, rows_count, recorded_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
        [
          instance.id,
          q.query_hash || '',
          q.avg_time_ms ?? 0,
          q.calls ?? 0,
          q.shared_blks_read ?? 0,
          q.shared_blks_hit ?? 0,
          q.rows ?? 0,
          recordedAt,
        ]
      );
    }
    res.json({ status: 'success', message: 'Metrics recorded' });
  } catch (err) {
    console.error('Ingest error:', err);
    res.status(500).json({ error: 'Failed to store metrics' });
  } finally {
    client.release();
  }
});

module.exports = router;
