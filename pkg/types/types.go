package types

import "time"

// ScoredQuery is one ranked row in a report.
type ScoredQuery struct {
	Rank             int     `json:"rank"`
	QueryHash        string  `json:"query_hash"`
	CostScore        float64 `json:"cost_score"`
	Calls            int64   `json:"calls"`
	MeanExecTimeMs   float64 `json:"mean_exec_time_ms"`
	TotalExecTimeMs  float64 `json:"total_exec_time_ms"`
	StddevExecTimeMs float64 `json:"stddev_exec_time_ms"`
	MinExecTimeMs    float64 `json:"min_exec_time_ms"`
	MaxExecTimeMs    float64 `json:"max_exec_time_ms"`
	SharedBlksRead   int64   `json:"shared_blks_read"`
	SharedBlksHit    int64   `json:"shared_blks_hit"`
	TempBlksRead     int64   `json:"temp_blks_read"`
	TempBlksWritten  int64   `json:"temp_blks_written"`
	Rows             int64   `json:"rows"`
	CacheHitRatio    float64 `json:"cache_hit_ratio"`
	Recommendation   string  `json:"recommendation,omitempty"`
}

// Report is the full analysis output (terminal, markdown, JSON).
type Report struct {
	DatabaseName    string        `json:"database_name"`
	HostEndpoint    string        `json:"host_endpoint"`
	GeneratedAt     time.Time     `json:"generated_at"`
	QueriesAnalyzed int           `json:"queries_analyzed"`
	TopNShown       int           `json:"top_n_shown"`
	CostCoveragePct float64       `json:"cost_coverage_pct"`
	LLMUsed         bool          `json:"llm_used,omitempty"`
	Rows            []ScoredQuery `json:"rows"`
	Warnings        []string      `json:"warnings,omitempty"`
}

// QueryContext is sent to the LLM (no raw SQL).
type QueryContext struct {
	Rank             int     `json:"rank"`
	QueryHash        string  `json:"query_hash"`
	CostScore        float64 `json:"cost_score"`
	Calls            float64 `json:"calls"`
	MeanExecTimeMs   float64 `json:"mean_exec_time_ms"`
	TotalExecTimeMs  float64 `json:"total_exec_time_ms"`
	StddevExecTimeMs float64 `json:"stddev_exec_time_ms"`
	MaxExecTimeMs    float64 `json:"max_exec_time_ms"`
	SharedBlksRead   float64 `json:"shared_blks_read"`
	SharedBlksHit    float64 `json:"shared_blks_hit"`
	TempBlksRead     float64 `json:"temp_blks_read"`
	Rows             float64 `json:"rows"`
	CacheHitRatio    float64 `json:"cache_hit_ratio"`
}
