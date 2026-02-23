package collector

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// QueryMetric represents one row from pg_stat_statements (anonymized).
type QueryMetric struct {
	QueryHash      string  `json:"query_hash"`
	AvgTimeMs      float64 `json:"avg_time_ms"`
	Calls          int64   `json:"calls"`
	SharedBlksRead int64   `json:"shared_blks_read"`
	SharedBlksHit  int64   `json:"shared_blks_hit"`
	Rows           int64   `json:"rows,omitempty"`
}

// IngestPayload is the JSON sent to the backend.
type IngestPayload struct {
	DBID      string        `json:"db_id"`
	Timestamp string        `json:"timestamp"`
	Queries   []QueryMetric `json:"queries"`
}

type Collector struct {
	db     *sql.DB
	dbID   string
	limit  int
	timeout time.Duration
}

func New(dsn string, dbID string, batchSize int, timeoutSec int) (*Collector, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	to := time.Duration(timeoutSec) * time.Second
	if to <= 0 {
		to = 30 * time.Second
	}
	return &Collector{
		db:     db,
		dbID:   dbID,
		limit:  batchSize,
		timeout: to,
	}, nil
}

func hashQuery(query string) string {
	h := sha256.Sum256([]byte(query))
	return hex.EncodeToString(h[:])
}

// Collect reads from pg_stat_statements and returns an IngestPayload.
// Query text is hashed so no raw SQL is sent.
func (c *Collector) Collect(ctx context.Context) (*IngestPayload, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// pg_stat_statements: queryid is optional; we use hashed query for privacy
	query := `
		SELECT
			query,
			CASE WHEN calls > 0 THEN total_exec_time / calls ELSE 0 END AS mean_exec_time_ms,
			calls,
			COALESCE(shared_blks_read, 0),
			COALESCE(shared_blks_hit, 0),
			COALESCE(pg_stat_statements.rows, 0)
		FROM pg_stat_statements
		WHERE query NOT LIKE '%%pg_stat_statements%%'
		ORDER BY total_exec_time DESC NULLS LAST
		LIMIT $1
	`
	rows, err := c.db.QueryContext(ctx, query, c.limit)
	if err != nil {
		return nil, fmt.Errorf("pg_stat_statements query: %w", err)
	}
	defer rows.Close()

	var queries []QueryMetric
	for rows.Next() {
		var rawQuery string
		var meanMs float64
		var calls, blksRead, blksHit, rowsCount int64
		if err := rows.Scan(&rawQuery, &meanMs, &calls, &blksRead, &blksHit, &rowsCount); err != nil {
			return nil, err
		}
		queries = append(queries, QueryMetric{
			QueryHash:      hashQuery(rawQuery),
			AvgTimeMs:      meanMs,
			Calls:          calls,
			SharedBlksRead: blksRead,
			SharedBlksHit:  blksHit,
			Rows:           rowsCount,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &IngestPayload{
		DBID:      c.dbID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Queries:   queries,
	}, nil
}

func (c *Collector) Close() error {
	return c.db.Close()
}
