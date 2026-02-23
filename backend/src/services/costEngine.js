// Cost formula: (cpu_time * CPU_UNIT_COST) + (io_reads * IO_UNIT_COST) + (execution_count * QUERY_OVERHEAD)
// Using block-based I/O: 1 block = 8KB. We approximate cost by shared_blks_read (disk) and shared_blks_hit (cache).
const CPU_UNIT_COST = 0.00001;   // per ms of execution
const IO_READ_COST = 0.0001;     // per block read from disk
const IO_HIT_COST = 0.000001;    // per block from cache
const QUERY_OVERHEAD = 0.00001;  // per call

function estimateQueryCost(q) {
  const cpuCost = (q.avg_time_ms || 0) * CPU_UNIT_COST * (q.calls || 0);
  const ioReadCost = (q.shared_blks_read || 0) * IO_READ_COST * (q.calls || 0);
  const ioHitCost = (q.shared_blks_hit || 0) * IO_HIT_COST * (q.calls || 0);
  const overheadCost = (q.calls || 0) * QUERY_OVERHEAD;
  return cpuCost + ioReadCost + ioHitCost + overheadCost;
}

function aggregateAndRank(rows) {
  const byHash = new Map();
  for (const r of rows) {
    const key = r.query_hash;
    if (!byHash.has(key)) {
      byHash.set(key, {
        query_hash: r.query_hash,
        avg_time_ms: 0,
        calls: 0,
        shared_blks_read: 0,
        shared_blks_hit: 0,
        rows_count: 0,
        total_time_ms: 0,
      });
    }
    const rec = byHash.get(key);
    rec.calls += Number(r.calls);
    rec.shared_blks_read += Number(r.shared_blks_read);
    rec.shared_blks_hit += Number(r.shared_blks_hit);
    rec.rows_count += Number(r.rows_count);
    rec.total_time_ms += (r.avg_time_ms || 0) * (r.calls || 0);
  }
  const list = [];
  for (const rec of byHash.values()) {
    rec.avg_time_ms = rec.calls > 0 ? rec.total_time_ms / rec.calls : 0;
    rec.estimated_cost = estimateQueryCost(rec);
    list.push(rec);
  }
  list.sort((a, b) => b.estimated_cost - a.estimated_cost);
  return list;
}

module.exports = { estimateQueryCost, aggregateAndRank };
