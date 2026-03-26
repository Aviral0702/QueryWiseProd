// Cost formula aligned with pg_stat_statements: calls / block counters are statement totals
// (avg_time_ms * calls ≈ total exec time; shared_blks_* are totals, not per-call).
const CPU_UNIT_COST = 0.00001; // per ms of total execution time
const IO_READ_COST = 0.0001; // per block read (total)
const IO_HIT_COST = 0.000001; // per cache hit block (total)
const QUERY_OVERHEAD = 0.00001; // per call

function estimateQueryCost(q) {
  const calls = Number(q.calls) || 0;
  const cpuCost = (q.avg_time_ms || 0) * CPU_UNIT_COST * calls;
  const ioReadCost = (q.shared_blks_read || 0) * IO_READ_COST;
  const ioHitCost = (q.shared_blks_hit || 0) * IO_HIT_COST;
  const overheadCost = calls * QUERY_OVERHEAD;
  return cpuCost + ioReadCost + ioHitCost + overheadCost;
}

/**
 * Deduplicate cumulative snapshots: keep the latest row per query_hash by recorded_at.
 * Then rank by estimated_cost (desc).
 */
function rankLatestSnapshots(rows) {
  const byHash = new Map();
  for (const r of rows) {
    const key = r.query_hash;
    const t = r.recorded_at ? new Date(r.recorded_at).getTime() : 0;
    const prev = byHash.get(key);
    if (!prev || t > new Date(prev.recorded_at).getTime()) {
      byHash.set(key, r);
    }
  }
  const list = [];
  for (const rec of byHash.values()) {
    const row = {
      query_hash: rec.query_hash,
      avg_time_ms: Number(rec.avg_time_ms) || 0,
      calls: Number(rec.calls) || 0,
      shared_blks_read: Number(rec.shared_blks_read) || 0,
      shared_blks_hit: Number(rec.shared_blks_hit) || 0,
      rows_count: Number(rec.rows_count) || 0,
    };
    row.estimated_cost = estimateQueryCost(row);
    list.push(row);
  }
  list.sort((a, b) => b.estimated_cost - a.estimated_cost);
  return list;
}

/** @deprecated Use rankLatestSnapshots — kept for tests or callers expecting the old name */
function aggregateAndRank(rows) {
  return rankLatestSnapshots(rows);
}

module.exports = { estimateQueryCost, rankLatestSnapshots, aggregateAndRank };
