const express = require('express');
const { pool } = require('../db');
const { aggregateAndRank } = require('../services/costEngine');

const router = express.Router();

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

  let rows;
  if (since) {
    rows = await pool.query(
      `SELECT query_hash, avg_time_ms, calls, shared_blks_read, shared_blks_hit, rows_count
       FROM query_metrics WHERE db_instance_id = $1 AND recorded_at >= $2
       ORDER BY recorded_at DESC`,
      [dbInstanceId, new Date(since)]
    );
  } else {
    rows = await pool.query(
      `SELECT query_hash, avg_time_ms, calls, shared_blks_read, shared_blks_hit, rows_count
       FROM query_metrics WHERE db_instance_id = $1
       ORDER BY recorded_at DESC LIMIT 5000`,
      [dbInstanceId]
    );
  }

  const ranked = aggregateAndRank(rows.rows).slice(0, limit);

  const trendSince = since ? new Date(since) : new Date(Date.now() - 24 * 60 * 60 * 1000);
  const trendIntervalHours = (Date.now() - trendSince.getTime()) / (60 * 60 * 1000) <= 48 ? 'hour' : 'day';
  const trendRows = await pool.query(
    `SELECT date_trunc($1, recorded_at) AS bucket,
            AVG(avg_time_ms) AS avg_ms,
            SUM(calls) AS calls
     FROM query_metrics
     WHERE db_instance_id = $2 AND recorded_at >= $3
     GROUP BY date_trunc($1, recorded_at)
     ORDER BY bucket`,
    [trendIntervalHours, dbInstanceId, trendSince]
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
