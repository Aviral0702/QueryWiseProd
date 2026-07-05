package db

// QueryStat is a sanitized row from pg_stat_statements. Raw SQL is never retained.
type QueryStat struct {
	QueryID *int64
	// QueryHash is SHA-256 hex of the normalized fingerprint from the server side.
	QueryHash         string
	Calls             int64
	TotalExecTimeMs   float64
	MeanExecTimeMs    float64
	StddevExecTimeMs  float64
	MinExecTimeMs     float64
	MaxExecTimeMs     float64
	Rows              int64
	SharedBlksHit     int64
	SharedBlksRead    int64
	SharedBlksWritten int64
	TempBlksRead      int64
	TempBlksWritten   int64
	// InsufficientPrivilege is true when the server returned
	// <insufficient privilege> as the query text for this row.
	InsufficientPrivilege bool
}
