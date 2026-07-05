# QueryWise

Pure Go CLI binary that connects to PostgreSQL, reads **`pg_stat_statements`**, hashes query text locally with SHA‑256 as soon as it is read (only the hash plus numeric metrics are kept in the report or sent to any API — never the raw SQL), ranks patterns with a configurable cost heuristic, optionally asks the Claude API for anonymized tuning hints, and prints or writes a report.

Requires the [**pg_stat_statements**](https://www.postgresql.org/docs/current/pgstatstatements.html) extension on the database you profile.

---

## Prerequisites & setup

**Minimum PostgreSQL version: 13+.** QueryWise reads the `stddev_exec_time`, `min_exec_time`, and `max_exec_time` columns, which are only exposed from PostgreSQL 13 onward.

`pg_stat_statements` ships with PostgreSQL but must be preloaded and created before use:

1. Add it to `shared_preload_libraries` in `postgresql.conf`:

   ```conf
   shared_preload_libraries = 'pg_stat_statements'
   ```

2. **Restart** PostgreSQL — this setting only takes effect at server start.

3. Enable the extension in the target database:

   ```sql
   CREATE EXTENSION pg_stat_statements;
   ```

### Required privileges

To read the **query text** for statements owned by *other* roles, the connecting role needs the `pg_read_all_stats` role (or superuser). Grant it explicitly:

```sql
GRANT pg_read_all_stats TO your_querywise_role;
```

Without it, `pg_stat_statements` returns the literal string `<insufficient privilege>` as the query text for rows the role does not own. QueryWise cannot fingerprint those rows distinctly — it emits a warning and a count of affected rows.

This is in deliberate tension with the "use a least‑privileged DB user" guidance: the *one* elevated grant QueryWise needs is `pg_read_all_stats`, and nothing more. Grant that specifically rather than reaching for superuser.

---

## Commands

```bash
# Analyze and print colored table report
QUERYWISE_DSN="postgres://user:pass@localhost:5432/mydb" ./bin/querywise analyze

# Analyze with explicit flags
./bin/querywise analyze --dsn "postgres://..." --top 20

# Markdown export
./bin/querywise analyze --dsn "postgres://..." --output markdown --file report.md

# JSON export
./bin/querywise analyze --dsn "postgres://..." --output json --file report.json

# LLM recommendations (sets report footer when used)
./bin/querywise analyze --dsn "postgres://..." --recommend # needs ANTHROPIC_API_KEY

# Config file overrides defaults; flags still beat env/YAML where set
./bin/querywise analyze --config ./.querywise.yml

# Render a saved JSON report to the terminal or another format
./bin/querywise report --from report.json
./bin/querywise report --from report.json --output markdown --file pretty.md

./bin/querywise version
```

Environment variables (**`QUERYWISE_` prefix**) and `.querywise.yml` keys:

| YAML key | Env | Purpose |
|---------|-----|---------|
| `dsn` | `QUERYWISE_DSN` | PostgreSQL URI |
| `top` | `QUERYWISE_TOP` | Top‑N ranked statements |
| `min_calls` | `QUERYWISE_MIN_CALLS` | Minimum `calls` filter |
| `anthropic_api_key` | `ANTHROPIC_API_KEY` | Claude credentials |
| `anthropic_model` | `QUERYWISE_ANTHROPIC_MODEL` | Model id for recommendations |
| `hash_key` | `QUERYWISE_HASH_KEY` | Secret key for keyed fingerprints (see below) |
| `score_time_weight` | `QUERYWISE_SCORE_TIME_WEIGHT` | Weight for the execution‑time share (default `0.4`) |
| `score_io_weight` | `QUERYWISE_SCORE_IO_WEIGHT` | Weight for the I/O share (default `0.4`) |
| `score_freq_weight` | `QUERYWISE_SCORE_FREQ_WEIGHT` | Weight for the call‑frequency share (default `0.2`) |

Corresponding flags: `--hash-key`, `--score-time-weight`, `--score-io-weight`, `--score-freq-weight`.

**`hash_key`** — when set, query fingerprints use keyed **HMAC‑SHA256** instead of plain SHA‑256. This prevents trivial dictionary reversal of fingerprints when reports are shared: without knowing the key, an attacker cannot pre‑compute hashes of candidate SQL. **Without a key, fingerprints are plain SHA‑256** — obfuscation only, and reversible for anyone who can enumerate a known schema's likely statements.

**Effective precedence:** explicit flags → environment → YAML (if loaded) → built‑in defaults.

---

## Scoring

Each statement is assigned a single cost percentage:

```
COST% = (0.4·timeShare + 0.4·ioShare + 0.2·freqShare) × 100
```

Each *share* is the statement's fraction of the database‑wide total for the corresponding metric:

- `timeShare` — this statement's `total_exec_time` ÷ sum of all `total_exec_time`
- `ioShare` — this statement's `shared_blks_read` ÷ sum of all `shared_blks_read`
- `freqShare` — this statement's `calls` ÷ sum of all `calls`

The weights (`0.4 / 0.4 / 0.2`) are configurable via the config keys `score_time_weight`, `score_io_weight`, and `score_freq_weight` (env `QUERYWISE_SCORE_TIME_WEIGHT`, `QUERYWISE_SCORE_IO_WEIGHT`, `QUERYWISE_SCORE_FREQ_WEIGHT`; flags `--score-time-weight`, `--score-io-weight`, `--score-freq-weight`). Weights are normalized before scoring so that COST% always stays a percentage regardless of the raw values supplied.

---

## Sample output

```text
QueryWise — top 10 by estimated cost

RANK  HASH      COST%   CALLS    AVG TIME   MAX TIME   TOTAL TIME   CACHE HIT   RECOMMENDATION
  1   9f3a1c2b  31.2%   482,109    12.4 ms   210.7 ms    99.6 min      98.1%    Add index on orders(customer_id, created_at)
  2   4b7e08d1  18.9%    12,540    88.1 ms   940.3 ms    18.4 min      72.3%    Sequential scan; consider covering index
  3   c1d9f6a4  12.4%   901,332     1.8 ms    46.2 ms    27.0 min      99.9%    Hot path OK; watch for plan drift
  4   7a2b5e90   8.7%     3,201   145.6 ms   1.9 s        7.8 min      61.0%    Rewrite correlated subquery as JOIN
  5   e0c4d783   6.1%   210,884     2.9 ms    88.4 ms    10.2 min      97.4%    Acceptable
  6   1d8f22ab   4.3%    54,900     6.7 ms   132.0 ms     6.1 min      95.2%    Acceptable
  7   b9033c17   2.8%     8,412    27.3 ms   410.5 ms     3.8 min      84.6%    Consider partial index on status='pending'
  8   3fa71e6c   1.6%   140,220     1.1 ms    22.8 ms     2.6 min      99.5%    Acceptable
  9   88ce40d2   0.9%     2,004    41.2 ms   690.1 ms     1.4 min      70.8%    Infrequent but heavy; review off-peak scheduling
 10   0aa5b7f9   0.5%    77,650     0.6 ms    18.3 ms     0.8 min      99.8%    Acceptable

Total estimated DB cost covered by top 10: 87.4%
```

---

## Developer workflow

```bash
make build
make test
make lint     # golangci-lint
make release  # snapshot via goreleaser (requires tooling)
```

Build with embedded version tag:

```bash
go build -ldflags="-X querywise/cmd.Version=v0.1.0" -o bin/querywise .
```

---

## Known limitations

1. **Cumulative stats only.** All figures are lifetime‑cumulative since the last `pg_stat_statements_reset()` — there is no time window. A snapshot‑diff / real‑time mode is planned but not yet available.
2. **No true percentiles.** Latency is reported as mean plus stddev and max; `pg_stat_statements` exposes no p95/p99, so QueryWise cannot either.
3. **`--min-calls` can hide rare heavy queries.** The default of `10` filters out infrequent statements, which may include rare‑but‑expensive jobs. Lower it (e.g. `--min-calls 1`) to catch those.
4. **Bounded catalog capacity.** `pg_stat_statements` tracks at most `pg_stat_statements.max` entries and evicts the least‑used once full. Analysis only covers statements still resident in the view.

---

## Security posture

1. Postgres returns statement text → it is hashed the moment it is read; only the resulting hash plus numeric metrics are retained in the report or transmitted. Note this is not a memory‑scrubbing guarantee: Go strings are immutable, so clearing the local variable does not zero the underlying heap, and the pgx driver's buffers hold row text for their own lifetime regardless. The claim is about what QueryWise *keeps and sends*, not about wiping process memory.
2. Network calls (**`--recommend`**) transmit JSON containing **hash + metrics only**.
3. Use least‑privileged DB users and TLS DSN params in production URIs — but see [Required privileges](#required-privileges) for the one grant the connecting role does need.

---

## License

Project license is determined by the repository owner unless otherwise noted.
