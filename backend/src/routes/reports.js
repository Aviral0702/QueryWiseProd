const express = require('express');
const { pool } = require('../db');
const { rankLatestSnapshots } = require('../services/costEngine');

const router = express.Router();

function authUserId(req) {
  if (!req.user?.id) return null;
  const id = parseInt(req.user.id, 10);
  return Number.isNaN(id) ? null : id;
}

async function assertInstanceAccess(req, dbInstanceId) {
  const uid = authUserId(req);
  if (uid == null) return true;
  const r = await pool.query(
    'SELECT owner_user_id FROM db_instances WHERE id = $1',
    [dbInstanceId]
  );
  if (r.rows.length === 0) return false;
  const owner = r.rows[0].owner_user_id;
  if (owner == null) return false;
  return Number(owner) === uid;
}

// GET /api/v1/reports/:db_id — cost report for dashboard (db_id = instance name or id)
router.get('/:db_id', async (req, res) => {
  const { db_id } = req.params;
  const limit = Math.min(parseInt(req.query.limit, 10) || 50, 200);
  const since = req.query.since; // optional ISO date

  let dbInstanceId;
  const byId = await pool.query('SELECT id FROM db_instances WHERE id = $1', [db_id]);
  if (byId.rows.length > 0) {
    dbInstanceId = byId.rows[0].id;
  } else {
    const byName = await pool.query('SELECT id FROM db_instances WHERE name = $1', [db_id]);
    if (byName.rows.length === 0) {
      return res.status(404).json({ error: 'Database instance not found' });
    }
    dbInstanceId = byName.rows[0].id;
  }

  const allowed = await assertInstanceAccess(req, dbInstanceId);
  if (!allowed) {
    return res.status(404).json({ error: 'Database instance not found' });
  }

  const sinceDate = since ? new Date(since) : null;
  const metricsSql = sinceDate
    ? `SELECT DISTINCT ON (query_hash)
         query_hash, avg_time_ms, calls, shared_blks_read, shared_blks_hit, rows_count, recorded_at
       FROM query_metrics
       WHERE db_instance_id = $1 AND recorded_at >= $2
       ORDER BY query_hash, recorded_at DESC`
    : `SELECT DISTINCT ON (query_hash)
         query_hash, avg_time_ms, calls, shared_blks_read, shared_blks_hit, rows_count, recorded_at
       FROM query_metrics
       WHERE db_instance_id = $1
       ORDER BY query_hash, recorded_at DESC`;

  const rows = sinceDate
    ? await pool.query(metricsSql, [dbInstanceId, sinceDate])
    : await pool.query(metricsSql, [dbInstanceId]);

  const ranked = rankLatestSnapshots(rows.rows).slice(0, limit);

  const trendSince = sinceDate && !Number.isNaN(sinceDate.getTime()) ? sinceDate : new Date(Date.now() - 24 * 60 * 60 * 1000);
  const trendGranularity =
    (Date.now() - trendSince.getTime()) / (60 * 60 * 1000) <= 48 ? 'hour' : 'day';

  const trendRows = await pool.query(
    `WITH per_bucket AS (
       SELECT date_trunc($1::text, recorded_at) AS bucket,
              query_hash,
              avg_time_ms,
              calls,
              recorded_at,
              ROW_NUMBER() OVER (
                PARTITION BY date_trunc($1::text, recorded_at), query_hash
                ORDER BY recorded_at DESC
              ) AS rn
       FROM query_metrics
       WHERE db_instance_id = $2 AND recorded_at >= $3
     )
     SELECT bucket,
            AVG(avg_time_ms) AS avg_ms,
            SUM(calls) AS calls
     FROM per_bucket
     WHERE rn = 1
     GROUP BY bucket
     ORDER BY bucket`,
    [trendGranularity, dbInstanceId, trendSince]
  );

  const avgResponseTimeTrend = trendRows.rows.map((r) => ({
    time: r.bucket,
    avg_time_ms: parseFloat(r.avg_ms) || 0,
    calls: parseInt(r.calls, 10) || 0,
  }));

  const totalCost = ranked.reduce((s, q) => s + (q.estimated_cost || 0), 0);
  const totalCalls = ranked.reduce((s, q) => s + (q.calls || 0), 0);
  const avgTimeOverall =
    totalCalls > 0
      ? ranked.reduce((s, q) => s + (q.avg_time_ms || 0) * (q.calls || 0), 0) / totalCalls
      : 0;

  const lastUpdatedRow = await pool.query(
    'SELECT MAX(recorded_at) AS last_updated FROM query_metrics WHERE db_instance_id = $1',
    [dbInstanceId]
  );
  const last_updated = lastUpdatedRow.rows[0]?.last_updated?.toISOString?.() ?? null;

  res.json({
    db_id: db_id,
    top_expensive_queries: ranked,
    total_estimated_cost: totalCost,
    avg_response_time_ms: avgTimeOverall,
    avg_response_time_trend: avgResponseTimeTrend,
    last_updated,
  });
});

module.exports = router;
