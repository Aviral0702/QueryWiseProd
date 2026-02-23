# Testing QueryWise with a real database

Use the QueryWise agent to collect metrics from PostgreSQL and view them in the dashboard.

## 1. Get your instance API key

- Open the app at **http://localhost:3001**.
- Create an instance (e.g. name `my-db`) if you haven’t, or use an existing one.
- Copy the **API key** (from the green card after creating, or **Instance settings → Reveal API key**).

## 2. Prepare PostgreSQL

The database must have the `pg_stat_statements` extension enabled. With Docker Postgres from this repo:

```bash
# Start stack if not already running
docker compose up -d db backend frontend

# Enable extension and create reader user (run once)
docker compose exec db psql -U querywise -d querywise -f - < scripts/init-db-for-agent.sql
```

If the init script fails (e.g. permission), run the SQL manually:

```bash
docker compose exec db psql -U querywise -d querywise -c "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"
```

For local Postgres (non-Docker), run the same `CREATE EXTENSION` and grants as in `scripts/init-db-for-agent.sql` with a superuser.

## 3. Generate some load (optional)

So the agent has something to collect, run a few queries:

```bash
docker compose exec db psql -U querywise -d querywise -c "SELECT 1; SELECT count(*) FROM pg_stat_statements;"
```

Or use your app against this database.

## 4. Configure the agent

Copy the example config and set **database** and **backend** (use the API key from step 1):

```bash
cd agent
cp ../config/agent.example.yml config/agent.yml
```

Edit `config/agent.yml`:

**For Docker Postgres (backend and agent on host):**

```yaml
database:
  host: localhost
  port: 5432
  user: querywise
  password: querywise
  dbname: querywise
  sslmode: disable

backend:
  url: http://localhost:3000
  api_key: YOUR_API_KEY_FROM_STEP_1
```

**If using the reader user:**

```yaml
database:
  host: localhost
  port: 5432
  user: querywise_reader
  password: querywise_reader
  dbname: querywise
  sslmode: disable

backend:
  url: http://localhost:3000
  api_key: YOUR_API_KEY_FROM_STEP_1
```

Leave `collection.interval_seconds` (e.g. 60) and other sections as in the example.

## 5. Build and run the agent

From the repo root:

```bash
cd agent
go build -o querywise-agent ./cmd/main.go
./querywise-agent -config config/agent.yml
```

You should see logs like:

- `QueryWise Agent starting...`
- After the first run (about 2 seconds): `Ingested N queries for db_id=...`

## 6. View data in the dashboard

- Open **http://localhost:3001**.
- Select your instance in the **Database instance** dropdown.
- Click **Refresh** or wait for the next agent run.
- You should see **Total estimated cost**, **Avg response time**, and the **Top expensive queries** table (and trend chart if there’s enough data).

## Troubleshooting

| Issue | What to do |
|-------|------------|
| Agent: "Collect error: permission denied for pg_stat_statements" | Run the init SQL (step 2) so the DB user has `SELECT` on `pg_stat_statements`. |
| Agent: "No query metrics collected" | Ensure `pg_stat_statements` is enabled and the DB has had some activity. |
| Agent: "Ingest error: status 401" | Use the exact API key from the dashboard for this instance (Reveal API key). |
| Agent: "connection refused" to backend | Backend must be running; use `http://localhost:3000` for backend URL when the agent runs on your host. |
| Dashboard shows "No metrics yet" | Wait for at least one agent run (e.g. 60 s), then click Refresh. |
