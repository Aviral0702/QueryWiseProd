<div align="center">

# QueryWise

### Find your most expensive PostgreSQL queries — and get AI tuning advice — without your SQL ever leaving your infrastructure.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13%2B-336791?logo=postgresql&logoColor=white)
![Deploy](https://img.shields.io/badge/deploy-single%20static%20binary-success)
![Privacy](https://img.shields.io/badge/raw%20SQL-never%20leaves%20your%20box-1f6feb)
![License](https://img.shields.io/badge/license-TBD-lightgrey)

</div>

---

QueryWise is a single-binary CLI that reads PostgreSQL's `pg_stat_statements`, ranks the query patterns that actually cost you time and I/O, and — optionally — asks Claude for tuning advice. **Your raw SQL is hashed locally the moment it's read; only fingerprints and numeric metrics are ever kept or sent anywhere.**

No agent. No server. No signup. No data egress. Point it at a database, get a report in seconds.

```bash
QUERYWISE_DSN="postgres://user:pass@db:5432/app" querywise analyze
```

```text
QueryWise — top 10 by estimated cost

RANK  HASH      COST%   CALLS    AVG TIME   MAX TIME   TOTAL TIME   CACHE HIT   RECOMMENDATION
  1   9f3a1c2b  31.2%   482,109    12.4 ms   210.7 ms    99.6 min      98.1%    Add index on orders(customer_id, created_at)
  2   4b7e08d1  18.9%    12,540    88.1 ms   940.3 ms    18.4 min      72.3%    Sequential scan; consider covering index
  3   c1d9f6a4  12.4%   901,332     1.8 ms    46.2 ms    27.0 min      99.9%    Hot path OK; watch for plan drift
  4   7a2b5e90   8.7%     3,201   145.6 ms   1.9 s        7.8 min      61.0%    Rewrite correlated subquery as JOIN
  5   e0c4d783   6.1%   210,884     2.9 ms    88.4 ms    10.2 min      97.4%    Acceptable

Total estimated DB cost covered by top 10: 87.4%
```

---

## Why QueryWise

**🔒 Safe to point at production.** Most query profilers ship your SQL to a SaaS or need a heavyweight agent. QueryWise fingerprints every statement locally with SHA-256 (or keyed HMAC) — the actual query text is never stored in the report and never crosses the network. For fintech, health, and GDPR-bound teams, this is the profiler you're *allowed* to use.

**⚡ Zero-friction for engineers.** One static Go binary, no runtime, no dependencies. Install, run, read a ranked report. Export to terminal, Markdown, or JSON. Get AI-powered tuning hints with a single `--recommend` flag — and even those calls send only hashes and numbers, never your SQL.

**🧰 Built for infra & DBAs.** TLS is enforced by default (no silent plaintext fallback). It needs exactly one narrow grant (`pg_read_all_stats`), not superuser. Fingerprints can be keyed so reports are safe to share across a team without leaking which queries they describe.

---

## Quick start

**Build from source** (prebuilt binaries and a Homebrew tap are on the [roadmap](#roadmap)):

```bash
git clone https://github.com/<your-org>/querywise.git
cd querywise
make build          # produces ./bin/querywise
```

**Run it:**

```bash
# Colored, ranked terminal report
./bin/querywise analyze --dsn "postgres://user:pass@localhost:5432/mydb"

# Or via environment
export QUERYWISE_DSN="postgres://user:pass@localhost:5432/mydb"
./bin/querywise analyze --top 20

# AI tuning hints (SQL stays private — only hashes + metrics are sent)
export ANTHROPIC_API_KEY="sk-ant-..."
./bin/querywise analyze --recommend
```

First run against a fresh database? See [Prerequisites & setup](#prerequisites--setup) — you need the `pg_stat_statements` extension enabled once.

---

## How it works

```
PostgreSQL ──▶ read pg_stat_statements ──▶ hash SQL locally ──▶ rank by cost ──▶ report
                                             (SHA-256 / HMAC)      heuristic       terminal · markdown · json
                                                                                        │
                                                                        optional ──▶ Claude API
                                                                        (hashes + metrics only, no SQL)
```

1. **Read** — pulls each statement's calls, timing, block I/O, and cache stats from `pg_stat_statements`.
2. **Fingerprint** — the query text is hashed the instant it's read and discarded; only the hash and numbers move forward.
3. **Rank** — a transparent, configurable cost heuristic surfaces the patterns that dominate your database's time and I/O.
4. **Report** — a clean table locally, or Markdown/JSON to share; optional anonymized LLM recommendations.

---

## How it compares

| | QueryWise | pganalyze | pgHero | pgBadger | Datadog DBM |
|---|:---:|:---:|:---:|:---:|:---:|
| Raw SQL stays on your box | ✅ | ❌ | ✅ | ✅ | ❌ |
| Single binary, no agent/server | ✅ | ❌ | ❌ | ✅ | ❌ |
| AI tuning hints | ✅ | ✅ | ❌ | ❌ | partial |
| Free / open | ✅ | ❌ | ✅ | ✅ | ❌ |
| Live `pg_stat_statements` (no log shipping) | ✅ | ✅ | ✅ | ❌ | ✅ |
| Deep historical dashboards | roadmap | ✅ | partial | ❌ | ✅ |

QueryWise wins on **portability, privacy, and simplicity** — 5-minute triage where nothing leaves your machine. It is not (yet) a full historical monitoring platform; see the [roadmap](#roadmap).

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

This is in deliberate tension with the "use a least-privileged DB user" guidance: the *one* elevated grant QueryWise needs is `pg_read_all_stats`, and nothing more. Grant that specifically rather than reaching for superuser.

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

# LLM recommendations (needs ANTHROPIC_API_KEY)
./bin/querywise analyze --dsn "postgres://..." --recommend

# Config file overrides defaults; flags still beat env/YAML where set
./bin/querywise analyze --config ./.querywise.yml

# Render a saved JSON report to the terminal or another format
./bin/querywise report --from report.json
./bin/querywise report --from report.json --output markdown --file pretty.md

./bin/querywise version
```

## Configuration

Environment variables (**`QUERYWISE_` prefix**) and `.querywise.yml` keys:

| YAML key | Env | Purpose |
|---------|-----|---------|
| `dsn` | `QUERYWISE_DSN` | PostgreSQL URI |
| `top` | `QUERYWISE_TOP` | Top-N ranked statements |
| `min_calls` | `QUERYWISE_MIN_CALLS` | Minimum `calls` filter |
| `anthropic_api_key` | `ANTHROPIC_API_KEY` | Claude credentials |
| `anthropic_model` | `QUERYWISE_ANTHROPIC_MODEL` | Model id for recommendations |
| `hash_key` | `QUERYWISE_HASH_KEY` | Secret key for keyed fingerprints (see below) |
| `score_time_weight` | `QUERYWISE_SCORE_TIME_WEIGHT` | Weight for the execution-time share (default `0.4`) |
| `score_io_weight` | `QUERYWISE_SCORE_IO_WEIGHT` | Weight for the I/O share (default `0.4`) |
| `score_freq_weight` | `QUERYWISE_SCORE_FREQ_WEIGHT` | Weight for the call-frequency share (default `0.2`) |

Corresponding flags: `--hash-key`, `--score-time-weight`, `--score-io-weight`, `--score-freq-weight`.

**`hash_key`** — when set, query fingerprints use keyed **HMAC-SHA256** instead of plain SHA-256. This prevents trivial dictionary reversal of fingerprints when reports are shared: without knowing the key, an attacker cannot pre-compute hashes of candidate SQL. **Without a key, fingerprints are plain SHA-256** — obfuscation only, and reversible for anyone who can enumerate a known schema's likely statements.

**Effective precedence:** explicit flags → environment → YAML (if loaded) → built-in defaults.

---

## Scoring

Each statement is assigned a single cost percentage:

```
COST% = (0.4·timeShare + 0.4·ioShare + 0.2·freqShare) × 100
```

Each *share* is the statement's fraction of the database-wide total for the corresponding metric:

- `timeShare` — this statement's `total_exec_time` ÷ sum of all `total_exec_time`
- `ioShare` — this statement's `shared_blks_read` ÷ sum of all `shared_blks_read`
- `freqShare` — this statement's `calls` ÷ sum of all `calls`

The weights (`0.4 / 0.4 / 0.2`) are configurable via the config keys `score_time_weight`, `score_io_weight`, and `score_freq_weight` (env `QUERYWISE_SCORE_TIME_WEIGHT`, `QUERYWISE_SCORE_IO_WEIGHT`, `QUERYWISE_SCORE_FREQ_WEIGHT`; flags `--score-time-weight`, `--score-io-weight`, `--score-freq-weight`). Weights are normalized before scoring so that COST% always stays a percentage regardless of the raw values supplied.

---

## Security posture

1. **Raw SQL is never kept or sent.** Postgres returns statement text → it is hashed the moment it's read; only the resulting hash plus numeric metrics are retained in the report or transmitted. *Note:* this is about what QueryWise **keeps and sends**, not a memory-scrubbing guarantee — Go strings are immutable, so clearing the local variable does not zero the underlying heap, and the pgx driver's buffers hold row text for their own lifetime.
2. **Encrypted by default.** When `sslmode` is unset, QueryWise forces `sslmode=require` so connections never silently fall back to plaintext; it warns if you explicitly select a weaker mode. Use `sslmode=verify-full` with `sslrootcert` in production to also verify server identity.
3. **AI calls carry hashes + metrics only.** `--recommend` transmits JSON containing fingerprints and numbers — never your SQL.
4. **Shareable reports.** Set `hash_key` to fingerprint with keyed HMAC-SHA256 so reports can be shared without leaking which queries they describe.

---

## Known limitations

1. **Cumulative stats only.** All figures are lifetime-cumulative since the last `pg_stat_statements_reset()` — there is no time window yet. A snapshot-diff / real-time mode is planned (see [roadmap](#roadmap)).
2. **No true percentiles.** Latency is reported as mean plus stddev and max; `pg_stat_statements` exposes no p95/p99, so QueryWise cannot either.
3. **`--min-calls` can hide rare heavy queries.** The default of `10` filters out infrequent statements. Lower it (e.g. `--min-calls 1`) to catch rare-but-expensive jobs.
4. **Bounded catalog capacity.** `pg_stat_statements` tracks at most `pg_stat_statements.max` entries and evicts the least-used once full. Analysis only covers statements still resident in the view.

---

## Roadmap

- ⏱️ **Live watch mode** — snapshot-diff to show cost over a real time window, not just lifetime totals.
- 📦 **Prebuilt binaries + Homebrew tap** for one-line install.
- 📈 **Historical trends & team dashboard** (planned commercial layer; the CLI stays free and open).

---

## Developer workflow

```bash
make build
make test
make lint     # golangci-lint
make release  # snapshot via goreleaser (requires tooling)
```

Build with an embedded version tag:

```bash
go build -ldflags="-X querywise/cmd.Version=v0.1.0" -o bin/querywise .
```

Contributions welcome — open an issue to discuss a change before a large PR.

---

## License

License is to be finalized by the repository owner (MIT/Apache-2.0 recommended for adoption).
