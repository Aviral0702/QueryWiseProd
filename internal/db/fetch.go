package db

import (
	"context"
	"errors"
	"fmt"

	"querywise/internal/hash"

	"github.com/jackc/pgx/v5/pgconn"
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
	stddev_exec_time,
	min_exec_time,
	max_exec_time,
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
func FetchStatements(ctx context.Context, pool *pgxpool.Pool, minCalls int64, hashKey string) ([]QueryStat, error) {
	rows, err := pool.Query(ctx, pgStatStatementsQuery, minCalls)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, fmt.Errorf("pg_stat_statements is not available: install it with 'CREATE EXTENSION pg_stat_statements;' and add it to shared_preload_libraries: %w", err)
		}
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
			&qs.StddevExecTimeMs,
			&qs.MinExecTimeMs,
			&qs.MaxExecTimeMs,
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

		if rawQuery == "<insufficient privilege>" {
			qs.InsufficientPrivilege = true
			qs.QueryHash = "insufficient-privilege"
		} else {
			qs.QueryHash = hash.Fingerprint(hashKey, rawQuery)
		}
		rawQuery = ""
		out = append(out, qs)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return out, nil
}
