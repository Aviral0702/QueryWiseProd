# QueryWise

Pure Go CLI binary that connects to PostgreSQL, reads **`pg_stat_statements`**, hashes query text locally with SHA‑256 (raw SQL is never stored in memory beyond the catalog read or sent to any API), ranks patterns with a configurable cost heuristic, optionally asks the Claude API for anonymized tuning hints, and prints or writes a report.

Requires the [**pg_stat_statements**](https://www.postgresql.org/docs/current/pgstatstatements.html) extension on the database you profile.

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

**Effective precedence:** explicit flags → environment → YAML (if loaded) → built‑in defaults.

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

## Security posture

1. Postgres returns statement text → hasher consumes it immediately; program state retains only hashes and numeric stats.
2. Network calls (**`--recommend`**) transmit JSON containing **hash + metrics only**.
3. Use least‑privileged DB users and TLS DSN params in production URIs.

---

## License

Project license is determined by the repository owner unless otherwise noted.
