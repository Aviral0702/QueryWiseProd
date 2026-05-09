package db

import (
	"context"
	"fmt"

	"querywise/internal/hash"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgStatStatementsQuery = `
SELECT
	queryid,
	query,
	calls,
	total_exec_time,
	mean_exec_time,
	"rows",
	shared_blks_hit,
	shared_blks_read,
	shared_blks_written,
	temp_blks_read,
	temp_blks_written
FROM pg_stat_statements
WHERE calls >= $1
ORDER BY calls DESC`

// FetchStatements loads pg_stat_statements rows above minCalls into QueryStat structs.
// Query text from the catalog is hashed and discarded immediately; it is never returned.
func FetchStatements(ctx context.Context, pool *pgxpool.Pool, minCalls int64) ([]QueryStat, error) {
	rows, err := pool.Query(ctx, pgStatStatementsQuery, minCalls)
	if err != nil {
		return nil, fmt.Errorf("query pg_stat_statements: %w", err)
	}
	defer rows.Close()

	var out []QueryStat
	for rows.Next() {
		var qs QueryStat
		var rawQuery string
		var queryID pgtype.Int8

		err := rows.Scan(
			&queryID,
			&rawQuery,
			&qs.Calls,
			&qs.TotalExecTimeMs,
			&qs.MeanExecTimeMs,
			&qs.Rows,
			&qs.SharedBlksHit,
			&qs.SharedBlksRead,
			&qs.SharedBlksWritten,
			&qs.TempBlksRead,
			&qs.TempBlksWritten,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pg_stat_statements row: %w", err)
		}

		if queryID.Valid {
			v := queryID.Int64
			qs.QueryID = &v
		}

		qs.QueryHash = hash.SHA256Hex(rawQuery)
		rawQuery = ""
		out = append(out, qs)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return out, nil
}
