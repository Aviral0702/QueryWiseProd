package report

import "querywise/pkg/types"

// ShortHint returns a non-LLM hint for table display.
func ShortHint(q types.ScoredQuery) string {
	if q.TempBlksRead > 0 || q.TempBlksWritten > 0 {
		if q.MeanExecTimeMs > 25 {
			return "High avg time + temp I/O — check sorts/hash joins and missing indexes"
		}
		return "Temp block usage — large sorts or hash ops; review work_mem / indexes"
	}
	if q.CacheHitRatio < 0.55 && (q.SharedBlksHit+q.SharedBlksRead) > 100 {
		return "Low cache hit ratio — likely seq scans; review indexes and statistics"
	}
	if q.Calls > 50000 && q.MeanExecTimeMs > 5 {
		return "High frequency with moderate latency — consider caching or statement tuning"
	}
	if q.MeanExecTimeMs > 100 {
		return "High average execution time — inspect plan and I/O for hot paths"
	}
	return "Review execution plan and wait events for this pattern"
}
